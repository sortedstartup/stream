package api

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	_ "modernc.org/sqlite"
	"sortedstartup.com/stream/userservice/config"
	"sortedstartup.com/stream/userservice/db"
	"sortedstartup.com/stream/userservice/proto"
)

type UserAPI struct {
	config    config.UserServiceConfig
	db        *sql.DB
	log       *slog.Logger
	dbQueries *db.Queries

	//implemented proto server
	proto.UnimplementedUserServiceServer
}

// NewUserAPITest creates a UserAPI instance with a mock database for testing.
func NewUserAPITest(mockDB *db.Queries, logger *slog.Logger) *UserAPI {
	return &UserAPI{
		log:       logger,
		dbQueries: mockDB,
	}
}

func NewUserAPIProduction(config config.UserServiceConfig) (*UserAPI, error) {
	slog.Info("NewUserAPIProduction")

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

func (s *UserAPI) Init() error {
	s.log.Info("Migrating database", "dbDriver", s.config.DB.Driver, "dbURL", s.config.DB.Url)
	err := db.MigrateDB(s.config.DB.Driver, s.config.DB.Url)
	if err != nil {
		return err
	}
	s.log.Info("Migrating database done")
	return nil
}

func (s *UserAPI) GetUser(ctx context.Context, req *proto.GetUserRequest) (*proto.User, error) {
	user, err := s.dbQueries.GetUserByID(ctx, req.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		s.log.Error("Error getting user", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return &proto.User{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt.Time),
	}, nil
}

func (s *UserAPI) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.User, error) {
	userID := generateUUID()

	user, err := s.dbQueries.CreateUser(ctx, db.CreateUserParams{
		ID:       userID,
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil {
		s.log.Error("Error creating user", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &proto.User{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt.Time),
	}, nil
}

func (s *UserAPI) UpdateUser(ctx context.Context, req *proto.UpdateUserRequest) (*proto.User, error) {
	// Get the current user first to handle optional updates
	currentUser, err := s.dbQueries.GetUserByID(ctx, req.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		s.log.Error("Error getting current user", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	username := currentUser.Username
	email := currentUser.Email

	if req.Username != nil {
		username = *req.Username
	}
	if req.Email != nil {
		email = *req.Email
	}

	user, err := s.dbQueries.UpdateUser(ctx, db.UpdateUserParams{
		ID:       req.UserId,
		Username: username,
		Email:    email,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		s.log.Error("Error updating user", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &proto.User{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt.Time),
	}, nil
}

func (s *UserAPI) DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*proto.Empty, error) {
	err := s.dbQueries.DeleteUser(ctx, req.UserId)
	if err != nil {
		s.log.Error("Error deleting user", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &proto.Empty{}, nil
}

func (s *UserAPI) ValidateUser(ctx context.Context, req *proto.ValidateUserRequest) (*proto.ValidateUserResponse, error) {
	count, err := s.dbQueries.ValidateUser(ctx, req.UserId)
	if err != nil {
		s.log.Error("Error validating user", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to validate user: %v", err)
	}

	isValid := count > 0
	var user *proto.User

	if isValid {
		dbUser, err := s.dbQueries.GetUserByID(ctx, req.UserId)
		if err != nil {
			s.log.Error("Error getting user for validation", "err", err)
			return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
		}

		user = &proto.User{
			Id:        dbUser.ID,
			Username:  dbUser.Username,
			Email:     dbUser.Email,
			CreatedAt: timestamppb.New(dbUser.CreatedAt.Time),
		}
	}

	return &proto.ValidateUserResponse{
		IsValid: isValid,
		User:    user,
	}, nil
}

func (s *UserAPI) GetUserByEmail(ctx context.Context, req *proto.GetUserByEmailRequest) (*proto.User, error) {
	user, err := s.dbQueries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		s.log.Error("Error getting user by email", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return &proto.User{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt.Time),
	}, nil
}

func generateUUID() string {
	return uuid.New().String()
}
