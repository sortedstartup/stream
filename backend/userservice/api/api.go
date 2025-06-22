package api

import (
	"database/sql"
	"log/slog"
	"net/http"

	"sortedstartup.com/stream/common/auth"
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

func (s *UserAPI) init() error {
	s.log.Info("Migrating database", "dbDriver", s.config.DB.Driver, "dbURL", s.config.DB.Url)
	err := db.MigrateDB(s.config.DB.Driver, s.config.DB.Url)
	if err != nil {
		return err
	}

	return nil
}

func NewUserAPI(config config.UserServiceConfig) (*UserAPI, error) {
	slog.Info("NewUserAPI")

	fbAuth, err := auth.NewFirebase()
	if err != nil {
		return nil, err
	}

	childLogger := slog.With("service", "UserAPI")

	_db, err := sql.Open(config.DB.Driver, config.DB.Url)
	if err != nil {
		return nil, err
	}

	dbQueries := db.New(_db)

	ServerMux := http.NewServeMux()

	userAPI := &UserAPI{
		config:    config,
		db:        _db,
		log:       childLogger,
		dbQueries: dbQueries,
	}

	ServerMux.Handle("/ensureUser", interceptors.FirebaseHTTPHeaderAuthMiddleware(fbAuth, http.HandlerFunc(userAPI.CreateUserIfNotExistsHandler)))

	return userAPI, nil
}
