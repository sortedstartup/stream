package api

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
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

	videos, err := s.dbQueries.GetAllVideoUploadedByUserPaginated(ctx, db.GetAllVideoUploadedByUserPaginatedParams{
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
	video, err := s.dbQueries.GetVideoByID(ctx, db.GetVideoByIDParams{
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

// Space-related API methods

func (s *VideoAPI) CreateSpace(ctx context.Context, req *proto.CreateSpaceRequest) (*proto.Space, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		s.log.Error("Error getting auth from context", "err", err)
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	spaceID := uuid.New().String()
	now := time.Now()

	err = s.dbQueries.CreateSpace(ctx, db.CreateSpaceParams{
		ID:          spaceID,
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		UserID:      authContext.User.ID,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		s.log.Error("Error creating space", "err", err)
		return nil, status.Error(codes.Internal, "failed to create space")
	}

	return &proto.Space{
		Id:          spaceID,
		Name:        req.Name,
		Description: req.Description,
		UserId:      authContext.User.ID,
		CreatedAt:   timestamppb.New(now),
		UpdatedAt:   timestamppb.New(now),
	}, nil
}

func (s *VideoAPI) ListSpaces(ctx context.Context, req *proto.ListSpacesRequest) (*proto.ListSpacesResponse, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		s.log.Error("Error getting auth from context", "err", err)
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	spaces, err := s.dbQueries.GetSpacesByUser(ctx, authContext.User.ID)
	if err != nil {
		s.log.Error("Error getting spaces", "err", err)
		return nil, status.Error(codes.Internal, "failed to get spaces")
	}

	protoSpaces := make([]*proto.Space, 0, len(spaces))
	for _, space := range spaces {
		protoSpaces = append(protoSpaces, &proto.Space{
			Id:          space.ID,
			Name:        space.Name,
			Description: space.Description.String,
			UserId:      space.UserID,
			CreatedAt:   timestamppb.New(space.CreatedAt),
			UpdatedAt:   timestamppb.New(space.UpdatedAt),
		})
	}

	return &proto.ListSpacesResponse{Spaces: protoSpaces}, nil
}

func (s *VideoAPI) GetSpace(ctx context.Context, req *proto.GetSpaceRequest) (*proto.Space, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		s.log.Error("Error getting auth from context", "err", err)
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	space, err := s.dbQueries.GetSpaceByID(ctx, db.GetSpaceByIDParams{
		ID:     req.SpaceId,
		UserID: authContext.User.ID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "space not found")
		}
		s.log.Error("Error getting space", "err", err)
		return nil, status.Error(codes.Internal, "failed to get space")
	}

	return &proto.Space{
		Id:          space.ID,
		Name:        space.Name,
		Description: space.Description.String,
		UserId:      space.UserID,
		CreatedAt:   timestamppb.New(space.CreatedAt),
		UpdatedAt:   timestamppb.New(space.UpdatedAt),
	}, nil
}

func (s *VideoAPI) ListVideosInSpace(ctx context.Context, req *proto.ListVideosInSpaceRequest) (*proto.ListVideosResponse, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		s.log.Error("Error getting auth from context", "err", err)
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	// First verify user has access to the space
	_, err = s.dbQueries.GetSpaceByID(ctx, db.GetSpaceByIDParams{
		ID:     req.SpaceId,
		UserID: authContext.User.ID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "space not found")
		}
		s.log.Error("Error getting space", "err", err)
		return nil, status.Error(codes.Internal, "failed to verify space access")
	}

	videos, err := s.dbQueries.GetVideosInSpace(ctx, db.GetVideosInSpaceParams{
		SpaceID: req.SpaceId,
		UserID:  authContext.User.ID,
	})
	if err != nil {
		s.log.Error("Error getting videos in space", "err", err)
		return nil, status.Error(codes.Internal, "failed to get videos")
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

func (s *VideoAPI) AddVideoToSpace(ctx context.Context, req *proto.AddVideoToSpaceRequest) (*proto.Empty, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		s.log.Error("Error getting auth from context", "err", err)
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	// Verify user owns the video
	_, err = s.dbQueries.GetVideoByID(ctx, db.GetVideoByIDParams{
		ID:     req.VideoId,
		UserID: authContext.User.ID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "video not found")
		}
		s.log.Error("Error getting video", "err", err)
		return nil, status.Error(codes.Internal, "failed to verify video access")
	}

	// Verify user owns the space
	_, err = s.dbQueries.GetSpaceByID(ctx, db.GetSpaceByIDParams{
		ID:     req.SpaceId,
		UserID: authContext.User.ID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "space not found")
		}
		s.log.Error("Error getting space", "err", err)
		return nil, status.Error(codes.Internal, "failed to verify space access")
	}

	now := time.Now()
	err = s.dbQueries.AddVideoToSpace(ctx, db.AddVideoToSpaceParams{
		VideoID:   req.VideoId,
		SpaceID:   req.SpaceId,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		s.log.Error("Error adding video to space", "err", err)
		return nil, status.Error(codes.Internal, "failed to add video to space")
	}

	return &proto.Empty{}, nil
}

func (s *VideoAPI) RemoveVideoFromSpace(ctx context.Context, req *proto.RemoveVideoFromSpaceRequest) (*proto.Empty, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		s.log.Error("Error getting auth from context", "err", err)
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	// Verify user owns the video
	_, err = s.dbQueries.GetVideoByID(ctx, db.GetVideoByIDParams{
		ID:     req.VideoId,
		UserID: authContext.User.ID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "video not found")
		}
		s.log.Error("Error getting video", "err", err)
		return nil, status.Error(codes.Internal, "failed to verify video access")
	}

	// Verify user owns the space
	_, err = s.dbQueries.GetSpaceByID(ctx, db.GetSpaceByIDParams{
		ID:     req.SpaceId,
		UserID: authContext.User.ID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "space not found")
		}
		s.log.Error("Error getting space", "err", err)
		return nil, status.Error(codes.Internal, "failed to verify space access")
	}

	err = s.dbQueries.RemoveVideoFromSpace(ctx, db.RemoveVideoFromSpaceParams{
		VideoID: req.VideoId,
		SpaceID: req.SpaceId,
	})
	if err != nil {
		s.log.Error("Error removing video from space", "err", err)
		return nil, status.Error(codes.Internal, "failed to remove video from space")
	}

	return &proto.Empty{}, nil
}
