package api

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sortedstartup.com/stream/common/constants"
	"sortedstartup.com/stream/common/interceptors"
	paymentProto "sortedstartup.com/stream/paymentservice/proto"
	"sortedstartup.com/stream/userservice/config"
	"sortedstartup.com/stream/userservice/db"
	"sortedstartup.com/stream/userservice/proto"
)

type UserAPI struct {
	config               config.UserServiceConfig
	db                   *sql.DB
	log                  *slog.Logger
	dbQueries            *db.Queries
	userCache            *lru.Cache
	paymentServiceClient paymentProto.PaymentServiceClient
	proto.UnimplementedUserServiceServer
	tenantAPI *TenantAPI
}

type TenantAPI struct {
	config               config.UserServiceConfig
	db                   *sql.DB
	log                  *slog.Logger
	dbQueries            *db.Queries
	paymentServiceClient paymentProto.PaymentServiceClient
	proto.UnimplementedTenantServiceServer
}

func NewUserAPI(config config.UserServiceConfig, paymentServiceClient paymentProto.PaymentServiceClient) (*UserAPI, *TenantAPI, error) {
	slog.Info("NewUserAPI")

	childLogger := slog.With("service", "UserAPI")

	_db, err := sql.Open(config.DB.Driver, config.DB.Url)
	if err != nil {
		return nil, nil, err
	}

	dbQueries := db.New(_db)

	cache, err := lru.New(config.CacheSize)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create user cache: %w", err)
	}

	tenantAPI := &TenantAPI{
		config:               config,
		db:                   _db,
		log:                  childLogger,
		dbQueries:            dbQueries,
		paymentServiceClient: paymentServiceClient,
	}

	userAPI := &UserAPI{
		config:               config,
		db:                   _db,
		log:                  childLogger,
		userCache:            cache,
		dbQueries:            dbQueries,
		paymentServiceClient: paymentServiceClient,
		tenantAPI:            tenantAPI,
	}

	return userAPI, tenantAPI, nil
}

func NewUserAPITest(querier db.Querier, cache *lru.Cache, tenantAPI *TenantAPI, logger *slog.Logger) *UserAPI {
	return &UserAPI{
		dbQueries: querier.(*db.Queries),
		userCache: cache,
		tenantAPI: tenantAPI,
		log:       logger,
	}
}

func NewTenantAPITest(querier db.Querier, logger *slog.Logger) *TenantAPI {
	return &TenantAPI{
		dbQueries: querier.(*db.Queries),
		log:       logger,
	}
}

func (s *UserAPI) Start() error {
	return nil
}

func (s *UserAPI) Init() error {
	s.log.Info("Migrating database", "dbDriver", s.config.DB.Driver, "dbURL", s.config.DB.Url)
	err := db.MigrateDB(s.config.DB.Driver, s.config.DB.Url)
	if err != nil {
		return err
	}

	// Run user count backfill for existing users (only if needed)
	s.log.Info("Checking for users that need user count backfill...")
	err = s.backfillUserCounts(context.Background())
	if err != nil {
		s.log.Error("User count backfill failed", "error", err)
		// Don't fail startup, just log the error
	} else {
		s.log.Info("User count backfill completed")
	}

	return nil
}

func (s *UserAPI) CreateUserIfNotExists(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	s.log.Info("CreateUserIfNotExists")
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userEmail := authContext.User.Email
	s.log.Info("userEmail", "userEmail", userEmail)

	// CACHE CHECK
	if _, found := s.userCache.Get(userEmail); found {
		s.log.Info("Cache hit: skipping DB check", "email", userEmail)
		return &proto.CreateUserResponse{
			Message: "User already exists (cache)",
		}, nil
	}

	s.log.Info("querying DB for email", "email", userEmail)

	// DB CHECK (and fallback to create)
	dbUser, err := s.dbQueries.GetUserByEmail(ctx, userEmail)
	s.log.Info("GetUserByEmail result", "error", err, "hasError", err != nil, "isNoRows", err == sql.ErrNoRows)

	var successMessage string

	if err != nil {
		if err == sql.ErrNoRows {
			s.log.Info("User doesn't exist, creating them")
			// User doesn't exist, create them
			createParams := db.CreateUserParams{
				ID:        authContext.User.ID,
				Username:  authContext.User.Email,
				Email:     authContext.User.Email,
				CreatedAt: time.Now(),
			}

			dbUser, err = s.dbQueries.CreateUser(ctx, createParams)
			if err != nil {
				s.log.Error("Failed to create user", "error", err)
				return nil, status.Error(codes.Internal, "failed to create user")
			}
			s.log.Info("User created successfully with email", "email", authContext.User.Email)
			successMessage = "User created successfully"
			s.userCache.Add(userEmail, true)
			s.log.Info("adding email to cache", "email", userEmail)

			// Initialize user subscription (handles both payment service and user limits)
			s.log.Info("Initializing user subscription for new user", "userID", authContext.User.ID)
			initResp, err := s.InitializeUserSubscription(ctx, &proto.InitializeUserSubscriptionRequest{
				UserId: authContext.User.ID,
			})
			if err != nil {
				s.log.Error("Failed to initialize user subscription", "error", err, "userID", authContext.User.ID)
				// Don't fail the entire request, just log the error
			} else if initResp != nil && !initResp.Success {
				s.log.Error("User subscription initialization failed", "error", initResp.ErrorMessage, "userID", authContext.User.ID)
			} else {
				s.log.Info("User subscription initialized successfully", "userID", authContext.User.ID)
			}

			err = s.tenantAPI.createPersonalTenant(ctx)
			if err != nil {
				s.log.Error("Failed to create personal tenant", "error", err)
				// Don't fail the entire request, just log the error
			}

		} else {
			s.log.Error("Database error while getting user", "error", err)
			return nil, status.Error(codes.Internal, "internal server error")
		}
	} else {
		s.log.Info("User already exists with email", "email", authContext.User.Email)
		// Check if the returned user is actually valid (not empty)
		if dbUser.ID == "" || dbUser.Email == "" {
			s.log.Warn("User exists but has empty fields, this might indicate a database issue", "dbUser", dbUser)
		} else {
			// ADD TO CACHE if user haven't logged in recently (Cache miss) to cache hit in future login
			s.userCache.Add(userEmail, true)
			s.log.Info("adding email to cache", "email", userEmail)

			// Check if existing user has payment service initialized (migration for old users)
			s.log.Info("Checking payment service for existing user", "userID", authContext.User.ID)
			paymentResp, err := s.paymentServiceClient.GetUserSubscription(ctx, &paymentProto.GetUserSubscriptionRequest{
				UserId: authContext.User.ID,
			})
			if err != nil || (paymentResp != nil && !paymentResp.Success) {
				s.log.Warn("Existing user has no subscription record, initializing", "userID", authContext.User.ID)
				// Initialize user subscription for existing user (migration)
				initResp, err := s.InitializeUserSubscription(ctx, &proto.InitializeUserSubscriptionRequest{
					UserId: authContext.User.ID,
				})
				if err != nil {
					s.log.Error("Failed to initialize user subscription for existing user", "error", err, "userID", authContext.User.ID)
				} else if initResp != nil && !initResp.Success {
					s.log.Error("User subscription initialization failed for existing user", "error", initResp.ErrorMessage, "userID", authContext.User.ID)
				} else {
					s.log.Info("User subscription initialized successfully for existing user", "userID", authContext.User.ID)
				}
			}
		}
		successMessage = "User already exists"
	}

	// Return success response with message
	return &proto.CreateUserResponse{
		Message: successMessage,
		User: &proto.User{
			Id:        dbUser.ID,
			Username:  dbUser.Username,
			Email:     dbUser.Email,
			CreatedAt: timestamppb.New(dbUser.CreatedAt),
		},
	}, nil
}

/**
* createPersonalTenant creates a personal tenant for a new user
* @param ctx context.Context
* @return error
 */
func (s *TenantAPI) createPersonalTenant(ctx context.Context) error {

	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get auth context: %w", err)
	}
	userName := authContext.User.Name

	// Create personal tenant directly
	tenantID := uuid.New().String()
	tenantParams := db.CreateTenantParams{
		ID:          tenantID,
		Name:        userName,
		Description: sql.NullString{String: "Personal workspace", Valid: true},
		IsPersonal:  true,
		CreatedAt:   time.Now(),
		CreatedBy:   authContext.User.ID,
	}

	tenant, err := s.dbQueries.CreateTenant(ctx, tenantParams)
	if err != nil {
		s.log.Error("Failed to create personal tenant", "error", err)
		return fmt.Errorf("failed to create personal tenant: %w", err)
	}

	// Add creator to personal tenant as super_admin
	tenantUserParams := db.CreateTenantUserParams{
		ID:        uuid.New().String(),
		TenantID:  tenant.ID,
		UserID:    authContext.User.ID,
		Role:      constants.TenantRoleSuperAdmin,
		CreatedAt: time.Now(),
	}

	_, err = s.dbQueries.CreateTenantUser(ctx, tenantUserParams)
	if err != nil {
		s.log.Error("Failed to add creator to personal tenant", "error", err)
		return fmt.Errorf("failed to add creator to personal tenant: %w", err)
	}
	s.log.Info("Personal tenant created successfully", "tenantID", tenantID, "userName", userName)
	return nil
}

/**
* CreateTenant creates a new organizational tenant
* Add creator to tenant as super_admin
* @param ctx context.Context
* @param req *proto.CreateTenantRequest
* @return *proto.CreateTenantResponse, error
 */
func (s *TenantAPI) CreateTenant(ctx context.Context, req *proto.CreateTenantRequest) (*proto.CreateTenantResponse, error) {
	s.log.Info("CreateTenant", "name", req.Name)

	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant name is required")
	}

	// Check if tenant name already exists for this user
	_, err = s.dbQueries.GetTenantByName(ctx, db.GetTenantByNameParams{
		Name:      req.Name,
		CreatedBy: authContext.User.ID,
	})
	if err == nil {
		// Tenant with this name already exists
		return nil, status.Error(codes.AlreadyExists, "A workspace with this name already exists")
	} else if err != sql.ErrNoRows {
		// Some other database error occurred
		s.log.Error("Failed to check for existing tenant name", "error", err)
		return nil, status.Error(codes.Internal, "failed to validate tenant name")
	}

	tenantID := uuid.New().String()
	tenantParams := db.CreateTenantParams{
		ID:          tenantID,
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		IsPersonal:  false,
		CreatedAt:   time.Now(),
		CreatedBy:   authContext.User.ID,
	}

	tenant, err := s.dbQueries.CreateTenant(ctx, tenantParams)
	if err != nil {
		s.log.Error("Failed to create tenant", "error", err)
		return nil, status.Error(codes.Internal, "failed to create tenant")
	}

	// Add creator to tenant as super_admin
	tenantUserParams := db.CreateTenantUserParams{
		ID:        uuid.New().String(),
		TenantID:  tenant.ID,
		UserID:    authContext.User.ID,
		Role:      constants.TenantRoleSuperAdmin,
		CreatedAt: time.Now(),
	}

	tenantUser, err := s.dbQueries.CreateTenantUser(ctx, tenantUserParams)
	if err != nil {
		s.log.Error("Failed to add creator to tenant", "error", err)
		return nil, status.Error(codes.Internal, "failed to add creator to tenant")
	}

	protoTenant := &proto.Tenant{
		Id:          tenant.ID,
		Name:        tenant.Name,
		Description: tenant.Description.String,
		IsPersonal:  tenant.IsPersonal,
		CreatedAt:   timestamppb.New(tenant.CreatedAt),
		CreatedBy:   tenant.CreatedBy,
	}

	return &proto.CreateTenantResponse{
		Message: "Tenant created successfully",
		TenantUser: &proto.TenantUser{
			Tenant: protoTenant,
			Role:   &proto.Role{Role: tenantUser.Role},
		},
	}, nil
}

/**
* GetUserTenants returns all tenants a user belongs to
* @param ctx context.Context
* @param req *proto.GetUserTenantsRequest
* @return *proto.GetUserTenantsResponse, error
 */
func (s *UserAPI) GetTenants(ctx context.Context, req *proto.GetTenantsRequest) (*proto.GetTenantsResponse, error) {

	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userID := authContext.User.ID
	s.log.Info("GetUserTenants", "userID", userID)

	tenantRows, err := s.dbQueries.GetUserTenants(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get user tenants", "error", err)
		return nil, status.Error(codes.Internal, "failed to get user tenants")
	}

	var tenants []*proto.TenantUser
	for _, row := range tenantRows {
		tenant := &proto.TenantUser{
			Tenant: &proto.Tenant{
				Id:          row.TenantID,
				Name:        row.Name,
				Description: row.Description.String,
				IsPersonal:  row.IsPersonal,
				CreatedAt:   timestamppb.New(row.CreatedAt),
				CreatedBy:   row.CreatedBy,
			},
			Role: &proto.Role{
				Role: row.Role,
			},
		}
		tenants = append(tenants, tenant)
	}

	return &proto.GetTenantsResponse{
		Message:     "User tenants retrieved successfully",
		TenantUsers: tenants,
	}, nil
}

/**
* AddUser adds a user to an existing tenant using username (email)
* @param ctx context.Context
* @param req *proto.AddUserRequest
* @return *proto.AddUserResponse, error
 */
func (s *TenantAPI) AddUser(ctx context.Context, req *proto.AddUserRequest) (*proto.AddUserResponse, error) {
	s.log.Info("AddUser", "tenantID", req.TenantId, "username", req.Username)

	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	// Validate input
	if req.TenantId == "" || req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant ID and username are required")
	}

	// Validate that the user exists and get their ID
	user, err := s.dbQueries.GetUserByEmail(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			s.log.Warn("Attempted to add non-existent user to tenant", "username", req.Username, "tenantID", req.TenantId)
			return nil, status.Error(codes.NotFound, "user not found")
		}
		s.log.Error("Failed to check if user exists", "error", err, "username", req.Username)
		return nil, status.Error(codes.Internal, "failed to validate user")
	}

	// Default role to member if not specified
	role := req.Role
	if role == "" {
		role = constants.TenantRoleMember
	}

	// Authorization check - only super_admin can add users to tenant
	userRole, err := s.dbQueries.GetUserRoleInTenant(ctx, db.GetUserRoleInTenantParams{
		TenantID: req.TenantId,
		UserID:   authContext.User.ID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			s.log.Warn("User attempted to add user to tenant they don't belong to", "userID", authContext.User.ID, "tenantID", req.TenantId)
			return nil, status.Error(codes.PermissionDenied, "access denied: you are not a member of this tenant")
		}
		s.log.Error("Failed to check user role in tenant", "error", err)
		return nil, status.Error(codes.Internal, "failed to check permissions")
	}

	// Only super_admin can add users to tenant
	if userRole != constants.TenantRoleSuperAdmin {
		s.log.Warn("Non-super-admin user attempted to add user to tenant", "userID", authContext.User.ID, "role", userRole, "tenantID", req.TenantId)
		return nil, status.Error(codes.PermissionDenied, "access denied: only super admins can add users to tenant")
	}

	// Get tenant details
	tenant, err := s.dbQueries.GetTenantByID(ctx, req.TenantId)
	if err != nil {
		s.log.Error("Failed to get tenant details", "error", err, "tenantID", req.TenantId)
		return nil, status.Error(codes.Internal, "failed to get tenant details")
	}

	// Check workspace type and apply appropriate logic
	if tenant.IsPersonal {
		s.log.Info("Adding user to personal workspace", "tenantOwner", tenant.CreatedBy, "tenantID", req.TenantId)
	} else {
		s.log.Info("Adding user to organizational workspace", "tenantOwner", tenant.CreatedBy, "tenantID", req.TenantId)
	}

	tenantUserParams := db.CreateTenantUserParams{
		ID:        uuid.New().String(),
		TenantID:  req.TenantId,
		UserID:    user.ID, // Use the user ID from the database lookup
		Role:      role,
		CreatedAt: time.Now(),
	}

	_, err = s.dbQueries.CreateTenantUser(ctx, tenantUserParams)
	if err != nil {
		s.log.Error("Failed to add user to tenant", "error", err)
		return nil, status.Error(codes.Internal, "failed to add user to tenant")
	}

	// Update payment service user count only if this user is NOT already in any tenant owned by the tenant owner
	shouldUpdateCount, err := s.shouldUpdateUserCount(ctx, tenant.CreatedBy, user.ID)
	if err != nil {
		s.log.Error("Failed to check if user count should be updated",
			"tenantOwner", tenant.CreatedBy,
			"newUserID", user.ID,
			"error", err)
		// Continue without updating count to avoid double counting
	} else if shouldUpdateCount {
		// Update user count in userservice_user_limits table
		s.log.Info("Updating user count for tenant owner", "tenantOwner", tenant.CreatedBy, "tenantID", req.TenantId, "newUserID", user.ID)

		// Get current user count
		currentUsage, err := s.dbQueries.GetUserLimits(ctx, tenant.CreatedBy)
		currentCount := int64(1) // Default to 1 (owner themselves)
		if err == nil {
			currentCount = currentUsage.UsersCount.Int64
		}

		// Increment by 1 for the new user
		newCount := currentCount + 1
		now := time.Now().Unix()

		err = s.dbQueries.UpsertUserLimits(ctx, db.UpsertUserLimitsParams{
			UserID:           tenant.CreatedBy,
			UsersCount:       sql.NullInt64{Int64: newCount, Valid: true},
			LastCalculatedAt: sql.NullInt64{Int64: now, Valid: true},
			CreatedAt:        now,
			UpdatedAt:        now,
		})
		if err != nil {
			s.log.Error("Failed to update user count", "error", err, "tenantOwner", tenant.CreatedBy, "newCount", newCount)
			// Don't fail the request, just log the error
		} else {
			s.log.Info("User count updated successfully", "tenantOwner", tenant.CreatedBy, "oldCount", currentCount, "newCount", newCount)
		}
	} else {
		s.log.Info("User already counted for this owner, skipping user count update",
			"tenantOwner", tenant.CreatedBy,
			"existingUserID", user.ID)
	}

	return &proto.AddUserResponse{
		Message: "User added to tenant successfully",
	}, nil
}

// shouldUpdateUserCount checks if a user is NOT already counted in any tenant owned by the tenant owner
func (s *TenantAPI) shouldUpdateUserCount(ctx context.Context, tenantOwnerID, newUserID string) (bool, error) {
	// Check if the newUserID is already a member of any tenant owned by tenantOwnerID
	// If they are, we shouldn't increment the count (avoid double counting)

	query := `
		SELECT COUNT(*) 
		FROM userservice_tenants t
		JOIN userservice_tenant_users tu ON t.id = tu.tenant_id
		WHERE t.created_by = ? AND tu.user_id = ?
	`

	var existingCount int
	err := s.db.QueryRowContext(ctx, query, tenantOwnerID, newUserID).Scan(&existingCount)
	if err != nil {
		return false, fmt.Errorf("failed to check existing user count: %w", err)
	}

	// If existingCount > 1, it means the user was already in another tenant owned by this owner
	// (We expect exactly 1 because we just added them to the current tenant)
	shouldUpdate := existingCount <= 1

	s.log.Info("User count check result",
		"tenantOwner", tenantOwnerID,
		"newUserID", newUserID,
		"existingCount", existingCount,
		"shouldUpdate", shouldUpdate)

	return shouldUpdate, nil
}

/**
* GetTenantUsers returns all users in a tenant - restricted to super_admin only
* @param ctx context.Context
* @param req *proto.GetTenantUsersRequest
* @return *proto.GetTenantUsersResponse, error
 */
func (s *TenantAPI) GetUsers(ctx context.Context, req *proto.GetUsersRequest) (*proto.GetUsersResponse, error) {

	s.log.Info("GetUsers", "tenantID", req.TenantId)

	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	// Validate input
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	// Authorization check - only super_admin can view tenant users
	userRole, err := s.dbQueries.GetUserRoleInTenant(ctx, db.GetUserRoleInTenantParams{
		TenantID: req.TenantId,
		UserID:   authContext.User.ID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			s.log.Warn("User attempted to access tenant they don't belong to", "userID", authContext.User.ID, "tenantID", req.TenantId)
			return nil, status.Error(codes.PermissionDenied, "access denied: you are not a member of this tenant")
		}
		s.log.Error("Failed to check user role in tenant", "error", err)
		return nil, status.Error(codes.Internal, "failed to check permissions")
	}

	// Only super_admin can view tenant users
	if userRole != constants.TenantRoleSuperAdmin {
		s.log.Warn("Non-super-admin user attempted to view tenant users", "userID", authContext.User.ID, "role", userRole, "tenantID", req.TenantId)
		return nil, status.Error(codes.PermissionDenied, "access denied: only super admins can view tenant members")
	}

	tenantUsers, err := s.dbQueries.GetTenantUsers(ctx, req.TenantId)
	if err != nil {
		s.log.Error("Failed to get tenant users", "error", err)
		return nil, status.Error(codes.Internal, "failed to get tenant users")
	}

	var tenantUsersProto []*proto.TenantUser

	for _, user := range tenantUsers {
		tenantUsersProto = append(tenantUsersProto, &proto.TenantUser{
			Tenant: &proto.Tenant{
				Name:      user.TenantName,
				CreatedAt: timestamppb.New(user.TenantCreatedAt),
			},
			User: &proto.User{
				Id:       user.UserID,
				Username: user.Username,
				Email:    user.Email,
			},
			Role: &proto.Role{
				Role: user.Role,
			},
		})
	}

	return &proto.GetUsersResponse{
		Message:     "Tenant users retrieved successfully",
		TenantUsers: tenantUsersProto,
	}, nil
}

// InitializeUserSubscription creates a free subscription for new users (moved from PaymentService)
func (s *UserAPI) InitializeUserSubscription(ctx context.Context, req *proto.InitializeUserSubscriptionRequest) (*proto.InitializeUserSubscriptionResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	s.log.Info("InitializeUserSubscription", "userID", req.UserId)

	// 1. Check if user already has subscription
	subscription, err := s.paymentServiceClient.GetUserSubscription(ctx, &paymentProto.GetUserSubscriptionRequest{
		UserId: req.UserId,
	})
	if err == nil && subscription.Success {
		// User already has subscription
		s.log.Info("User already has subscription", "userID", req.UserId)
		return &proto.InitializeUserSubscriptionResponse{
			Success: true,
		}, nil
	}

	// 2. Create free subscription via PaymentService
	_, err = s.paymentServiceClient.CreateUserSubscription(ctx, &paymentProto.CreateUserSubscriptionRequest{
		UserId: req.UserId,
		PlanId: "free",
	})
	if err != nil {
		s.log.Error("Failed to create user subscription", "error", err, "userID", req.UserId)
		return &proto.InitializeUserSubscriptionResponse{
			Success:      false,
			ErrorMessage: "Failed to create free subscription",
		}, nil
	}

	// 3. Initialize user count to 1 (themselves) in UserService database
	now := time.Now().Unix()
	queries := s.dbQueries
	err = queries.UpsertUserLimits(ctx, db.UpsertUserLimitsParams{
		UserID:           req.UserId,
		UsersCount:       sql.NullInt64{Int64: 1, Valid: true}, // Start with 1 user (themselves)
		LastCalculatedAt: sql.NullInt64{Int64: now, Valid: true},
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		s.log.Error("Failed to initialize user count", "error", err, "userID", req.UserId)
		// Don't fail - subscription was created successfully
	}

	s.log.Info("User subscription initialized successfully", "userID", req.UserId)
	return &proto.InitializeUserSubscriptionResponse{
		Success: true,
	}, nil
}

// CheckUserAccess checks if user can add more users to their tenant (moved from PaymentService)
func (s *UserAPI) CheckUserAccess(ctx context.Context, req *proto.CheckUserAccessRequest) (*proto.CheckUserAccessResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	s.log.Info("CheckUserAccess", "userID", req.UserId, "requestedUserCount", req.RequestedUserCount)

	// 1. Get user's subscription from PaymentService
	subscription, err := s.paymentServiceClient.GetUserSubscription(ctx, &paymentProto.GetUserSubscriptionRequest{
		UserId: req.UserId,
	})
	if err != nil {
		s.log.Error("Failed to get user subscription", "error", err, "userID", req.UserId)
		return nil, status.Error(codes.Internal, "failed to get user subscription")
	}

	if !subscription.Success {
		s.log.Warn("User has no subscription", "userID", req.UserId)
		return &proto.CheckUserAccessResponse{
			HasAccess:      false,
			Reason:         "no_subscription",
			IsNearLimit:    false,
			WarningMessage: "No subscription found. Please upgrade to continue.",
		}, nil
	}

	// 2. Get plan limits from application-specific config
	planID := subscription.SubscriptionInfo.Plan.Id
	appPlanLimit := s.config.GetPlanLimitByID(planID)
	if appPlanLimit == nil {
		s.log.Error("Unknown plan", "planID", planID)
		return nil, status.Error(codes.Internal, "unknown plan")
	}

	// 3. Get current user count from UserService database
	queries := s.dbQueries
	usage, err := queries.GetUserLimits(ctx, req.UserId)
	if err != nil && err != sql.ErrNoRows {
		s.log.Error("Failed to get user limits", "error", err, "userID", req.UserId)
		return nil, status.Error(codes.Internal, "failed to get user limits")
	}

	currentUsers := int64(0)
	if err == nil {
		currentUsers = usage.UsersCount.Int64
	}

	// 4. Check subscription status
	if subscription.SubscriptionInfo.Subscription.Status != "active" {
		return &proto.CheckUserAccessResponse{
			HasAccess: false,
			Reason:    "subscription_inactive",
			UserInfo: &proto.UserLimitInfo{
				CurrentUsers: currentUsers,
				LimitUsers:   appPlanLimit.UsersLimit,
				UsagePercent: float64(currentUsers) / float64(appPlanLimit.UsersLimit) * 100,
				PlanId:       planID,
			},
			IsNearLimit:    false,
			WarningMessage: "Subscription is inactive",
		}, nil
	}

	// 5. Check if request would exceed user limit
	wouldExceed := currentUsers+req.RequestedUserCount > appPlanLimit.UsersLimit
	usagePercent := float64(currentUsers) / float64(appPlanLimit.UsersLimit) * 100
	isNearLimit := usagePercent > 75

	hasAccess := !wouldExceed
	reason := ""
	warningMessage := ""

	if !hasAccess {
		reason = "user_limit_exceeded"
	}

	if isNearLimit && hasAccess {
		warningMessage = "User limit %.1f%% full. Consider upgrading your plan."
	}

	return &proto.CheckUserAccessResponse{
		HasAccess: hasAccess,
		Reason:    reason,
		UserInfo: &proto.UserLimitInfo{
			CurrentUsers: currentUsers,
			LimitUsers:   appPlanLimit.UsersLimit,
			UsagePercent: usagePercent,
			PlanId:       planID,
		},
		IsNearLimit:    isNearLimit,
		WarningMessage: warningMessage,
	}, nil
}

// UpdateUserUsage updates user count (moved from PaymentService)
func (s *UserAPI) UpdateUserUsage(ctx context.Context, req *proto.UpdateUserUsageRequest) (*proto.UpdateUserUsageResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	now := time.Now().Unix()

	// Use UpsertUserLimits to handle both create and update
	queries := s.dbQueries
	err := queries.UpsertUserLimits(ctx, db.UpsertUserLimitsParams{
		UserID:           req.UserId,
		UsersCount:       sql.NullInt64{Int64: req.UsageChange, Valid: true}, // For upsert, this is the new total
		LastCalculatedAt: sql.NullInt64{Int64: now, Valid: true},
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		s.log.Error("Failed to update user count", "error", err, "userID", req.UserId)
		return nil, status.Error(codes.Internal, "failed to update user count")
	}

	// Get updated usage to return
	usage, err := queries.GetUserLimits(ctx, req.UserId)
	if err != nil {
		s.log.Error("Failed to get updated user limits", "error", err, "userID", req.UserId)
		return &proto.UpdateUserUsageResponse{
			Success: true,
		}, nil
	}

	return &proto.UpdateUserUsageResponse{
		Success: true,
		UpdatedInfo: &proto.UserLimitInfo{
			CurrentUsers: usage.UsersCount.Int64,
			LimitUsers:   0, // Will be filled by caller if needed
			PlanId:       "",
		},
	}, nil
}

// GetUserUsage gets current user count limits
func (s *UserAPI) GetUserUsage(ctx context.Context, req *proto.GetUserUsageRequest) (*proto.GetUserUsageResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Get user's subscription to determine plan
	subscription, err := s.paymentServiceClient.GetUserSubscription(ctx, &paymentProto.GetUserSubscriptionRequest{
		UserId: req.UserId,
	})
	if err != nil || !subscription.Success {
		return nil, status.Error(codes.Internal, "failed to get user subscription")
	}

	// Get plan limits from UserService config
	planID := subscription.SubscriptionInfo.Plan.Id
	appPlanLimit := s.config.GetPlanLimitByID(planID)
	if appPlanLimit == nil {
		return nil, status.Error(codes.Internal, "unknown plan")
	}

	queries := s.dbQueries
	usage, err := queries.GetUserLimits(ctx, req.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			return &proto.GetUserUsageResponse{
				Success: true,
				UserInfo: &proto.UserLimitInfo{
					CurrentUsers: 0,
					LimitUsers:   appPlanLimit.UsersLimit,
					UsagePercent: 0,
					PlanId:       planID,
				},
			}, nil
		}
		return nil, status.Error(codes.Internal, "failed to get user limits")
	}

	// Calculate usage percentage
	usagePercent := float64(0)
	if appPlanLimit.UsersLimit > 0 {
		usagePercent = (float64(usage.UsersCount.Int64) / float64(appPlanLimit.UsersLimit)) * 100
	}

	return &proto.GetUserUsageResponse{
		Success: true,
		UserInfo: &proto.UserLimitInfo{
			CurrentUsers: usage.UsersCount.Int64,
			LimitUsers:   appPlanLimit.UsersLimit,
			UsagePercent: usagePercent,
			PlanId:       planID,
		},
	}, nil
}

// GetPlanInfo gets plan information with application-specific limits
// This is a simplified version that only returns user limits from UserService config
// Storage limits would need to be added via VideoService integration
func (s *UserAPI) GetPlanInfo(ctx context.Context, req *proto.GetPlanInfoRequest) (*proto.GetPlanInfoResponse, error) {
	if req.PlanId == "" {
		return nil, status.Error(codes.InvalidArgument, "plan_id is required")
	}

	// Get basic plan info from PaymentService
	plansResponse, err := s.paymentServiceClient.GetPlans(ctx, &paymentProto.GetPlansRequest{})
	if err != nil || !plansResponse.Success {
		return nil, status.Error(codes.Internal, "failed to get plans from payment service")
	}

	// Find the requested plan
	var paymentPlan *paymentProto.Plan
	for _, plan := range plansResponse.Plans {
		if plan.Id == req.PlanId {
			paymentPlan = plan
			break
		}
	}

	if paymentPlan == nil {
		return nil, status.Error(codes.NotFound, "plan not found")
	}

	// Get plan limits from UserService config (now includes both user and storage limits)
	appPlanLimit := s.config.GetPlanLimitByID(req.PlanId)
	usersLimit := int64(0)
	storageLimitBytes := int64(0)
	if appPlanLimit != nil {
		usersLimit = appPlanLimit.UsersLimit
		storageLimitBytes = appPlanLimit.StorageLimitBytes()
	}

	return &proto.GetPlanInfoResponse{
		Success: true,
		PlanInfo: &proto.PlanInfo{
			Id:                paymentPlan.Id,
			Name:              paymentPlan.Name,
			PriceCents:        paymentPlan.PriceCents,
			IsActive:          paymentPlan.IsActive,
			StorageLimitBytes: storageLimitBytes,
			UsersLimit:        usersLimit,
		},
	}, nil
}
