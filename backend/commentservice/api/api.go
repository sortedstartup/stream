package api

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	_ "modernc.org/sqlite"
	"sortedstartup.com/stream/commentservice/config"
	"sortedstartup.com/stream/commentservice/db"
	"sortedstartup.com/stream/commentservice/proto"
	"sortedstartup.com/stream/common/interceptors"
)

type CommentAPI struct {
	config        config.CommentServiceConfig
	HTTPServerMux *http.ServeMux
	db            *sql.DB

	log       *slog.Logger
	dbQueries db.Querier

	//implemented proto server
	proto.UnimplementedCommentServiceServer
}

// NewCommentAPITest creates a CommentAPI instance with a mock database for testing.
func NewCommentAPITest(mockDB db.Querier, logger *slog.Logger) *CommentAPI {
	return &CommentAPI{
		log:       logger,
		dbQueries: mockDB, // Use the sqlc-generated Querier interface
	}
}

func NewCommentAPIProduction(config config.CommentServiceConfig) (*CommentAPI, error) {
	slog.Info("NewCommentAPIProduction")

	// fbAuth, err := auth.NewFirebase()
	// if err != nil {
	// 	return nil, err
	// }

	childLogger := slog.With("service", "CommentAPI")

	_db, err := sql.Open(config.DB.Driver, config.DB.Url)
	if err != nil {
		return nil, err
	}

	dbQueries := db.New(_db)

	commentAPI := &CommentAPI{
		config:    config,
		db:        _db,
		log:       childLogger,
		dbQueries: dbQueries,
	}

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

func (s *CommentAPI) CreateComment(ctx context.Context, req *proto.CreateCommentRequest) (*proto.Comment, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	commentID := generateUUID()

	// Check if ParentCommentId is nil
	var parentCommentID sql.NullString
	if req.ParentCommentId != nil {
		parentCommentID = sql.NullString{String: *req.ParentCommentId, Valid: *req.ParentCommentId != ""}
	}

	err = s.dbQueries.CreateComment(ctx, db.CreateCommentParams{
		ID:              commentID,
		Content:         req.Content,
		VideoID:         req.VideoId,
		UserID:          authContext.User.ID,
		ParentCommentID: parentCommentID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create comment: %v", err)
	}

	return &proto.Comment{
		Id:      commentID,
		Content: req.Content,
		VideoId: req.VideoId,
		UserId:  authContext.User.ID,
	}, nil
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
			UserId:  comment.UserID,
		})
	}

	return &proto.ListCommentsResponse{Comments: protoComments}, nil
}

func (s *CommentAPI) GetComment(ctx context.Context, req *proto.GetCommentRequest) (*proto.GetCommentResponse, error) {
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

	// ✅ Return GetCommentResponse instead of just Comment
	return &proto.GetCommentResponse{
		Comment: &proto.Comment{
			Id:      comment.ID,
			Content: comment.Content,
			VideoId: comment.VideoID,
			UserId:  comment.UserID,
		},
	}, nil
}

func generateUUID() string {
	return uuid.New().String()
}
