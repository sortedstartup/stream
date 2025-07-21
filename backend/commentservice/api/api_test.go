package api

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
	"sortedstartup.com/stream/commentservice/db"
	"sortedstartup.com/stream/commentservice/proto"
	"sortedstartup.com/stream/common/auth"
)

// MockDBQueries implements the db.Querier interface for testing.
type MockDBQueries struct {
	mock.Mock
}

// Implement all required methods from db.Querier
func (m *MockDBQueries) CheckUserLikedComment(ctx context.Context, params db.CheckUserLikedCommentParams) (int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDBQueries) CreateComment(ctx context.Context, params db.CreateCommentParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockDBQueries) DeleteComment(ctx context.Context, params db.DeleteCommentParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockDBQueries) GetAllCommentsByUserPaginated(ctx context.Context, params db.GetAllCommentsByUserPaginatedParams) ([]db.CommentserviceComment, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]db.CommentserviceComment), args.Error(1)
}

func (m *MockDBQueries) GetCommentByID(ctx context.Context, params db.GetCommentByIDParams) (db.CommentserviceComment, error) {
	args := m.Called(ctx, params)

	// Safely type assert to db.CommentserviceComment
	comment, ok := args.Get(0).(db.CommentserviceComment)
	if !ok {
		// If the value is a *db.CommentserviceComment, dereference it
		if commentPtr, ok := args.Get(0).(*db.CommentserviceComment); ok {
			return *commentPtr, args.Error(1)
		}
		// If it's neither, return an empty db.CommentserviceComment with an error
		return db.CommentserviceComment{}, fmt.Errorf("unexpected type for mock return value: %T", args.Get(0))
	}

	return comment, args.Error(1)
}

func (m *MockDBQueries) GetCommentCount(ctx context.Context, videoID string) (int64, error) {
	args := m.Called(ctx, videoID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDBQueries) GetCommentLikesCount(ctx context.Context, commentID string) (int64, error) {
	args := m.Called(ctx, commentID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDBQueries) GetCommentsByVideo(ctx context.Context, videoID string) ([]db.CommentserviceComment, error) {
	args := m.Called(ctx, videoID)
	return args.Get(0).([]db.CommentserviceComment), args.Error(1)
}

func (m *MockDBQueries) GetCommentsByVideoPaginated(ctx context.Context, params db.GetCommentsByVideoPaginatedParams) ([]db.CommentserviceComment, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]db.CommentserviceComment), args.Error(1)
}

func (m *MockDBQueries) GetRepliesByCommentID(ctx context.Context, commentID sql.NullString) ([]db.CommentserviceComment, error) {
	args := m.Called(ctx, commentID)
	return args.Get(0).([]db.CommentserviceComment), args.Error(1)
}

func (m *MockDBQueries) GetComentsAndRepliesForVideoID(ctx context.Context, videoID string) ([]db.GetComentsAndRepliesForVideoIDRow, error) {
	args := m.Called(ctx, videoID)
	return args.Get(0).([]db.GetComentsAndRepliesForVideoIDRow), args.Error(1)
}

func (m *MockDBQueries) test(ctx context.Context) ([]db.CommentserviceComment, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.CommentserviceComment), args.Error(1)
}

func (m *MockDBQueries) LikeComment(ctx context.Context, params db.LikeCommentParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockDBQueries) UnlikeComment(ctx context.Context, params db.UnlikeCommentParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockDBQueries) UpdateComment(ctx context.Context, params db.UpdateCommentParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

// Test CreateComment
func TestCreateComment(t *testing.T) {
	mockDB := new(MockDBQueries)
	logger := slog.Default()

	commentAPI := NewCommentAPITest(mockDB, logger)
	assert.NotNil(t, commentAPI, "commentAPI should not be nil")

	authUser := &auth.AuthContext{
		User: &auth.User{
			ID:    "test-user-id",
			Name:  "Test User",
			Email: "test@example.com",
			Roles: []auth.Role{},
		},
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, auth.AUTH_CONTEXT_KEY, authUser)

	authToken := "Bearer test-token"
	md := metadata.Pairs("authorization", authToken)
	ctx = metadata.NewOutgoingContext(ctx, md)

	mockDB.On("CreateComment", mock.Anything, mock.MatchedBy(func(params db.CreateCommentParams) bool {
		return params.Content == "This is a test comment" &&
			params.VideoID == "test-video-id" &&
			params.UserID == "test-user-id" &&
			!params.ParentCommentID.Valid
	})).Return(nil).Once()

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

	mockDB.AssertExpectations(t)
}

// Test ListComments
func TestListComments(t *testing.T) {
	mockDB := new(MockDBQueries)
	logger := slog.Default()
	commentAPI := NewCommentAPITest(mockDB, logger)

	authUser := &auth.AuthContext{
		User: &auth.User{
			ID:    "test-user-id",
			Name:  "Test User",
			Email: "test@example.com",
			Roles: []auth.Role{},
		},
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, auth.AUTH_CONTEXT_KEY, authUser)

	authToken := "Bearer test-token"
	md := metadata.Pairs("authorization", authToken)
	ctx = metadata.NewOutgoingContext(ctx, md)

	mockComments := []db.CommentserviceComment{
		{ID: "123", Content: "Test Comment", VideoID: "vid1", UserID: "user1"},
	}
	mockDB.On("GetAllCommentsByUserPaginated", ctx, mock.Anything).Return(mockComments, nil)

	resp, err := commentAPI.ListComments(ctx, &proto.ListCommentsRequest{
		PageSize:   10,
		PageNumber: 1,
	})

	t.Logf("Received response: %+v", resp)
	t.Logf("Received error: %v", err)

	assert.NoError(t, err, "Expected no error but got one")
	assert.NotNil(t, resp, "Expected response but got nil")
	assert.Len(t, resp.Comments, 1, "Expected exactly 1 comment")

	assert.Equal(t, "Test Comment", resp.Comments[0].Content)

	mockDB.AssertExpectations(t)
}

// Test GetCommentfunc
func TestGetComment(t *testing.T) {
	mockDB := new(MockDBQueries)
	logger := slog.Default()
	commentAPI := NewCommentAPITest(mockDB, logger)

	authUser := &auth.AuthContext{
		User: &auth.User{
			ID:    "test-user-id",
			Name:  "Test User",
			Email: "test@example.com",
			Roles: []auth.Role{},
		},
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, auth.AUTH_CONTEXT_KEY, authUser)

	authToken := "Bearer test-token"
	md := metadata.Pairs("authorization", authToken)
	ctx = metadata.NewOutgoingContext(ctx, md)

	mockComment := db.CommentserviceComment{
		ID:      "mock-id",
		Content: "Mock comment",
		VideoID: "vid1",
		UserID:  "test-user-id",
	}

	mockDB.On("GetCommentByID", ctx, mock.Anything).Return(mockComment, nil)

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

	mockDB.AssertExpectations(t)
}
