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
	userProto "sortedstartup.com/stream/userservice/proto"
	"sortedstartup.com/stream/videoservice/config"
	"sortedstartup.com/stream/videoservice/db"
	"sortedstartup.com/stream/videoservice/proto"
)

type VideoAPI struct {
	config        config.VideoServiceConfig
	HTTPServerMux *http.ServeMux
	db            *sql.DB

	log       *slog.Logger
	dbQueries  db.DBQuerier 

	// gRPC clients for other services
	userServiceClient userProto.UserServiceClient

	//implemented proto server
	proto.UnimplementedVideoServiceServer
	TenantCheckFunc func(ctx context.Context, tenantID, userID string) error
}

func NewVideoAPIProduction(config config.VideoServiceConfig, userServiceClient userProto.UserServiceClient) (*VideoAPI, error) {
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

	ServerMux := http.NewServeMux()

	videoAPI := &VideoAPI{
		HTTPServerMux:     ServerMux,
		config:            config,
		db:                _db,
		log:               childLogger,
		dbQueries:         dbQueries,
		userServiceClient: userServiceClient,
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

// isUserInTenant checks if the user is part of the specified tenant
// by calling the userservice to get user's tenants and checking if the tenant is in the list
func (s *VideoAPI) isUserInTenant(ctx context.Context, tenantID, userID string) error {
	if s.TenantCheckFunc != nil {
        return s.TenantCheckFunc(ctx, tenantID, userID)
    }

	// Default tenant check implementation
	if tenantID == "" {
		return status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	// Call userservice to get user's tenants
	resp, err := s.userServiceClient.GetTenants(ctx, &userProto.GetTenantsRequest{})
	if err != nil {
		s.log.Error("Failed to get user tenants from userservice", "error", err, "userID", userID)
		return status.Error(codes.Internal, "failed to check tenant access")
	}

	// Check if the requested tenant is in the user's tenant list
	for _, tenantUser := range resp.TenantUsers {
		if tenantUser.Tenant.Id == tenantID {
			// User is a member of this tenant
			return nil
		}
	}

	// User is not a member of this tenant
	return status.Error(codes.PermissionDenied, "access denied: you are not a member of this tenant")
}

func (s *VideoAPI) ListVideos(ctx context.Context, req *proto.ListVideosRequest) (*proto.ListVideosResponse, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		s.log.Error("Error getting auth from context", "err", err)
		return nil, err
	}

	// Get tenant ID from headers/metadata
	tenantID, err := interceptors.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	// Validate user has access to this tenant
	err = s.isUserInTenant(ctx, tenantID, authContext.User.ID)
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
	tenantID, err := interceptors.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Validate user has access to this tenant
	err = s.isUserInTenant(ctx, tenantID, authContext.User.ID)
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
		Visibility:  proto.Visibility_VISIBILITY_PRIVATE, // All videos are private for now
		CreatedAt:   timestamppb.New(video.CreatedAt),
	}, nil
}
