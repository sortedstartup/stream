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
	dbQueries *db.Queries

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
		dbQueries:     dbQueries,
	}

	ServerMux.Handle("/upload", interceptors.CookieAuthMiddleware(fbAuth, http.HandlerFunc(videoAPI.uploadHandler)))
	ServerMux.Handle("/video/", interceptors.CookieAuthMiddleware(fbAuth, http.HandlerFunc(videoAPI.serveVideoHandler)))

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

	_, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		slog.Error("Error getting auth from context", "err", err)
		return nil, err
	}
	// Note: We still verify authentication, but now show videos from all users

	// Get all videos for all users (since authentication is already verified)
	videos, err := s.dbQueries.GetAllVideosForAllUsers(ctx)
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
	// Get auth context to verify user is authenticated
	_, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		s.log.Error("Error getting auth from context", "err", err)
		return nil, err
	}

	// Get video from database (now allows any authenticated user to access any video)
	video, err := s.dbQueries.GetVideoByIDForAllUsers(ctx, req.VideoId)
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
		CreatedAt:   timestamppb.New(video.CreatedAt),
	}, nil
}
