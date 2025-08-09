package api

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"log/slog"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"

	"sortedstartup.com/stream/commentservice/db"
	mockdb "sortedstartup.com/stream/commentservice/db/mocks"
	"sortedstartup.com/stream/commentservice/proto"
	"sortedstartup.com/stream/common/auth"
)

// Helper to build an auth context
func buildAuthContext() context.Context {
	user := &auth.AuthContext{
		User: &auth.User{
			ID:    "test-user-id",
			Name:  "Test User",
			Email: "test@example.com",
			Roles: []auth.Role{},
		},
	}

	ctx := context.WithValue(context.Background(), auth.AUTH_CONTEXT_KEY, user)
	md := metadata.Pairs("authorization", "Bearer test-token")
	ctx = metadata.NewOutgoingContext(ctx, md)

	return ctx
}

// Test CreateComment
func TestCreateComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)
	logger := slog.Default()

	commentAPI := NewCommentAPITest(mockDB, logger)
	assert.NotNil(t, commentAPI, "commentAPI should not be nil")

	ctx := buildAuthContext()

	mockDB.EXPECT().
		CreateComment(gomock.Any(), gomock.AssignableToTypeOf(db.CreateCommentParams{})).
		DoAndReturn(func(ctx context.Context, params db.CreateCommentParams) error {
			assert.Equal(t, "This is a test comment", params.Content)
			assert.Equal(t, "test-video-id", params.VideoID)
			assert.Equal(t, "test-user-id", params.UserID)
			assert.False(t, params.ParentCommentID.Valid)
			return nil
		}).
		Times(1)

	comment, err := commentAPI.CreateComment(ctx, &proto.CreateCommentRequest{
		Content: "This is a test comment",
		VideoId: "test-video-id",
	})

	t.Logf("Received comment: %+v", comment)
	t.Logf("Received error: %v", err)

	assert.NoError(t, err, "Expected no error but got one")
	assert.NotNil(t, comment, "Expected comment but got nil")

	if comment != nil {
		assert.Equal(t, "This is a test comment", comment.Content)
		assert.Equal(t, "test-video-id", comment.VideoId)
		assert.Equal(t, "test-user-id", comment.UserId)
	}
}

// Test ListComments
func TestListComments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)
	logger := slog.Default()

	commentAPI := NewCommentAPITest(mockDB, logger)
	ctx := buildAuthContext()

	mockDB.EXPECT().
		GetComentsAndRepliesForVideoID(gomock.Any(), "test-video-id").
		Return([]db.GetComentsAndRepliesForVideoIDRow{
			{
				ID:        "comment-1",
				Content:   "Test comment 1",
				VideoID:   "test-video-id",
				UserID:    "user-1",
				Username:  sqlNullString("User1"),
				CreatedAt: getNow().Time,
				UpdatedAt: getNow().Time,
				Replies:   `[]`,
			},
		}, nil).
		Times(1)

	resp, err := commentAPI.ListComments(ctx, &proto.ListCommentsRequest{
		VideoId: "test-video-id",
	})

	t.Logf("Received response: %+v", resp)
	t.Logf("Received error: %v", err)

	assert.NoError(t, err, "Expected no error but got one")
	assert.NotNil(t, resp, "Expected response but got nil")
	assert.Len(t, resp.Comments, 1, "Expected exactly 1 comment")

	if len(resp.Comments) > 0 {
		assert.Equal(t, "Test comment 1", resp.Comments[0].Content)
		assert.Equal(t, "test-video-id", resp.Comments[0].VideoId)
	}
}

// Test GetComment
func TestGetComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)
	logger := slog.Default()

	commentAPI := NewCommentAPITest(mockDB, logger)
	ctx := buildAuthContext()

	mockComment := db.Comment{
		ID:      "mock-id",
		Content: "Mock comment",
		VideoID: "vid1",
		UserID:  "test-user-id",
		UserID:  "test-user-id",
	}

	mockDB.EXPECT().
		GetCommentByID(gomock.Any(), gomock.Any()).
		Return(mockComment, nil).
		Times(1)

	resp, err := commentAPI.GetComment(ctx, &proto.GetCommentRequest{
		CommentId: "mock-id",
	})

	t.Logf("Received response: %+v", resp)
	t.Logf("Received error: %v", err)

	assert.NoError(t, err, "Expected no error but got one")
	assert.NotNil(t, resp, "Expected response but got nil")

	assert.NotNil(t, resp.Comment, "Expected Comment object but got nil")
	assert.Equal(t, "mock-id", resp.Comment.Id, "Expected CommentId to match")
	assert.Equal(t, "Mock comment", resp.Comment.Content, "Expected Content to match")
}

func TestCreateComment_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)
	logger := slog.Default()
	commentAPI := NewCommentAPITest(mockDB, logger)
	ctx := buildAuthContext()

	mockDB.EXPECT().
		CreateComment(gomock.Any(), gomock.Any()).
		Return(fmt.Errorf("db failure")).
		Times(1)

	comment, err := commentAPI.CreateComment(ctx, &proto.CreateCommentRequest{
		Content: "This is a test comment",
		VideoId: "test-video-id",
	})

	assert.Error(t, err)
	assert.Nil(t, comment)
	assert.Contains(t, err.Error(), "failed to create comment")
}

func TestListComments_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)
	logger := slog.Default()
	commentAPI := NewCommentAPITest(mockDB, logger)
	ctx := buildAuthContext()

	mockDB.EXPECT().
		GetComentsAndRepliesForVideoID(gomock.Any(), "test-video-id").
		Return(nil, fmt.Errorf("db failure")).
		Times(1)

	resp, err := commentAPI.ListComments(ctx, &proto.ListCommentsRequest{
		VideoId: "test-video-id",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to fetch comments")
}

func TestGetComment_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)
	logger := slog.Default()
	commentAPI := NewCommentAPITest(mockDB, logger)
	ctx := buildAuthContext()

	mockDB.EXPECT().
		GetCommentByID(gomock.Any(), gomock.Any()).
		Return(db.Comment{}, sql.ErrNoRows).
		Times(1)

	resp, err := commentAPI.GetComment(ctx, &proto.GetCommentRequest{
		CommentId: "nonexistent-id",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "comment not found")
}

func TestGetComment_PermissionDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)
	logger := slog.Default()
	commentAPI := NewCommentAPITest(mockDB, logger)

	ctx := buildAuthContext()

	mockComment := db.Comment{
		ID:      "mock-id",
		Content: "Mock comment",
		VideoID: "vid1",
		UserID:  "other-user-id", // different user id
	}

	mockDB.EXPECT().
		GetCommentByID(gomock.Any(), gomock.Any()).
		Return(mockComment, nil).
		Times(1)

	resp, err := commentAPI.GetComment(ctx, &proto.GetCommentRequest{
		CommentId: "mock-id",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestCreateComment_WithParentID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)
	logger := slog.Default()

	commentAPI := NewCommentAPITest(mockDB, logger)
	ctx := buildAuthContext()

	mockDB.EXPECT().
		CreateComment(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CreateCommentParams) error {
			if params.Content != "Reply comment" {
				t.Errorf("unexpected content: got %s", params.Content)
				return fmt.Errorf("unexpected content")
			}
			if params.VideoID != "test-video-id" {
				t.Errorf("unexpected videoID: got %s", params.VideoID)
				return fmt.Errorf("unexpected videoID")
			}
			if params.UserID != "test-user-id" {
				t.Errorf("unexpected userID: got %s", params.UserID)
				return fmt.Errorf("unexpected userID")
			}
			if !params.ParentCommentID.Valid || params.ParentCommentID.String != "parent-comment-id" {
				t.Errorf("unexpected parentCommentID: got %v", params.ParentCommentID)
				return fmt.Errorf("unexpected parentCommentID")
			}
			return nil
		}).
		Times(1)

	parentCommentID := "parent-comment-id"
	_, err := commentAPI.CreateComment(ctx, &proto.CreateCommentRequest{
		Content:         "Reply comment",
		VideoId:         "test-video-id",
		ParentCommentId: &parentCommentID,
	})

	if err != nil {
		t.Fatalf("CreateComment failed: %v", err)
	}
}

func TestListComments_MalformedRepliesJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)
	logger := slog.Default()

	commentAPI := NewCommentAPITest(mockDB, logger)
	ctx := buildAuthContext()

	mockDB.EXPECT().
		GetComentsAndRepliesForVideoID(gomock.Any(), "test-video-id").
		Return([]db.GetComentsAndRepliesForVideoIDRow{
			{
				ID:        "comment-1",
				Content:   "Test comment 1",
				VideoID:   "test-video-id",
				UserID:    "user-1",
				Username:  sqlNullString("User1"),
				CreatedAt: getNow().Time,
				UpdatedAt: getNow().Time,
				Replies:   "malformed json here", // Invalid JSON string intentionally
			},
		}, nil).
		Times(1)

	resp, err := commentAPI.ListComments(ctx, &proto.ListCommentsRequest{
		VideoId: "test-video-id",
	})

	// Assert error is returned because JSON was malformed
	assert.Error(t, err)
	// Response should be nil due to error
	assert.Nil(t, resp)
}

// --- Helper functions ---

func sqlNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func getNow() sql.NullTime {
	return sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
}
