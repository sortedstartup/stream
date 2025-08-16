package api_test

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sortedstartup.com/stream/common/auth"
	"sortedstartup.com/stream/common/constants"
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

func TestCreatePersonalTenant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mocks.NewMockQuerier(ctrl)
	logger := slog.Default()
	tenantAPI := api.NewTenantAPITest(mockQuerier, logger)

	testUser := &auth.User{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Tester",
	}
	ctx := withAuthContext(context.Background(), testUser)

	t.Run("success - tenant and tenantUser created", func(t *testing.T) {
		mockQuerier.EXPECT().
			CreateTenant(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, p db.CreateTenantParams) (db.UserserviceTenant, error) {
				return db.UserserviceTenant{
					ID:          p.ID,
					Name:        p.Name,
					Description: p.Description,
					IsPersonal:  true,
					CreatedAt:   time.Now(),
					CreatedBy:   testUser.ID,
				}, nil
			})

		mockQuerier.EXPECT().
			CreateTenantUser(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, p db.CreateTenantUserParams) (db.UserserviceTenantUser, error) {
				return db.UserserviceTenantUser{
					ID:        p.ID,
					TenantID:  p.TenantID,
					UserID:    p.UserID,
					Role:      constants.TenantRoleSuperAdmin,
					CreatedAt: time.Now(),
				}, nil
			})

		err := tenantAPI.CreatePersonalTenant(ctx)
		assert.NoError(t, err)
	})

	t.Run("fail - missing auth context", func(t *testing.T) {
		err := tenantAPI.CreatePersonalTenant(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get auth context")
	})

	t.Run("fail - CreateTenant returns error", func(t *testing.T) {
		mockQuerier.EXPECT().
			CreateTenant(gomock.Any(), gomock.Any()).
			Return(db.UserserviceTenant{}, errors.New("db insert failed"))

		err := tenantAPI.CreatePersonalTenant(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create personal tenant")
	})

	t.Run("fail - CreateTenantUser returns error", func(t *testing.T) {
		mockQuerier.EXPECT().
			CreateTenant(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, p db.CreateTenantParams) (db.UserserviceTenant, error) {
				return db.UserserviceTenant{
					ID:          uuid.New().String(),
					Name:        p.Name,
					Description: p.Description,
					IsPersonal:  true,
					CreatedAt:   time.Now(),
					CreatedBy:   testUser.ID,
				}, nil
			})

		mockQuerier.EXPECT().
			CreateTenantUser(gomock.Any(), gomock.Any()).
			Return(db.UserserviceTenantUser{}, errors.New("insert failed"))

		err := tenantAPI.CreatePersonalTenant(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to add creator to personal tenant")
	})
}

func TestCreateTenant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mocks.NewMockQuerier(ctrl)
	logger := slog.Default()
	tenantAPI := api.NewTenantAPITest(mockQuerier, logger)

	testUser := &auth.User{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Tester",
	}
	ctx := withAuthContext(context.Background(), testUser)

	t.Run("success - tenant created", func(t *testing.T) {
		req := &proto.CreateTenantRequest{Name: "My Team", Description: "Team workspace"}

		mockQuerier.EXPECT().
			CreateTenant(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, p db.CreateTenantParams) (db.UserserviceTenant, error) {
				return db.UserserviceTenant{
					ID:          p.ID,
					Name:        p.Name,
					Description: p.Description,
					IsPersonal:  false,
					CreatedAt:   time.Now(),
					CreatedBy:   testUser.ID,
				}, nil
			})

		mockQuerier.EXPECT().
			CreateTenantUser(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, p db.CreateTenantUserParams) (db.UserserviceTenantUser, error) {
				return db.UserserviceTenantUser{
					ID:        p.ID,
					TenantID:  p.TenantID,
					UserID:    testUser.ID,
					Role:      constants.TenantRoleSuperAdmin,
					CreatedAt: time.Now(),
				}, nil
			})

		resp, err := tenantAPI.CreateTenant(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, "Tenant created successfully", resp.Message)
		assert.Equal(t, req.Name, resp.TenantUser.Tenant.Name)
		assert.Equal(t, constants.TenantRoleSuperAdmin, resp.TenantUser.Role.Role)
	})

	t.Run("fail - unauthenticated", func(t *testing.T) {
		req := &proto.CreateTenantRequest{Name: "Should Fail"}
		resp, err := tenantAPI.CreateTenant(context.Background(), req)
		assert.Nil(t, resp)
		assert.Error(t, err)

		st, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("fail - tenant name empty", func(t *testing.T) {
		req := &proto.CreateTenantRequest{Name: ""}
		resp, err := tenantAPI.CreateTenant(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)

		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("fail - CreateTenant db error", func(t *testing.T) {
		req := &proto.CreateTenantRequest{Name: "ErrorTenant"}

		mockQuerier.EXPECT().
			CreateTenant(gomock.Any(), gomock.Any()).
			Return(db.UserserviceTenant{}, errors.New("db insert failed"))

		resp, err := tenantAPI.CreateTenant(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)

		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "failed to create tenant", st.Message())
	})

	t.Run("fail - CreateTenantUser db error", func(t *testing.T) {
		req := &proto.CreateTenantRequest{Name: "BadTenant"}

		mockQuerier.EXPECT().
			CreateTenant(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, p db.CreateTenantParams) (db.UserserviceTenant, error) {
				return db.UserserviceTenant{
					ID:          uuid.New().String(),
					Name:        p.Name,
					Description: sql.NullString{String: p.Description.String, Valid: true},
					IsPersonal:  false,
					CreatedAt:   time.Now(),
					CreatedBy:   testUser.ID,
				}, nil
			})

		mockQuerier.EXPECT().
			CreateTenantUser(gomock.Any(), gomock.Any()).
			Return(db.UserserviceTenantUser{}, errors.New("insert failed"))

		resp, err := tenantAPI.CreateTenant(ctx, req)
		assert.Nil(t, resp)
		assert.Error(t, err)

		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "failed to add creator to tenant", st.Message())
	})
}

func TestGetTenants(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mocks.NewMockQuerier(ctrl)
	logger := slog.Default()
	userAPI := api.NewUserAPITest(mockQuerier, nil, nil, logger)

	testUser := &auth.User{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Tester",
	}
	ctx := withAuthContext(context.Background(), testUser)

	t.Run("success - tenants returned", func(t *testing.T) {
		tenantRows := []db.GetUserTenantsRow{
			{
				TenantID:    "tenant-1",
				Name:        "Team Alpha",
				Description: sql.NullString{String: "First tenant", Valid: true},
				IsPersonal:  false,
				CreatedAt:   time.Now(),
				CreatedBy:   testUser.ID,
				Role:        "admin",
			},
			{
				TenantID:    "tenant-2",
				Name:        "Personal Space",
				Description: sql.NullString{String: "Personal workspace", Valid: true},
				IsPersonal:  true,
				CreatedAt:   time.Now(),
				CreatedBy:   testUser.ID,
				Role:        "super_admin",
			},
		}

		mockQuerier.EXPECT().
			GetUserTenants(gomock.Any(), testUser.ID).
			Return(tenantRows, nil)

		resp, err := userAPI.GetTenants(ctx, &proto.GetTenantsRequest{})
		assert.NoError(t, err)
		assert.Equal(t, "User tenants retrieved successfully", resp.Message)
		assert.Len(t, resp.TenantUsers, 2)
		assert.Equal(t, "Team Alpha", resp.TenantUsers[0].Tenant.Name)
		assert.Equal(t, "admin", resp.TenantUsers[0].Role.Role)
		assert.Equal(t, "super_admin", resp.TenantUsers[1].Role.Role)
	})

	t.Run("fail - unauthenticated", func(t *testing.T) {
		resp, err := userAPI.GetTenants(context.Background(), &proto.GetTenantsRequest{})
		assert.Nil(t, resp)
		assert.Error(t, err)

		st, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("fail - db error", func(t *testing.T) {
		mockQuerier.EXPECT().
			GetUserTenants(gomock.Any(), testUser.ID).
			Return(nil, errors.New("db connection failed"))

		resp, err := userAPI.GetTenants(ctx, &proto.GetTenantsRequest{})
		assert.Nil(t, resp)
		assert.Error(t, err)

		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "failed to get user tenants", st.Message())
	})

	t.Run("success - no tenants", func(t *testing.T) {
		mockQuerier.EXPECT().
			GetUserTenants(gomock.Any(), testUser.ID).
			Return([]db.GetUserTenantsRow{}, nil)

		resp, err := userAPI.GetTenants(ctx, &proto.GetTenantsRequest{})
		assert.NoError(t, err)
		assert.Equal(t, "User tenants retrieved successfully", resp.Message)
		assert.Empty(t, resp.TenantUsers)
	})
}

func TestAddUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mocks.NewMockQuerier(ctrl)
	logger := slog.Default()
	tenantAPI := api.NewTenantAPITest(mockQuerier, logger)

	testUser := &auth.User{
		ID:    "caller-123",
		Email: "caller@example.com",
		Name:  "Caller",
	}
	ctx := withAuthContext(context.Background(), testUser)

	targetUser := db.UserserviceUser{
		ID:       "target-456",
		Email:    "target@example.com",
		Username: "target@example.com",
	}

	t.Run("success - super_admin adds user", func(t *testing.T) {
		req := &proto.AddUserRequest{TenantId: "tenant-1", Username: "target@example.com"}

		mockQuerier.EXPECT().
			GetUserByEmail(gomock.Any(), "target@example.com").
			Return(targetUser, nil)

		mockQuerier.EXPECT().
			GetUserRoleInTenant(gomock.Any(), db.GetUserRoleInTenantParams{
				TenantID: "tenant-1",
				UserID:   testUser.ID,
			}).
			Return(constants.TenantRoleSuperAdmin, nil)

		mockQuerier.EXPECT().
			CreateTenantUser(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, p db.CreateTenantUserParams) (db.UserserviceTenantUser, error) {
				return db.UserserviceTenantUser{
					ID:        uuid.New().String(),
					TenantID:  p.TenantID,
					UserID:    p.UserID,
					Role:      p.Role,
					CreatedAt: time.Now(),
				}, nil
			})

		resp, err := tenantAPI.AddUser(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, "User added to tenant successfully", resp.Message)
	})

	t.Run("fail - unauthenticated", func(t *testing.T) {
		req := &proto.AddUserRequest{TenantId: "tenant-1", Username: "target@example.com"}
		resp, err := tenantAPI.AddUser(context.Background(), req)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("fail - invalid input", func(t *testing.T) {
		req := &proto.AddUserRequest{TenantId: "", Username: ""}
		resp, err := tenantAPI.AddUser(ctx, req)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("fail - user not found", func(t *testing.T) {
		req := &proto.AddUserRequest{TenantId: "tenant-1", Username: "notfound@example.com"}
		mockQuerier.EXPECT().
			GetUserByEmail(gomock.Any(), "notfound@example.com").
			Return(db.UserserviceUser{}, sql.ErrNoRows)

		resp, err := tenantAPI.AddUser(ctx, req)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("fail - db error checking user existence", func(t *testing.T) {
		req := &proto.AddUserRequest{TenantId: "tenant-1", Username: "target@example.com"}
		mockQuerier.EXPECT().
			GetUserByEmail(gomock.Any(), "target@example.com").
			Return(db.UserserviceUser{}, errors.New("db error"))

		resp, err := tenantAPI.AddUser(ctx, req)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
	})

	t.Run("fail - caller not in tenant", func(t *testing.T) {
		req := &proto.AddUserRequest{TenantId: "tenant-1", Username: "target@example.com"}
		mockQuerier.EXPECT().
			GetUserByEmail(gomock.Any(), "target@example.com").
			Return(targetUser, nil)
		mockQuerier.EXPECT().
			GetUserRoleInTenant(gomock.Any(), gomock.Any()).
			Return("", sql.ErrNoRows)

		resp, err := tenantAPI.AddUser(ctx, req)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})

	t.Run("fail - db error checking role", func(t *testing.T) {
		req := &proto.AddUserRequest{TenantId: "tenant-1", Username: "target@example.com"}
		mockQuerier.EXPECT().
			GetUserByEmail(gomock.Any(), "target@example.com").
			Return(targetUser, nil)
		mockQuerier.EXPECT().
			GetUserRoleInTenant(gomock.Any(), gomock.Any()).
			Return("", errors.New("db role error"))

		resp, err := tenantAPI.AddUser(ctx, req)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
	})

	t.Run("fail - caller is not super_admin", func(t *testing.T) {
		req := &proto.AddUserRequest{TenantId: "tenant-1", Username: "target@example.com"}
		mockQuerier.EXPECT().
			GetUserByEmail(gomock.Any(), "target@example.com").
			Return(targetUser, nil)
		mockQuerier.EXPECT().
			GetUserRoleInTenant(gomock.Any(), gomock.Any()).
			Return(constants.TenantRoleMember, nil)

		resp, err := tenantAPI.AddUser(ctx, req)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})

	t.Run("fail - db error creating tenant user", func(t *testing.T) {
		req := &proto.AddUserRequest{TenantId: "tenant-1", Username: "target@example.com"}
		mockQuerier.EXPECT().
			GetUserByEmail(gomock.Any(), "target@example.com").
			Return(targetUser, nil)
		mockQuerier.EXPECT().
			GetUserRoleInTenant(gomock.Any(), gomock.Any()).
			Return(constants.TenantRoleSuperAdmin, nil)
		mockQuerier.EXPECT().
			CreateTenantUser(gomock.Any(), gomock.Any()).
			Return(db.UserserviceTenantUser{}, errors.New("insert failed"))

		resp, err := tenantAPI.AddUser(ctx, req)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
	})
}

func TestGetUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	logger := slog.Default()
	tenantAPI := api.NewTenantAPITest(mockDB, logger)

	testUser := &auth.User{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Tester",
	}

	withAuth := func(u *auth.User) context.Context {
		return withAuthContext(context.Background(), u)
	}

	t.Run("success - super admin retrieves tenant users", func(t *testing.T) {
		ctx := withAuth(testUser)

		// mock: user is super_admin
		mockDB.EXPECT().
			GetUserRoleInTenant(gomock.Any(), db.GetUserRoleInTenantParams{
				TenantID: "tenant-1",
				UserID:   testUser.ID,
			}).
			Return(constants.TenantRoleSuperAdmin, nil)

		// mock: tenant users
		mockDB.EXPECT().
			GetTenantUsers(gomock.Any(), "tenant-1").
			Return([]db.GetTenantUsersRow{
				{
					TenantName:      "Tenant One",
					TenantCreatedAt: time.Now(),
					UserID:          "u-1",
					Username:        "alice",
					Email:           "alice@example.com",
					Role:            constants.TenantRoleMember,
				},
				{
					TenantName:      "Tenant One",
					TenantCreatedAt: time.Now(),
					UserID:          "u-2",
					Username:        "bob",
					Email:           "bob@example.com",
					Role:            constants.TenantRoleSuperAdmin,
				},
			}, nil)

		resp, err := tenantAPI.GetUsers(ctx, &proto.GetUsersRequest{TenantId: "tenant-1"})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Tenant users retrieved successfully", resp.Message)
		assert.Len(t, resp.TenantUsers, 2)
	})

	t.Run("fail - missing auth context", func(t *testing.T) {
		_, err := tenantAPI.GetUsers(context.Background(), &proto.GetUsersRequest{TenantId: "tenant-1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthenticated")
	})

	t.Run("fail - missing tenant ID", func(t *testing.T) {
		ctx := withAuth(testUser)
		_, err := tenantAPI.GetUsers(ctx, &proto.GetUsersRequest{TenantId: ""})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tenant ID is required")
	})

	t.Run("fail - user not in tenant (no rows)", func(t *testing.T) {
		ctx := withAuth(testUser)

		mockDB.EXPECT().
			GetUserRoleInTenant(gomock.Any(), db.GetUserRoleInTenantParams{
				TenantID: "tenant-1",
				UserID:   testUser.ID,
			}).
			Return("", sql.ErrNoRows)

		_, err := tenantAPI.GetUsers(ctx, &proto.GetUsersRequest{TenantId: "tenant-1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("fail - db error on role lookup", func(t *testing.T) {
		ctx := withAuth(testUser)

		mockDB.EXPECT().
			GetUserRoleInTenant(gomock.Any(), db.GetUserRoleInTenantParams{
				TenantID: "tenant-1",
				UserID:   testUser.ID,
			}).
			Return("", errors.New("db connection lost"))

		_, err := tenantAPI.GetUsers(ctx, &proto.GetUsersRequest{TenantId: "tenant-1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check permissions")
	})

	t.Run("fail - user is not super admin", func(t *testing.T) {
		ctx := withAuth(testUser)

		mockDB.EXPECT().
			GetUserRoleInTenant(gomock.Any(), db.GetUserRoleInTenantParams{
				TenantID: "tenant-1",
				UserID:   testUser.ID,
			}).
			Return(constants.TenantRoleMember, nil)

		_, err := tenantAPI.GetUsers(ctx, &proto.GetUsersRequest{TenantId: "tenant-1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only super admins can view tenant members")
	})

	t.Run("fail - db error on GetTenantUsers", func(t *testing.T) {
		ctx := withAuth(testUser)

		mockDB.EXPECT().
			GetUserRoleInTenant(gomock.Any(), db.GetUserRoleInTenantParams{
				TenantID: "tenant-1",
				UserID:   testUser.ID,
			}).
			Return(constants.TenantRoleSuperAdmin, nil)

		mockDB.EXPECT().
			GetTenantUsers(gomock.Any(), "tenant-1").
			Return(nil, errors.New("query failed"))

		_, err := tenantAPI.GetUsers(ctx, &proto.GetUsersRequest{TenantId: "tenant-1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get tenant users")
	})
}
