package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"sortedstartup.com/stream/common/auth"
	"sortedstartup.com/stream/common/interceptors"
	"sortedstartup.com/stream/mono/config"

	videoAPI "sortedstartup.com/stream/videoservice/api"
	videoProto "sortedstartup.com/stream/videoservice/proto"

	commentAPI "sortedstartup.com/stream/commentservice/api"
	commentProto "sortedstartup.com/stream/commentservice/proto"
)

//go:embed webapp/dist
var staticFiles embed.FS

var (
	Version   string
	BuildTime string
)

type Monolith struct {
	Config   *config.MonolithConfig
	Firebase *auth.Firebase

	VideoAPI   *videoAPI.VideoAPI
	CommentAPI *commentAPI.CommentAPI

	GRPCServer    *grpc.Server
	GRPCWebServer *http.Server

	log *slog.Logger
}

func main() {

	startTime := time.Now()
	monolith, err := NewMonolith()
	if err != nil {
		slog.Error("Could not create monolith", "err", err)
		panic(fmt.Errorf("could not create monolith: %w", err))
	}

	setupLogging(monolith.Config.LogLevel)

	err = monolith.InitServices()
	if err != nil {
		slog.Error("Could not migrate databases", "err", err)
		panic(fmt.Errorf("could not migrate databases: %w", err))
	}

	err = monolith.StartServices()
	if err != nil {
		slog.Error("Could not migrate databases", "err", err)
		panic(fmt.Errorf("could not migrate databases: %w", err))
	}

	endTime := time.Now()
	slog.Info("StartupTime", "time", endTime.Sub(startTime))

	err = monolith.startServer()
	if err != nil {
		slog.Error("Server was abruptly stopped", "err", err)
		panic(fmt.Errorf("server was abruptly stopped: %w", err))
	}
}

func NewMonolith() (*Monolith, error) {

	log, err := setupLogging("INFO")
	if err != nil {
		return nil, err
	}

	log.Info("Reading monolith configuration")
	config, err := config.New()
	if err != nil {
		log.Error("Could not read config file", "err", err)
		return nil, err
	}

	//TODO: config may have secrets, it may not be a good idea to print it
	//or we could have secret config, or we could have a way to mask secrets
	log.Info("Using monolith configuration", "config", config)

	log.Info("Creating monolith components")

	log.Info("Creating firebase")
	firebase, err := auth.NewFirebase()
	if err != nil {
		return nil, err
	}

	log.Info("Creating videoservice API")
	videoAPI, err := videoAPI.NewVideoAPIProduction(config.VideoService)
	if err != nil {
		log.Error("Could not create videoservice API", "err", err)
		return nil, err
	}

	log.Info("Creating commentservice API")
	commentAPI, err := commentAPI.NewCommentAPIProduction(config.CommentService)
	if err != nil {
		log.Error("Could not create commentservice API", "err", err)
		return nil, err
	}

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors.PanicRecoveryInterceptor(), interceptors.FirebaseAuthInterceptor(firebase)))

	// GRPC Web is a http server 1.0 server that wraps a grpc server
	// Browsers JS clients can only talk to GRPC web for now
	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	// Create a sub-filesystem for the dist directory
	distFS, err := fs.Sub(staticFiles, "webapp/dist")
	if err != nil {
		log.Error("Could not create sub-filesystem for static files", "err", err)
		return nil, err
	}

	// Create a file server for the static files
	staticFileServer := http.FileServer(http.FS(distFS))

	parentMux := http.NewServeMux()
	parentMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(r) || wrappedGrpc.IsAcceptableGrpcCorsRequest(r) {
			wrappedGrpc.ServeHTTP(w, r)
			return
		}
		// Serve static files for non-gRPC requests
		staticFileServer.ServeHTTP(w, r)
	})

	// parentMux.Handle("/test/*", aNewMux())
	// this muxOne can be got from video service struct
	parentMux.Handle("/api/videoservice/", http.StripPrefix("/api/videoservice", videoAPI.HTTPServerMux))
	parentMux.Handle("/api/commentservice/", http.StripPrefix("/api/commentservice", commentAPI.HTTPServerMux))

	httpServer := &http.Server{
		Addr:    config.Server.GrpcWebAddrPortString(),
		Handler: enableCORS(parentMux),
	}

	return &Monolith{
		Config:        &config,
		VideoAPI:      videoAPI,
		CommentAPI:    commentAPI,
		Firebase:      firebase,
		GRPCServer:    grpcServer,
		GRPCWebServer: httpServer,
		log:           log,
	}, nil
}

func (m *Monolith) InitServices() error {

	m.log.Info("Initializing Task Service")
	err := m.VideoAPI.Init()
	if err != nil {
		return err
	}

	m.log.Info("Initializing Comment Service")
	err = m.CommentAPI.Init()
	if err != nil {
		return err
	}

	return nil
}

func (m *Monolith) StartServices() error {

	m.log.Info("Starting Task Service")
	err := m.VideoAPI.Start()
	if err != nil {
		return err
	}

	m.log.Info("Starting Comment Service")
	err = m.CommentAPI.Start()
	if err != nil {
		return err
	}

	return nil

}

func (m *Monolith) startServer() error {

	listener, err := net.Listen("tcp", m.Config.Server.GRPCAddrPortString())
	if err != nil {
		panic(err)
	}

	videoProto.RegisterVideoServiceServer(m.GRPCServer, m.VideoAPI)
	commentProto.RegisterCommentServiceServer(m.GRPCServer, m.CommentAPI)

	reflection.Register(m.GRPCServer)

	serverErr := make(chan error)

	go func() {
		m.log.Info("Starting gRPC server", "addr", m.Config.Server.GRPCAddrPortString())
		err = m.GRPCServer.Serve(listener)
		if err != nil {
			serverErr <- err
		}
	}()

	go func() {
		m.log.Info("Starting gRPC web server", "addr", m.Config.Server.GrpcWebAddrPortString())
		err = m.GRPCWebServer.ListenAndServe()
		if err != nil {
			serverErr <- err
		}
	}()

	err = <-serverErr

	return err
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set necessary headers for CORS
		w.Header().Set("Access-Control-Allow-Origin", "*") // Adjust in production
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-User-Agent, X-Grpc-Web, family-id")

		// Check for preflight request
		if r.Method == "OPTIONS" {
			return
		}

		// Serve the request
		next.ServeHTTP(w, r)
	})
}

func setupLogging(loglevel string) (*slog.Logger, error) {

	slog.Info("Setting up logging")
	logger := slog.New(slog.Default().Handler())

	switch loglevel {
	case "DEBUG":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "INFO":
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case "WARN":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "ERROR":
		slog.SetLogLoggerLevel(slog.LevelError)
	default:
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	slog.SetDefault(logger)
	return logger, nil
}
