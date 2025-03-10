package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	// Fetch comments and their replies for the given video ID
	commentsWithReplies, err := s.dbQueries.GetComentsAndRepliesForVideoID(ctx, req.VideoId)
	if err != nil {
		s.log.Error("Error fetching comments and replies", "err", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch comments: %v", err)
	}

	// Convert database model to proto response
	var protoComments []*proto.Comment
	for _, comment := range commentsWithReplies {
		var replies []struct {
			ID              string    `json:"id"`
			Content         string    `json:"content"`
			UserID          string    `json:"user_id"`
			VideoID         string    `json:"video_id"`
			ParentCommentID string    `json:"parent_comment_id"`
			CreatedAt       time.Time `json:"created_at"` // Now using time.Time
			UpdatedAt       time.Time `json:"updated_at"` // Now using time.Time
		}

		// Ensure Replies is a valid and JSON-encoded string
		if repliesJSON, ok := comment.Replies.(string); ok && repliesJSON != "" {
			err := json.Unmarshal([]byte(repliesJSON), &replies)
			if err != nil {
				s.log.Error("Error unmarshalling replies JSON", "err", err)
				return nil, status.Errorf(codes.Internal, "failed to parse replies: %v", err)
			}
		}

		// Convert time.Time to protobuf timestamps
		createdAtProto := timestamppb.New(comment.CreatedAt)
		updatedAtProto := timestamppb.New(comment.UpdatedAt)

		// Convert replies to proto format
		var protoReplies []*proto.Comment
		for _, r := range replies {
			protoReplies = append(protoReplies, &proto.Comment{
				Id:              r.ID,
				Content:         r.Content,
				UserId:          r.UserID,
				VideoId:         r.VideoID,
				ParentCommentId: r.ParentCommentID,
				CreatedAt:       timestamppb.New(r.CreatedAt),
				UpdatedAt:       timestamppb.New(r.UpdatedAt),
			})
		}

		// Append the main comment
		protoComments = append(protoComments, &proto.Comment{
			Id:              comment.ID,
			Content:         comment.Content,
			VideoId:         comment.VideoID,
			UserId:          comment.UserID,
			ParentCommentId: comment.ParentCommentID.String,
			CreatedAt:       createdAtProto,
			UpdatedAt:       updatedAtProto,
			Replies:         protoReplies,
		})
	}

	return &proto.ListCommentsResponse{
		Comments: protoComments,
	}, nil
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
