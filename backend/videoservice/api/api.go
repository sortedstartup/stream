package api

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	_ "modernc.org/sqlite"
	"sortedstartup.com/stream/common/auth"
	"sortedstartup.com/stream/common/interceptors"
	userDB "sortedstartup.com/stream/userservice/db"
	"sortedstartup.com/stream/videoservice/config"
	"sortedstartup.com/stream/videoservice/db"
	"sortedstartup.com/stream/videoservice/proto"
)

type VideoAPI struct {
	config        config.VideoServiceConfig
	HTTPServerMux *http.ServeMux
	db            *sql.DB

	log           *slog.Logger
	dbQueries     *db.Queries
	userDBQueries *userDB.Queries

	//implemented proto server
	proto.UnimplementedVideoServiceServer
}

func NewVideoAPIProduction(config config.VideoServiceConfig) (*VideoAPI, error) {
	slog.Info("NewVideoAPIProduction")

	fbAuth, err := auth.NewFirebase()
	if err != nil {
		return nil, err
	}

	childLogger := slog.With("service", "VideoAPI")

	_db, err := sql.Open(config.DB.Driver, config.DB.Url)
	if err != nil {
		return nil, err
	}

	dbQueries := db.New(_db)

	// Also open connection to userservice database for tenant validation
	userDBConnection, err := sql.Open("sqlite", "backend/userservice/db.sqlite")
	if err != nil {
		return nil, err
	}
	userDBQueries := userDB.New(userDBConnection)

	ServerMux := http.NewServeMux()

	videoAPI := &VideoAPI{
		HTTPServerMux: ServerMux,
		config:        config,
		db:            _db,
		log:           childLogger,
		dbQueries:     dbQueries,
		userDBQueries: userDBQueries,
	}

	// The authentication is handled in mono/main.go
	ServerMux.Handle("/upload", interceptors.FirebaseHTTPHeaderAuthMiddleware(fbAuth, http.HandlerFunc(videoAPI.uploadHandler)))
	//the cookie auth middleware is just to allow if the user is logged in
	ServerMux.Handle("/video/", interceptors.FirebaseCookieAuthMiddleware(fbAuth, http.HandlerFunc(videoAPI.serveVideoHandler)))

	return videoAPI, nil
}

func (s *VideoAPI) Start() error {
	return nil
}

func (s *VideoAPI) Init() error {
	s.log.Info("Migrating database", "dbDriver", s.config.DB.Driver, "dbURL", s.config.DB.Url)
	err := db.MigrateDB(s.config.DB.Driver, s.config.DB.Url)
	if err != nil {
		return err
	}
	s.log.Info("Migrating database done")
	return nil
}

// validateTenantAccess checks if the user has access to the specified tenant
func (s *VideoAPI) validateTenantAccess(ctx context.Context, tenantID, userID string) error {
	if tenantID == "" {
		return status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	// Check if user has access to this tenant
	_, err := s.userDBQueries.GetUserRoleInTenant(ctx, userDB.GetUserRoleInTenantParams{
		TenantID: tenantID,
		UserID:   userID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return status.Error(codes.PermissionDenied, "access denied: you are not a member of this tenant")
		}
		s.log.Error("Failed to check user role in tenant", "error", err)
		return status.Error(codes.Internal, "failed to check tenant access")
	}

	return nil
}

// getTenantIDFromMetadata extracts tenant ID from gRPC metadata
func (s *VideoAPI) getTenantIDFromMetadata(ctx context.Context) (string, error) {
	// For gRPC, we need to extract the X-Tenant-ID header from metadata
	// The grpc-web client should pass this header, and it gets converted to metadata

	// Import google.golang.org/grpc/metadata if not already imported
	// For now, we'll implement a simple context value extraction
	// This assumes the interceptor has already extracted the header and put it in context

	if tenantID, ok := ctx.Value("X-Tenant-ID").(string); ok && tenantID != "" {
		return tenantID, nil
	}

	return "", status.Error(codes.InvalidArgument, "tenant ID header (X-Tenant-ID) is required")
}

func (s *VideoAPI) ListVideos(ctx context.Context, req *proto.ListVideosRequest) (*proto.ListVideosResponse, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		s.log.Error("Error getting auth from context", "err", err)
		return nil, err
	}

	// Get tenant ID from headers/metadata
	tenantID, err := s.getTenantIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	// Validate user has access to this tenant
	err = s.validateTenantAccess(ctx, tenantID, authContext.User.ID)
	if err != nil {
		return nil, err
	}

	// Get videos for this tenant
	videos, err := s.dbQueries.GetVideosByTenantID(ctx, sql.NullString{
		String: tenantID,
		Valid:  true,
	})
	if err != nil {
		s.log.Error("Error getting videos for tenant", "err", err, "tenantID", tenantID)
		return nil, status.Error(codes.Internal, "failed to get videos")
	}

	protoVideos := make([]*proto.Video, 0, len(videos))

	for _, video := range videos {
		protoVideos = append(protoVideos, &proto.Video{
			Id:          video.ID,
			Title:       video.Title,
			Description: video.Description,
			Url:         video.Url,
			TenantId:    video.TenantID.String,
			Visibility:  proto.Visibility_VISIBILITY_PRIVATE, // All videos are private for now
			CreatedAt:   timestamppb.New(video.CreatedAt),
		})
	}

	return &proto.ListVideosResponse{Videos: protoVideos}, nil
}

func (s *VideoAPI) GetVideo(ctx context.Context, req *proto.GetVideoRequest) (*proto.Video, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		s.log.Error("Error getting auth from context", "err", err)
		return nil, err
	}

	// Get tenant ID from headers/metadata
	tenantID, err := s.getTenantIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	// Validate user has access to this tenant
	err = s.validateTenantAccess(ctx, tenantID, authContext.User.ID)
	if err != nil {
		return nil, err
	}

	// Get video from database with tenant validation
	video, err := s.dbQueries.GetVideoByVideoIDAndTenantID(ctx, db.GetVideoByVideoIDAndTenantIDParams{
		ID: req.VideoId,
		TenantID: sql.NullString{
			String: tenantID,
			Valid:  true,
		},
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "video not found")
		}
		s.log.Error("Error getting video", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	// Convert to proto message
	return &proto.Video{
		Id:          video.ID,
		Title:       video.Title,
		Description: video.Description,
		Url:         video.Url,
		TenantId:    video.TenantID.String,
		Visibility:  proto.Visibility_VISIBILITY_PRIVATE, // All videos are private for now
		CreatedAt:   timestamppb.New(video.CreatedAt),
	}, nil
}
