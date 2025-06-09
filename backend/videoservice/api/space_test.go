package api

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
	"sortedstartup.com/stream/common/auth"
	"sortedstartup.com/stream/videoservice/db"
	"sortedstartup.com/stream/videoservice/proto"
)

// createTestVideoAPIWithDB creates a VideoAPI instance with an in-memory SQLite database for testing
func createTestVideoAPIWithDB(t *testing.T) *VideoAPI {
	// Create a temporary database file for migrations (in-memory doesn't work well with migrations)
	tempDB := "./test_" + t.Name() + ".db"
	testDatabase, err := sql.Open("sqlite", tempDB)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run migrations on the temporary database
	err = db.MigrateDB("sqlite", tempDB)
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	dbQueries := db.New(testDatabase)
	logger := slog.Default()

	// Cleanup function to remove test database file
	t.Cleanup(func() {
		testDatabase.Close()
		os.Remove(tempDB)
	})

	return &VideoAPI{
		db:        testDatabase,
		log:       logger,
		dbQueries: dbQueries,
	}
}

// Test space creation functionality
func TestCreateSpace(t *testing.T) {
	api := createTestVideoAPIWithDB(t)
	defer api.db.Close()

	authUser := &auth.AuthContext{
		User: &auth.User{
			ID:    "test-user-id",
			Name:  "Test User",
			Email: "test@example.com",
			Roles: []auth.Role{},
		},
	}
	ctx := context.WithValue(context.Background(), auth.AUTH_CONTEXT_KEY, authUser)

	space, err := api.CreateSpace(ctx, &proto.CreateSpaceRequest{
		Name:        "Test Space",
		Description: "A test space for unit testing",
	})

	assert.NoError(t, err, "Expected no error creating space")
	assert.NotNil(t, space, "Expected space to be created")
	assert.Equal(t, "Test Space", space.Name)
	assert.Equal(t, "A test space for unit testing", space.Description)
	assert.Equal(t, "test-user-id", space.UserId)
	assert.NotEmpty(t, space.Id, "Space ID should not be empty")
}

// Test listing spaces for a user
func TestListSpaces(t *testing.T) {
	api := createTestVideoAPIWithDB(t)
	defer api.db.Close()

	authUser := &auth.AuthContext{
		User: &auth.User{
			ID:    "test-user-id",
			Name:  "Test User",
			Email: "test@example.com",
			Roles: []auth.Role{},
		},
	}
	ctx := context.WithValue(context.Background(), auth.AUTH_CONTEXT_KEY, authUser)

	// Create a space first
	_, err := api.CreateSpace(ctx, &proto.CreateSpaceRequest{
		Name:        "Test Space 1",
		Description: "First test space",
	})
	assert.NoError(t, err)

	// List spaces
	response, err := api.ListSpaces(ctx, &proto.ListSpacesRequest{})

	assert.NoError(t, err, "Expected no error listing spaces")
	assert.NotNil(t, response, "Expected response")
	assert.GreaterOrEqual(t, len(response.Spaces), 1, "Expected at least 1 space")

	if len(response.Spaces) > 0 {
		assert.Equal(t, "Test Space 1", response.Spaces[0].Name)
		assert.Equal(t, "test-user-id", response.Spaces[0].UserId)
	}
}

// Test space sharing functionality
func TestSpaceSharing(t *testing.T) {
	api := createTestVideoAPIWithDB(t)
	defer api.db.Close()

	// Create two users
	owner := &auth.AuthContext{
		User: &auth.User{
			ID:    "owner-id",
			Name:  "Space Owner",
			Email: "owner@example.com",
			Roles: []auth.Role{},
		},
	}

	user2 := &auth.AuthContext{
		User: &auth.User{
			ID:    "user2-id",
			Name:  "User Two",
			Email: "user2@example.com",
			Roles: []auth.Role{},
		},
	}

	ownerCtx := context.WithValue(context.Background(), auth.AUTH_CONTEXT_KEY, owner)
	user2Ctx := context.WithValue(context.Background(), auth.AUTH_CONTEXT_KEY, user2)

	// Owner creates a space
	space, err := api.CreateSpace(ownerCtx, &proto.CreateSpaceRequest{
		Name:        "Shared Space",
		Description: "A space to be shared",
	})
	assert.NoError(t, err)

	// Add user2 to the space with view permissions
	_, err = api.AddUserToSpace(ownerCtx, &proto.AddUserToSpaceRequest{
		SpaceId:     space.Id,
		UserId:      "user2-id",
		AccessLevel: proto.AccessLevel_ACCESS_LEVEL_VIEW,
	})
	assert.NoError(t, err, "Expected no error adding user to space")

	// List space members
	members, err := api.ListSpaceMembers(ownerCtx, &proto.ListSpaceMembersRequest{
		SpaceId: space.Id,
	})
	assert.NoError(t, err, "Expected no error listing space members")
	assert.GreaterOrEqual(t, len(members.Members), 1, "Expected at least 1 member")

	// User2 should now see the shared space in their spaces list
	user2Spaces, err := api.ListSpaces(user2Ctx, &proto.ListSpacesRequest{})
	assert.NoError(t, err, "Expected no error listing spaces for user2")

	foundSharedSpace := false
	for _, s := range user2Spaces.Spaces {
		if s.Id == space.Id {
			foundSharedSpace = true
			assert.Equal(t, "view", s.AccessLevel, "User2 should have view access")
			break
		}
	}
	assert.True(t, foundSharedSpace, "User2 should see the shared space")
}

// Test permission levels for space access
func TestSpacePermissionLevels(t *testing.T) {
	api := createTestVideoAPIWithDB(t)
	defer api.db.Close()

	// Test different access levels
	accessLevels := []proto.AccessLevel{
		proto.AccessLevel_ACCESS_LEVEL_VIEW,
		proto.AccessLevel_ACCESS_LEVEL_EDIT,
		proto.AccessLevel_ACCESS_LEVEL_ADMIN,
	}

	owner := &auth.AuthContext{
		User: &auth.User{
			ID:    "owner-id",
			Name:  "Space Owner",
			Email: "owner@example.com",
			Roles: []auth.Role{},
		},
	}
	ownerCtx := context.WithValue(context.Background(), auth.AUTH_CONTEXT_KEY, owner)

	// Create a space
	space, err := api.CreateSpace(ownerCtx, &proto.CreateSpaceRequest{
		Name: "Permission Test Space",
	})
	assert.NoError(t, err)

	// Test adding users with different permission levels
	for i, level := range accessLevels {
		userID := "user-" + string(rune('1'+i))

		_, err = api.AddUserToSpace(ownerCtx, &proto.AddUserToSpaceRequest{
			SpaceId:     space.Id,
			UserId:      userID,
			AccessLevel: level,
		})
		assert.NoError(t, err, "Expected no error adding user with access level %v", level)
	}

	// List members to verify permissions were set correctly
	members, err := api.ListSpaceMembers(ownerCtx, &proto.ListSpaceMembersRequest{
		SpaceId: space.Id,
	})
	assert.NoError(t, err)

	// Should have at least 3 members (the users we just added)
	assert.GreaterOrEqual(t, len(members.Members), 3, "Expected at least 3 members")
}

// Test video access through shared spaces
func TestVideoAccessThroughSharedSpaces(t *testing.T) {
	api := createTestVideoAPIWithDB(t)
	defer api.db.Close()

	// Create owner and viewer users
	owner := &auth.AuthContext{
		User: &auth.User{
			ID:    "owner-id",
			Name:  "Video Owner",
			Email: "owner@example.com",
			Roles: []auth.Role{},
		},
	}

	viewer := &auth.AuthContext{
		User: &auth.User{
			ID:    "viewer-id",
			Name:  "Video Viewer",
			Email: "viewer@example.com",
			Roles: []auth.Role{},
		},
	}

	ownerCtx := context.WithValue(context.Background(), auth.AUTH_CONTEXT_KEY, owner)
	viewerCtx := context.WithValue(context.Background(), auth.AUTH_CONTEXT_KEY, viewer)

	// Owner creates a space
	space, err := api.CreateSpace(ownerCtx, &proto.CreateSpaceRequest{
		Name: "Video Sharing Space",
	})
	assert.NoError(t, err)

	// Add viewer to space with view permissions
	_, err = api.AddUserToSpace(ownerCtx, &proto.AddUserToSpaceRequest{
		SpaceId:     space.Id,
		UserId:      "viewer-id",
		AccessLevel: proto.AccessLevel_ACCESS_LEVEL_VIEW,
	})
	assert.NoError(t, err)

	// Note: In a real test, we would:
	// 1. Create a video as the owner
	// 2. Add the video to the space
	// 3. Verify that the viewer can access the video through the shared space
	// 4. Verify that the viewer cannot access the video directly (outside the space)

	// For now, we test that the space setup works correctly
	viewerSpaces, err := api.ListSpaces(viewerCtx, &proto.ListSpacesRequest{})
	assert.NoError(t, err)

	foundSpace := false
	for _, s := range viewerSpaces.Spaces {
		if s.Id == space.Id {
			foundSpace = true
			assert.Equal(t, "view", s.AccessLevel)
			break
		}
	}
	assert.True(t, foundSpace, "Viewer should have access to the shared space")
}

// Test access control - users can only modify spaces they own or have admin access to
func TestSpaceAccessControl(t *testing.T) {
	api := createTestVideoAPIWithDB(t)
	defer api.db.Close()

	owner := &auth.AuthContext{
		User: &auth.User{
			ID:    "owner-id",
			Name:  "Space Owner",
			Email: "owner@example.com",
			Roles: []auth.Role{},
		},
	}

	nonOwner := &auth.AuthContext{
		User: &auth.User{
			ID:    "nonowner-id",
			Name:  "Non Owner",
			Email: "nonowner@example.com",
			Roles: []auth.Role{},
		},
	}

	ownerCtx := context.WithValue(context.Background(), auth.AUTH_CONTEXT_KEY, owner)
	nonOwnerCtx := context.WithValue(context.Background(), auth.AUTH_CONTEXT_KEY, nonOwner)

	// Owner creates a space
	space, err := api.CreateSpace(ownerCtx, &proto.CreateSpaceRequest{
		Name: "Owner's Space",
	})
	assert.NoError(t, err)

	// Non-owner tries to add someone to the space (should fail)
	_, err = api.AddUserToSpace(nonOwnerCtx, &proto.AddUserToSpaceRequest{
		SpaceId:     space.Id,
		UserId:      "some-user-id",
		AccessLevel: proto.AccessLevel_ACCESS_LEVEL_VIEW,
	})
	assert.Error(t, err, "Non-owner should not be able to add users to space")

	// Non-owner tries to list space members (should fail)
	_, err = api.ListSpaceMembers(nonOwnerCtx, &proto.ListSpaceMembersRequest{
		SpaceId: space.Id,
	})
	assert.Error(t, err, "Non-owner should not be able to list space members")
}
