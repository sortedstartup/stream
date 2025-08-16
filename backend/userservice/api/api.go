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
		dbQueries: querier,
		userCache: cache,
		tenantAPI: tenantAPI,
		log:       logger,
	}
}

func NewTenantAPITest(querier db.Querier, logger *slog.Logger) *TenantAPI {
	return &TenantAPI{
		dbQueries: querier,
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

			// Initialize payment service for new user
			s.log.Info("Initializing payment service for new user", "userID", authContext.User.ID)
			paymentResp, err := s.paymentServiceClient.InitializeUser(ctx, &paymentProto.InitializeUserRequest{
				UserId: authContext.User.ID,
			})
			if err != nil {
				s.log.Error("Failed to initialize payment service for user", "error", err, "userID", authContext.User.ID)
				// Don't fail the entire request, just log the error
			} else if paymentResp != nil && !paymentResp.Success {
				s.log.Error("Payment service initialization failed", "error", paymentResp.ErrorMessage, "userID", authContext.User.ID)
			} else {
				s.log.Info("Payment service initialized successfully", "userID", authContext.User.ID)
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
				s.log.Warn("Existing user has no payment service record, initializing", "userID", authContext.User.ID)
				// Initialize payment service for existing user (migration)
				initResp, err := s.paymentServiceClient.InitializeUser(ctx, &paymentProto.InitializeUserRequest{
					UserId: authContext.User.ID,
				})
				if err != nil {
					s.log.Error("Failed to initialize payment service for existing user", "error", err, "userID", authContext.User.ID)
				} else if initResp != nil && !initResp.Success {
					s.log.Error("Payment service initialization failed for existing user", "error", initResp.ErrorMessage, "userID", authContext.User.ID)
				} else {
					s.log.Info("Payment service initialized successfully for existing user", "userID", authContext.User.ID)
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

	// Check payment access for creating new organizational workspace (free users restricted to personal only)
	s.log.Info("Checking payment access for workspace creation", "userID", authContext.User.ID)
	accessResp, err := s.paymentServiceClient.GetUserSubscription(ctx, &paymentProto.GetUserSubscriptionRequest{
		UserId: authContext.User.ID,
	})
	if err != nil {
		s.log.Error("Payment service error while checking workspace creation", "error", err, "userID", authContext.User.ID)
		return nil, status.Error(codes.Internal, "payment service unavailable")
	}

	if !accessResp.Success {
		return nil, status.Error(codes.FailedPrecondition, "payment service not initialized")
	}

	// Free users can only use personal workspace - block organizational workspace creation
	if accessResp.SubscriptionInfo.Plan.Id == "free" {
		s.log.Warn("Free user attempted to create organizational workspace", "userID", authContext.User.ID, "planID", accessResp.SubscriptionInfo.Plan.Id)
		return nil, status.Error(codes.PermissionDenied, "workspace creation requires paid subscription. Please upgrade your plan to create additional workspaces.")
	}

	// Paid users can create unlimited organizational workspaces
	s.log.Info("Paid user creating workspace", "userID", authContext.User.ID, "planID", accessResp.SubscriptionInfo.Plan.Id)

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

	// Check payment access for adding users (check against tenant owner's subscription)
	// Get tenant owner
	tenant, err := s.dbQueries.GetTenantByID(ctx, req.TenantId)
	if err != nil {
		s.log.Error("Failed to get tenant details", "error", err, "tenantID", req.TenantId)
		return nil, status.Error(codes.Internal, "failed to get tenant details")
	}

	// Check workspace type and apply appropriate logic
	if tenant.IsPersonal {
		s.log.Info("Adding user to personal workspace", "tenantOwner", tenant.CreatedBy, "tenantID", req.TenantId)
		// For personal workspaces: Check global user limit (5 free, 50 paid)
	} else {
		s.log.Info("Adding user to organizational workspace", "tenantOwner", tenant.CreatedBy, "tenantID", req.TenantId)
		// For organizational workspaces: Only paid users can create these, so check 50-user limit

		// First verify the owner has paid subscription (organizational workspaces require payment)
		subscriptionResp, err := s.paymentServiceClient.GetUserSubscription(ctx, &paymentProto.GetUserSubscriptionRequest{
			UserId: tenant.CreatedBy,
		})
		if err != nil {
			s.log.Error("Failed to get subscription for organizational workspace owner", "error", err, "tenantOwner", tenant.CreatedBy)
			return nil, status.Error(codes.Internal, "payment service unavailable")
		}

		if !subscriptionResp.Success || subscriptionResp.SubscriptionInfo.Plan.Id == "free" {
			s.log.Error("Free user owns organizational workspace - data inconsistency", "tenantOwner", tenant.CreatedBy, "tenantID", req.TenantId)
			return nil, status.Error(codes.FailedPrecondition, "organizational workspace requires paid subscription")
		}
	}

	// Check if tenant owner can add more users (applies to both personal and organizational)
	s.log.Info("Checking user limit for adding member", "tenantOwner", tenant.CreatedBy, "tenantID", req.TenantId, "isPersonal", tenant.IsPersonal)
	accessResp, err := s.paymentServiceClient.CheckUserAccess(ctx, &paymentProto.CheckUserAccessRequest{
		UserId:         tenant.CreatedBy, // Check tenant owner's subscription
		UsageType:      "users",
		RequestedUsage: 1, // Adding 1 user
	})
	if err != nil {
		s.log.Error("Payment service error while checking user limits", "error", err, "tenantOwner", tenant.CreatedBy)
		// Don't block the operation, just log the error for now
	} else if !accessResp.HasAccess {
		var errorMessage string
		switch accessResp.Reason {
		case "users_limit_exceeded":
			errorMessage = "Cannot add user: User limit exceeded. Please upgrade your plan to add more members."
		case "subscription_inactive":
			errorMessage = "Cannot add user: Subscription is inactive. Please reactivate to add members."
		default:
			errorMessage = "Cannot add user: Access denied. Please check your subscription status."
		}

		s.log.Warn("User addition blocked due to payment restrictions", "tenantOwner", tenant.CreatedBy, "reason", accessResp.Reason)
		return nil, status.Error(codes.FailedPrecondition, errorMessage)
	} else if accessResp.IsNearLimit && accessResp.WarningMessage != "" {
		s.log.Warn("Tenant owner approaching user limit", "tenantOwner", tenant.CreatedBy, "warning", accessResp.WarningMessage)
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

	return &proto.AddUserResponse{
		Message: "User added to tenant successfully",
	}, nil
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
