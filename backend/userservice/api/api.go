package api

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sortedstartup.com/stream/common/interceptors"
	"sortedstartup.com/stream/userservice/config"
	"sortedstartup.com/stream/userservice/db"
	"sortedstartup.com/stream/userservice/proto"
)

type UserAPI struct {
	config    config.UserServiceConfig
	db        *sql.DB
	log       *slog.Logger
	dbQueries *db.Queries
	proto.UnimplementedUserServiceServer
}

func NewUserAPI(config config.UserServiceConfig) (*UserAPI, error) {
	slog.Info("NewUserAPI")

	childLogger := slog.With("service", "UserAPI")

	_db, err := sql.Open(config.DB.Driver, config.DB.Url)
	if err != nil {
		return nil, err
	}

	dbQueries := db.New(_db)

	userAPI := &UserAPI{
		config:    config,
		db:        _db,
		log:       childLogger,
		dbQueries: dbQueries,
	}
	return userAPI, nil
}

func (s *UserAPI) Start() error {
	return nil
}

func (s *UserAPI) init() error {
	s.log.Info("Migrating database", "dbDriver", s.config.DB.Driver, "dbURL", s.config.DB.Url)
	err := db.MigrateDB(s.config.DB.Driver, s.config.DB.Url)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserAPI) CreateUserIfNotExistsHandler(ctx context.Context, req *proto.GetUserByEmailRequest) (*proto.User, error) {
	s.log.Info("CreateUserIfNotExistsHandler")
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userEmail := authContext.User.Email

	dbUser, err := s.dbQueries.GetUserByEmail(ctx, userEmail)
	if err != nil {
		if err == sql.ErrNoRows {
			// User doesn't exist, create them
			createParams := db.CreateUserParams{
				ID:        authContext.User.ID,
				Username:  authContext.User.Email,
				Email:     authContext.User.Email,
				AvatarUrl: "https://example.com/avatar.png",
				CreatedAt: time.Now(),
			}

			dbUser, err = s.dbQueries.CreateUser(ctx, createParams)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to create user")
			}
		} else {
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	// Convert db.User to proto.User
	protoUser := &proto.User{
		Id:        dbUser.ID,
		Username:  dbUser.Username,
		Email:     dbUser.Email,
		AvatarUrl: dbUser.AvatarUrl,
		CreatedAt: timestamppb.New(dbUser.CreatedAt),
	}

	return protoUser, nil
}
