package api

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	_ "modernc.org/sqlite"
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
	childLogger := slog.With("service", "VideoAPI")

	_db, err := sql.Open(config.DB.Driver, config.DB.Url)
	if err != nil {
		return nil, err
	}

	ServerMux := http.NewServeMux()
	ServerMux.HandleFunc("/upload", uploadHandler)

	return &VideoAPI{
		HTTPServerMux: ServerMux,
		config:        config,
		db:            _db,
		log:           childLogger,
	}, nil
	// return &VideoAPI{}, nil
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

	videos, err := s.dbQueries.GetAllVideo(ctx)
	if err != nil {
		return nil, err
	}

	protoVideos := make([]*proto.Video, 0, len(videos))

	for _, video := range videos {
		protoVideos = append(protoVideos, &proto.Video{
			Id:          video.ID,
			Title:       video.Title,
			Description: video.Description,
			Url:         video.Url,
		})
	}

	return &proto.ListVideosResponse{Videos: protoVideos}, nil
}
