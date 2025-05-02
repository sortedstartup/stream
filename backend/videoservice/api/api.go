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
	"sortedstartup.com/stream/videoservice/config"
	"sortedstartup.com/stream/videoservice/db"
	"sortedstartup.com/stream/videoservice/proto"
)

type VideoAPI struct {
	config        config.VideoServiceConfig
	HTTPServerMux *http.ServeMux
	db            *sql.DB

	log       *slog.Logger
	DBQueries db.QueriesInterface
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

	ServerMux := http.NewServeMux()

	videoAPI := &VideoAPI{
		HTTPServerMux: ServerMux,
		config:        config,
		db:            _db,
		log:           childLogger,
		DBQueries:     dbQueries,
	}

	ServerMux.Handle("/upload", interceptors.FirebaseHTTPAuthMiddleware(fbAuth, http.HandlerFunc(videoAPI.uploadHandler)))
	//TODO: implement auth middleware
	ServerMux.Handle("/video/", interceptors.FirebaseHTTPAuthMiddleware(fbAuth, http.HandlerFunc(videoAPI.serveVideoHandler)))

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

func (s *VideoAPI) ListVideos(ctx context.Context, req *proto.ListVideosRequest) (*proto.ListVideosResponse, error) {

	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		slog.Error("Error getting auth from context", "err", err)
		return nil, err
	}
	userID := authContext.User.ID
	pageSize := req.PageSize
	pageNumber := req.PageNumber

	if pageSize == 0 {
		pageSize = 10
	}

	videos, err := s.DBQueries.GetAllVideoUploadedByUserPaginated(ctx, db.GetAllVideoUploadedByUserPaginatedParams{
		UserID:     userID,
		PageSize:   int64(pageSize),
		PageNumber: int64(pageNumber),
	})
	if err != nil {
		slog.Error("Error getting videos", "err", err)
		return nil, err
	}

	protoVideos := make([]*proto.Video, 0, len(videos))

	for _, video := range videos {
		protoVideos = append(protoVideos, &proto.Video{
			Id:          video.ID,
			Title:       video.Title,
			Description: video.Description,
			Url:         video.Url,
			CreatedAt:   timestamppb.New(video.CreatedAt),
		})
	}

	return &proto.ListVideosResponse{Videos: protoVideos}, nil
}

func (s *VideoAPI) GetVideo(ctx context.Context, req *proto.GetVideoRequest) (*proto.Video, error) {
	// Get auth context to verify user has access
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		s.log.Error("Error getting auth from context", "err", err)
		return nil, err
	}

	// Get video from database
	video, err := s.DBQueries.GetVideoByID(ctx, db.GetVideoByIDParams{
		ID:     req.VideoId,
		UserID: authContext.User.ID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "video not found")
		}
		s.log.Error("Error getting video", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	// Verify user has access to this video
	if video.UploadedUserID != authContext.User.ID {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	// Convert to proto message
	return &proto.Video{
		Id:          video.ID,
		Title:       video.Title,
		Description: video.Description,
		Url:         video.Url,
		CreatedAt:   timestamppb.New(video.CreatedAt),
	}, nil
}
