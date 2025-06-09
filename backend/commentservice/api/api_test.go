package api

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	_ "modernc.org/sqlite"
	"sortedstartup.com/stream/commentservice/db"
	"sortedstartup.com/stream/commentservice/proto"
	"sortedstartup.com/stream/common/auth"
)

// createTestCommentAPI creates a CommentAPI instance with an in-memory SQLite database for testing
func createTestCommentAPI(t *testing.T) *CommentAPI {
	// Create a temporary database file for migrations (in-memory doesn't work well with migrations)
	tempDB := "./test_comment_" + t.Name() + ".db"
	database, err := sql.Open("sqlite", tempDB)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations on the temporary database
	err = db.MigrateDB("sqlite", tempDB)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	dbQueries := db.New(database)
	logger := slog.Default()

	// Cleanup function to remove test database file
	t.Cleanup(func() {
		database.Close()
		os.Remove(tempDB)
	})

	return &CommentAPI{
		db:        database,
		log:       logger,
		dbQueries: dbQueries,
	}
}

// Test CreateComment
func TestCreateComment(t *testing.T) {
	commentAPI := createTestCommentAPI(t)
	defer commentAPI.db.Close()

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
		assert.NotEmpty(t, comment.Id, "Comment ID should not be empty")
	}
}

// Test ListComments
func TestListComments(t *testing.T) {
	commentAPI := createTestCommentAPI(t)
	defer commentAPI.db.Close()

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

	// First create a comment
	_, err := commentAPI.CreateComment(ctx, &proto.CreateCommentRequest{
		Content: "Test Comment for Listing",
		VideoId: "vid1",
	})
	assert.NoError(t, err)

	// Now list comments for the video
	resp, err := commentAPI.ListComments(ctx, &proto.ListCommentsRequest{
		VideoId: "vid1",
	})

	t.Logf("Received response: %+v", resp)
	t.Logf("Received error: %v", err)

	assert.NoError(t, err, "Expected no error but got one")
	assert.NotNil(t, resp, "Expected response but got nil")
	assert.GreaterOrEqual(t, len(resp.Comments), 1, "Expected at least 1 comment")

	if len(resp.Comments) > 0 {
		assert.Equal(t, "Test Comment for Listing", resp.Comments[0].Content)
		assert.Equal(t, "vid1", resp.Comments[0].VideoId)
	}
}

// Test GetComment
func TestGetComment(t *testing.T) {
	commentAPI := createTestCommentAPI(t)
	defer commentAPI.db.Close()

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

	// First create a comment
	createdComment, err := commentAPI.CreateComment(ctx, &proto.CreateCommentRequest{
		Content: "Mock comment for GetComment test",
		VideoId: "vid1",
	})
	assert.NoError(t, err)
	assert.NotNil(t, createdComment)

	// Now get the comment
	resp, err := commentAPI.GetComment(ctx, &proto.GetCommentRequest{
		CommentId: createdComment.Id,
	})

	t.Logf("Received response: %+v", resp)
	t.Logf("Received error: %v", err)

	assert.NoError(t, err, "Expected no error but got one")
	assert.NotNil(t, resp, "Expected response but got nil")

	assert.NotNil(t, resp.Comment, "Expected Comment object but got nil")
	assert.Equal(t, createdComment.Id, resp.Comment.Id, "Expected CommentId to match")
	assert.Equal(t, "Mock comment for GetComment test", resp.Comment.Content, "Expected Content to match")
}

// Test for space sharing functionality (new tests)
func TestCommentPermissions(t *testing.T) {
	commentAPI := createTestCommentAPI(t)
	defer commentAPI.db.Close()

	// Test that users can only access their own comments
	authUser1 := &auth.AuthContext{
		User: &auth.User{
			ID:    "user-1",
			Name:  "User One",
			Email: "user1@example.com",
			Roles: []auth.Role{},
		},
	}

	authUser2 := &auth.AuthContext{
		User: &auth.User{
			ID:    "user-2",
			Name:  "User Two",
			Email: "user2@example.com",
			Roles: []auth.Role{},
		},
	}

	ctx1 := context.WithValue(context.Background(), auth.AUTH_CONTEXT_KEY, authUser1)
	ctx2 := context.WithValue(context.Background(), auth.AUTH_CONTEXT_KEY, authUser2)

	// User 1 creates a comment
	comment, err := commentAPI.CreateComment(ctx1, &proto.CreateCommentRequest{
		Content: "User 1's comment",
		VideoId: "test-video",
	})
	assert.NoError(t, err)

	// User 2 tries to access User 1's comment - should fail
	_, err = commentAPI.GetComment(ctx2, &proto.GetCommentRequest{
		CommentId: comment.Id,
	})
	assert.Error(t, err, "User 2 should not be able to access User 1's comment")
}
