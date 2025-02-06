package api

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	_ "modernc.org/sqlite"
	"sortedstartup.com/stream/common/auth"
	"sortedstartup.com/stream/common/interceptors"
	"sortedstartup.com/stream/commentservice/config"
	"sortedstartup.com/stream/commentservice/db"
	"sortedstartup.com/stream/commentservice/proto"
)

type CommentAPI struct {
	config        config.CommentServiceConfig
	HTTPServerMux *http.ServeMux
	db            *sql.DB

	log       *slog.Logger
	dbQueries *db.Queries

	//implemented proto server
	proto.UnimplementedCommentServiceServer
}

func NewCommentAPIProduction(config config.CommentServiceConfig) (*CommentAPI, error) {
	slog.Info("NewCommentAPIProduction")

	fbAuth, err := auth.NewFirebase()
	if err != nil {
		return nil, err
	}

	childLogger := slog.With("service", "CommentAPI")

	_db, err := sql.Open(config.DB.Driver, config.DB.Url)
	if err != nil {
		return nil, err
	}

	dbQueries := db.New(_db)

	ServerMux := http.NewServeMux()

	commentAPI := &CommentAPI{
		HTTPServerMux: ServerMux,
		config:        config,
		db:            _db,
		log:           childLogger,
		dbQueries:     dbQueries,
	}

	ServerMux.Handle("/comment", interceptors.FirebaseHTTPAuthMiddleware(fbAuth, http.HandlerFunc(commentAPI.commentHandler)))

	return commentAPI, nil
}

func (s *CommentAPI) Start() error {
	return nil
}

func (s *CommentAPI) Init() error {
	s.log.Info("Migrating database", "dbDriver", s.config.DB.Driver, "dbURL", s.config.DB.Url)
	err := db.MigrateDB(s.config.DB.Driver, s.config.DB.Url)
	if err != nil {
		return err
	}
	s.log.Info("Migrating database done")
	return nil
}

func (s *CommentAPI) ListComments(ctx context.Context, req *proto.ListCommentsRequest) (*proto.ListCommentsResponse, error) {

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

	comments, err := s.dbQueries.GetAllCommentsByUserPaginated(ctx, db.GetAllCommentsByUserPaginatedParams{
		UserID:     userID,
		PageSize:   int64(pageSize),
		PageNumber: int64(pageNumber),
	})
	if err != nil {
		slog.Error("Error getting comments", "err", err)
		return nil, err
	}

	protoComments := make([]*proto.Comment, 0, len(comments))

	for _, comment := range comments {
		protoComments = append(protoComments, &proto.Comment{
			Id:      comment.ID,
			Content: comment.Content,
			VideoId: comment.VideoID,
		})
	}

	return &proto.ListCommentsResponse{Comments: protoComments}, nil
}

func (s *CommentAPI) GetComment(ctx context.Context, req *proto.GetCommentRequest) (*proto.Comment, error) {
	// Get auth context to verify user has access
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		s.log.Error("Error getting auth from context", "err", err)
		return nil, err
	}

	// Get comment from database
	comment, err := s.dbQueries.GetCommentByID(ctx, db.GetCommentByIDParams{
		ID:     req.CommentId,
		UserID: authContext.User.ID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "comment not found")
		}
		s.log.Error("Error getting comment", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	// Verify user has access to this comment
	if comment.UserID != authContext.User.ID {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	// Convert to proto message
	return &proto.Comment{
		Id:      comment.ID,
		Content: comment.Content,
		VideoId: comment.VideoID,
	}, nil
}

func (api *CommentAPI) commentHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract comment content from request body
	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Extract video ID from request body
	videoID := r.FormValue("video_id")
	if videoID == "" {
		http.Error(w, "Video ID is required", http.StatusBadRequest)
		return
	}

	// Get auth context to verify user
	authContext, err := interceptors.AuthFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		slog.Error("Unauthorized", "err", err)
		return
	}
	userID := authContext.User.ID

	// Create a new comment in the database
	commentID := uuid.New().String()
	err = api.dbQueries.CreateComment(r.Context(), db.CreateCommentParams{
		ID:      commentID,
		Content: content,
		VideoID: videoID,
		UserID:  userID,
	})
	if err != nil {
		slog.Error("Failed to add comment to the database", "err", err)
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "Comment added successfully", "comment_id": "%s"}`, commentID)))
}
