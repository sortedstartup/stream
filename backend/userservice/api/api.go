package api

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	userCache *lru.Cache
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

	cache, err := lru.New(config.CacheSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create user cache: %w", err)
	}

	userAPI := &UserAPI{
		config:    config,
		db:        _db,
		log:       childLogger,
		userCache: cache,
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

	return nil
}

func (s *UserAPI) CreateUserIfNotExists(ctx context.Context, req *proto.GetUserByEmailRequest) (*proto.GetUserByEmailResponse, error) {
	s.log.Info("CreateUserIfNotExists")
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userEmail := authContext.User.Email
	s.log.Info("userEmail", "userEmail", userEmail)

	// CACHE CHECK
	if _, found := s.userCache.Get(userEmail); found {
		s.log.Info("Cache hit: skipping DB check", "email", userEmail)
		return &proto.GetUserByEmailResponse{
			Message: "User already exists (cache)",
			Success: true,
		}, nil
	}

	s.log.Info("DB_CHECK: querying DB for email", "email", userEmail)

	// DB CHECK (and fallback to create)
	dbUser, err := s.dbQueries.GetUserByEmail(ctx, userEmail)
	s.log.Info("GetUserByEmail result", "error", err, "hasError", err != nil, "isNoRows", err == sql.ErrNoRows)

	var successMessage string

	if err != nil {
		if err == sql.ErrNoRows {
			s.log.Info("User doesn't exist, creating them")
			// User doesn't exist, create them
			createParams := db.CreateUserParams{
				ID:        authContext.User.ID,
				Username:  authContext.User.Email,
				Email:     authContext.User.Email,
				CreatedAt: time.Now(),
			}

			dbUser, err = s.dbQueries.CreateUser(ctx, createParams)
			if err != nil {
				s.log.Error("Failed to create user", "error", err)
				return nil, status.Error(codes.Internal, "failed to create user")
			}
			s.log.Info("User created successfully with email", "email", authContext.User.Email)
			successMessage = "User created successfully"
			s.userCache.Add(userEmail, true)
			s.log.Info("CACHE_ADD: adding email to cache", "email", userEmail)
		} else {
			s.log.Error("Database error while getting user", "error", err)
			return nil, status.Error(codes.Internal, "internal server error")
		}
	} else {
		s.log.Info("User already exists with email", "email", authContext.User.Email)
		// Check if the returned user is actually valid (not empty)
		if dbUser.ID == "" || dbUser.Email == "" {
			s.log.Warn("User exists but has empty fields, this might indicate a database issue", "dbUser", dbUser)
		}
		successMessage = "User already exists"
	}

	// ADD TO CACHE if user haven't logged in recently (Cache miss) to cache hit in future login
	s.userCache.Add(userEmail, true)
	s.log.Info("CACHE_ADD: adding email to cache", "email", userEmail)

	// Return success response with message
	return &proto.GetUserByEmailResponse{
		Message: successMessage,
		Success: true,
	}, nil
}
