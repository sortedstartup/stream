package main

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io"
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

	userAPI "sortedstartup.com/stream/userservice/api"
	userProto "sortedstartup.com/stream/userservice/proto"

	paymentAPI "sortedstartup.com/stream/paymentservice/api"
	paymentDB "sortedstartup.com/stream/paymentservice/db"
	paymentProto "sortedstartup.com/stream/paymentservice/proto"
)

//go:embed webapp/dist
var staticFiles embed.FS

var (
	Version   string
	BuildTime string
)

// UserServiceClientWrapper wraps the UserAPI to implement the UserServiceClient interface
// This avoids circular dependency issues in the monolith
type UserServiceClientWrapper struct {
	userAPI *userAPI.UserAPI
}

func (w *UserServiceClientWrapper) CreateUserIfNotExists(ctx context.Context, req *userProto.CreateUserRequest, opts ...grpc.CallOption) (*userProto.CreateUserResponse, error) {
	return w.userAPI.CreateUserIfNotExists(ctx, req)
}

func (w *UserServiceClientWrapper) GetTenants(ctx context.Context, req *userProto.GetTenantsRequest, opts ...grpc.CallOption) (*userProto.GetTenantsResponse, error) {
	return w.userAPI.GetTenants(ctx, req)
}

// TenantServiceClientWrapper wraps the TenantAPI to implement the TenantServiceClient interface
type TenantServiceClientWrapper struct {
	tenantAPI *userAPI.TenantAPI
}

func (w *TenantServiceClientWrapper) CreateTenant(ctx context.Context, req *userProto.CreateTenantRequest, opts ...grpc.CallOption) (*userProto.CreateTenantResponse, error) {
	return w.tenantAPI.CreateTenant(ctx, req)
}

func (w *TenantServiceClientWrapper) AddUser(ctx context.Context, req *userProto.AddUserRequest, opts ...grpc.CallOption) (*userProto.AddUserResponse, error) {
	return w.tenantAPI.AddUser(ctx, req)
}

func (w *TenantServiceClientWrapper) GetUsers(ctx context.Context, req *userProto.GetUsersRequest, opts ...grpc.CallOption) (*userProto.GetUsersResponse, error) {
	return w.tenantAPI.GetUsers(ctx, req)
}

// PaymentServiceClientWrapper wraps the PaymentAPI to implement the PaymentServiceClient interface
type PaymentServiceClientWrapper struct {
	paymentAPI *paymentAPI.PaymentServer
}

func (w *PaymentServiceClientWrapper) CheckUserAccess(ctx context.Context, req *paymentProto.CheckUserAccessRequest, opts ...grpc.CallOption) (*paymentProto.CheckUserAccessResponse, error) {
	return w.paymentAPI.CheckUserAccess(ctx, req)
}

func (w *PaymentServiceClientWrapper) GetUserSubscription(ctx context.Context, req *paymentProto.GetUserSubscriptionRequest, opts ...grpc.CallOption) (*paymentProto.GetUserSubscriptionResponse, error) {
	return w.paymentAPI.GetUserSubscription(ctx, req)
}

func (w *PaymentServiceClientWrapper) CreateCheckoutSession(ctx context.Context, req *paymentProto.CreateCheckoutSessionRequest, opts ...grpc.CallOption) (*paymentProto.CreateCheckoutSessionResponse, error) {
	return w.paymentAPI.CreateCheckoutSession(ctx, req)
}

func (w *PaymentServiceClientWrapper) UpdateUserUsage(ctx context.Context, req *paymentProto.UpdateUserUsageRequest, opts ...grpc.CallOption) (*paymentProto.UpdateUserUsageResponse, error) {
	return w.paymentAPI.UpdateUserUsage(ctx, req)
}

func (w *PaymentServiceClientWrapper) InitializeUser(ctx context.Context, req *paymentProto.InitializeUserRequest, opts ...grpc.CallOption) (*paymentProto.InitializeUserResponse, error) {
	return w.paymentAPI.InitializeUser(ctx, req)
}

type Monolith struct {
	Config   *config.MonolithConfig
	Firebase *auth.Firebase

	VideoAPI       *videoAPI.VideoAPI
	CommentAPI     *commentAPI.CommentAPI
	UserAPI        *userAPI.UserAPI
	TenantAPI      *userAPI.TenantAPI
	ChannelAPI     *videoAPI.ChannelAPI
	PaymentAPI     *paymentAPI.PaymentServer
	PaymentHTTPAPI *paymentAPI.HTTPServer
	GRPCServer     *grpc.Server
	GRPCWebServer  *http.Server

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

	log.Info("Creating paymentservice API")
	// Create payment service database connection
	paymentDBConn, err := sql.Open(config.PaymentService.DB.Driver, config.PaymentService.DB.Url)
	if err != nil {
		log.Error("Could not create payment database", "err", err)
		return nil, err
	}

	// Run payment service migrations
	err = paymentDB.MigrateDB(config.PaymentService.DB.Driver, config.PaymentService.DB.Url)
	if err != nil {
		log.Error("Could not migrate payment database", "err", err)
		return nil, err
	}

	paymentQueries := paymentDB.New(paymentDBConn)
	paymentAPIServer := paymentAPI.NewPaymentServer(paymentQueries, &config.PaymentService)
	paymentHTTPServer := paymentAPI.NewHTTPServer(paymentQueries, &config.PaymentService)

	// Create payment service client wrapper for other services
	paymentServiceClientWrapper := &PaymentServiceClientWrapper{paymentAPI: paymentAPIServer}

	log.Info("Creating userservice API")
	userAPI, tenantAPI, err := userAPI.NewUserAPI(config.UserService, paymentServiceClientWrapper)
	if err != nil {
		log.Error("Could not create userservice API", "err", err)
		return nil, err
	}

	log.Info("Creating videoservice API")
	// Create wrapper to avoid circular dependency
	userServiceClientWrapper := &UserServiceClientWrapper{userAPI: userAPI}
	tenantServiceClientWrapper := &TenantServiceClientWrapper{tenantAPI: tenantAPI}
	videoAPI, channelAPI, err := videoAPI.NewVideoAPIProduction(config.VideoService, userServiceClientWrapper, tenantServiceClientWrapper, paymentServiceClientWrapper)
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

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		interceptors.PanicRecoveryInterceptor(),
		interceptors.FirebaseAuthInterceptor(firebase),
		interceptors.TenantInterceptor(),
	))

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

	// Create a handler for gRPC web requests with Firebase auth
	grpcWebHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrappedGrpc.ServeHTTP(w, r)
	})

	// Wrap the gRPC web handler with Firebase auth middleware
	authenticatedGrpcWebHandler := interceptors.FirebaseHTTPHeaderAuthMiddleware(firebase, grpcWebHandler)

	// Register webhook endpoint (no auth required for Stripe webhooks)
	parentMux.HandleFunc("/api/paymentservice/webhook/stripe", paymentHTTPServer.StripeWebhook)

	parentMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(r) || wrappedGrpc.IsAcceptableGrpcCorsRequest(r) {
			authenticatedGrpcWebHandler.ServeHTTP(w, r)
			return
		}

		// SPA fallback behavior: try to serve the requested file,
		// if it doesn't exist, serve index.html
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Try to open the requested file
		file, err := distFS.Open(path[1:]) // Remove leading slash
		if err != nil {
			// File doesn't exist, serve index.html for SPA routing
			indexFile, indexErr := distFS.Open("index.html")
			if indexErr != nil {
				http.Error(w, "index.html not found", http.StatusNotFound)
				return
			}
			defer indexFile.Close()

			// Get file info for modification time
			var modTime time.Time
			if stat, statErr := indexFile.Stat(); statErr == nil {
				modTime = stat.ModTime()
			} else {
				modTime = time.Now()
			}

			// Read the index.html content
			content, readErr := io.ReadAll(indexFile)
			if readErr != nil {
				http.Error(w, "failed to read index.html", http.StatusInternalServerError)
				return
			}

			// Set content type to HTML
			w.Header().Set("Content-Type", "text/html; charset=utf-8")

			// Serve index.html with proper HTTP caching support
			http.ServeContent(w, r, "index.html", modTime, bytes.NewReader(content))
			return
		}
		defer file.Close()

		// File exists, serve it normally
		staticFileServer.ServeHTTP(w, r)
	})

	parentMux.Handle("/api/videoservice/", http.StripPrefix("/api/videoservice", videoAPI.HTTPServerMux))

	httpServer := &http.Server{
		Addr:    config.Server.GrpcWebAddrPortString(),
		Handler: enableCORS(parentMux),
	}

	return &Monolith{
		Config:         &config,
		VideoAPI:       videoAPI,
		ChannelAPI:     channelAPI,
		CommentAPI:     commentAPI,
		UserAPI:        userAPI,
		TenantAPI:      tenantAPI,
		PaymentAPI:     paymentAPIServer,
		PaymentHTTPAPI: paymentHTTPServer,
		Firebase:       firebase,
		GRPCServer:     grpcServer,
		GRPCWebServer:  httpServer,
		log:            log,
	}, nil
}

func (m *Monolith) InitServices() error {

	m.log.Info("Initializing User Service")
	err := m.UserAPI.Init()
	if err != nil {
		return err
	}

	m.log.Info("Initializing Payment Service")
	err = m.PaymentAPI.Init(context.Background())
	if err != nil {
		return err
	}

	// Video service depends on user service

	m.log.Info("Initializing Video Service")
	err = m.VideoAPI.Init()
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

	m.log.Info("Starting User Service")
	err = m.UserAPI.Start()
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
	videoProto.RegisterChannelServiceServer(m.GRPCServer, m.ChannelAPI)
	commentProto.RegisterCommentServiceServer(m.GRPCServer, m.CommentAPI)
	userProto.RegisterUserServiceServer(m.GRPCServer, m.UserAPI)
	userProto.RegisterTenantServiceServer(m.GRPCServer, m.TenantAPI)
	paymentProto.RegisterPaymentServiceServer(m.GRPCServer, m.PaymentAPI)

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
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-User-Agent, X-Grpc-Web, x-tenant-id")

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
