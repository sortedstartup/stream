package api_test

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	lru "github.com/hashicorp/golang-lru"
	"github.com/stretchr/testify/assert"
	"sortedstartup.com/stream/common/auth"
	"sortedstartup.com/stream/userservice/api"
	"sortedstartup.com/stream/userservice/db"
	"sortedstartup.com/stream/userservice/db/mocks"
	"sortedstartup.com/stream/userservice/proto"
)

func withAuthContext(ctx context.Context, user *auth.User) context.Context {
	return context.WithValue(ctx, auth.AUTH_CONTEXT_KEY, &auth.AuthContext{User: user})
}

func TestCreateUserIfNotExists_CreatesUserWhenNotExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock Querier once, shared for userAPI and tenantAPI
	mockQuerier := mocks.NewMockQuerier(ctrl)

	// Auth user for context
	testUser := &auth.User{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	ctx := withAuthContext(context.Background(), testUser)

	// Create LRU cache
	cache, _ := lru.New(128)

	// Set up expected calls for UserAPI flow:

	// GetUserByEmail returns no rows → triggers CreateUser
	mockQuerier.EXPECT().
		GetUserByEmail(gomock.Any(), testUser.Email).
		Return(db.UserserviceUser{}, sql.ErrNoRows)

	// CreateUser returns a user object converted from params
	mockQuerier.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CreateUserParams) (db.UserserviceUser, error) {
			return db.UserserviceUser(params), nil
		})

	// Now mock calls needed for createPersonalTenant inside TenantAPI:

	// Mock CreateTenant (returns a valid Tenant struct)
	mockQuerier.EXPECT().
		CreateTenant(gomock.Any(), gomock.Any()).
		Return(db.UserserviceTenant{
			ID:          "tenant-123",
			Name:        testUser.Name,
			Description: sql.NullString{String: "Personal workspace", Valid: true},
			IsPersonal:  true,
			CreatedAt:   time.Now(),
			CreatedBy:   testUser.ID,
		}, nil)

	// Mock CreateTenantUser (returns valid TenantUser)
	mockQuerier.EXPECT().
		CreateTenantUser(gomock.Any(), gomock.Any()).
		Return(db.UserserviceTenantUser{
			ID:        "tenantuser-123",
			TenantID:  "tenant-123",
			UserID:    testUser.ID,
			Role:      "super_admin",
			CreatedAt: time.Now(),
		}, nil)

	logger := slog.Default()
	tenantAPI := api.NewTenantAPITest(mockQuerier, logger)
	userAPI := api.NewUserAPITest(mockQuerier, cache, tenantAPI, logger)

	// Call the method under test
	resp, err := userAPI.CreateUserIfNotExists(ctx, &proto.CreateUserRequest{})

	// Assert no error and expected response
	assert.NoError(t, err)
	assert.Equal(t, "User created successfully", resp.Message)
	assert.Equal(t, testUser.Email, resp.User.Email)
}

func TestCreateUserIfNotExists_UserAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mocks.NewMockQuerier(ctrl)
	cache, _ := lru.New(128)
	logger := slog.Default()
	user := &auth.User{ID: "123", Email: "test@example.com", Name: "Tester"}
	ctx := withAuthContext(context.Background(), user)

	mockQuerier.EXPECT().
		GetUserByEmail(gomock.Any(), user.Email).
		Return(db.UserserviceUser{ID: user.ID, Email: user.Email, Username: user.Email, CreatedAt: time.Now()}, nil)

	tenantAPI := api.NewTenantAPITest(mockQuerier, logger)
	userAPI := api.NewUserAPITest(mockQuerier, cache, tenantAPI, logger)

	resp, err := userAPI.CreateUserIfNotExists(ctx, &proto.CreateUserRequest{})
	assert.NoError(t, err)
	assert.Equal(t, "User already exists", resp.Message)
	assert.Equal(t, user.Email, resp.User.Email)
}

func TestCreateUserIfNotExists_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mocks.NewMockQuerier(ctrl)
	cache, _ := lru.New(128)
	logger := slog.Default()
	user := &auth.User{ID: "123", Email: "test@example.com", Name: "Tester"}
	ctx := withAuthContext(context.Background(), user)
	cache.Add(user.Email, true)

	tenantAPI := api.NewTenantAPITest(mockQuerier, logger)
	userAPI := api.NewUserAPITest(mockQuerier, cache, tenantAPI, logger)

	resp, err := userAPI.CreateUserIfNotExists(ctx, &proto.CreateUserRequest{})
	assert.NoError(t, err)
	assert.Equal(t, "User already exists (cache)", resp.Message)
}

func TestCreateUserIfNotExists_GetUserByEmailFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mocks.NewMockQuerier(ctrl)
	cache, _ := lru.New(128)
	logger := slog.Default()
	user := &auth.User{ID: "123", Email: "test@example.com", Name: "Tester"}
	ctx := withAuthContext(context.Background(), user)

	tenantAPI := api.NewTenantAPITest(mockQuerier, logger)
	userAPI := api.NewUserAPITest(mockQuerier, cache, tenantAPI, logger)

	mockQuerier.EXPECT().
		GetUserByEmail(gomock.Any(), user.Email).
		Return(db.UserserviceUser{}, errors.New("db connection error"))

	resp, err := userAPI.CreateUserIfNotExists(ctx, &proto.CreateUserRequest{})
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "internal")
}

func TestCreateUserIfNotExists_CreateUserFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mocks.NewMockQuerier(ctrl)
	cache, _ := lru.New(128)
	logger := slog.Default()
	user := &auth.User{ID: "123", Email: "test@example.com", Name: "Tester"}
	ctx := withAuthContext(context.Background(), user)

	tenantAPI := api.NewTenantAPITest(mockQuerier, logger)
	userAPI := api.NewUserAPITest(mockQuerier, cache, tenantAPI, logger)

	mockQuerier.EXPECT().
		GetUserByEmail(gomock.Any(), user.Email).
		Return(db.UserserviceUser{}, sql.ErrNoRows)

	mockQuerier.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return(db.UserserviceUser{}, errors.New("insert failed"))

	resp, err := userAPI.CreateUserIfNotExists(ctx, &proto.CreateUserRequest{})
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create user")
}
