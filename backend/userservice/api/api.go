package api

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sortedstartup.com/stream/common/interceptors"
	"sortedstartup.com/stream/userservice/config"
	"sortedstartup.com/stream/userservice/db"
	"sortedstartup.com/stream/userservice/proto"
)

type UserAPI struct {
	config    config.UserServiceConfig
	db        *sql.DB
	log       *slog.Logger
	dbQueries *db.Queries
	proto.UnimplementedUserServiceServer
}

func NewUserAPI(config config.UserServiceConfig) (*UserAPI, error) {
	slog.Info("NewUserAPI")

	childLogger := slog.With("service", "UserAPI")

	_db, err := sql.Open(config.DB.Driver, config.DB.Url)
	if err != nil {
		return nil, err
	}

	dbQueries := db.New(_db)

	userAPI := &UserAPI{
		config:    config,
		db:        _db,
		log:       childLogger,
		dbQueries: dbQueries,
	}
	return userAPI, nil
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

func (s *UserAPI) CreateUserIfNotExists(ctx context.Context, req *proto.GetUserByEmailRequest) (*proto.GetUserByEmailResponse, error) {
	s.log.Info("CreateUserIfNotExists")
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userEmail := authContext.User.Email
	s.log.Info("userEmail", "userEmail", userEmail)

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

			err = s.createPersonalTenant(ctx)
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
		}
		successMessage = "User already exists"
	}

	// Return success response with message
	return &proto.GetUserByEmailResponse{
		Message: successMessage,
		Success: true,
	}, nil
}

/**
* createPersonalTenant creates a personal tenant for a new user
* @param ctx context.Context
* @return error
 */
func (s *UserAPI) createPersonalTenant(ctx context.Context) error {

	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get auth context: %w", err)
	}
	userName := authContext.User.Name

	req := &proto.CreateTenantRequest{
		Name:        userName,
		Description: "Personal workspace",
	}

	tenant, err := s.CreateTenant(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create personal tenant: %w", err)
	}

	tenantID := tenant.Tenant.Id
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
func (s *UserAPI) CreateTenant(ctx context.Context, req *proto.CreateTenantRequest) (*proto.CreateTenantResponse, error) {
	s.log.Info("CreateTenant", "name", req.Name)

	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant name is required")
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
		Role:      "super_admin",
		CreatedAt: time.Now(),
	}

	_, err = s.dbQueries.CreateTenantUser(ctx, tenantUserParams)
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
		Success: true,
		Tenant:  protoTenant,
	}, nil
}

/**
* GetUserTenants returns all tenants a user belongs to
* @param ctx context.Context
* @param req *proto.GetUserTenantsRequest
* @return *proto.GetUserTenantsResponse, error
 */
func (s *UserAPI) GetUserTenants(ctx context.Context, req *proto.GetUserTenantsRequest) (*proto.GetUserTenantsResponse, error) {

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

	var tenants []*proto.Tenant
	for _, row := range tenantRows {
		tenant := &proto.Tenant{
			Id:          row.TenantID,
			Name:        row.Name,
			Description: row.Description.String,
			IsPersonal:  row.IsPersonal,
			CreatedAt:   timestamppb.New(row.CreatedAt),
			CreatedBy:   row.CreatedBy,
		}
		tenants = append(tenants, tenant)
	}

	return &proto.GetUserTenantsResponse{
		Message: "User tenants retrieved successfully",
		Success: true,
		Tenants: tenants,
	}, nil
}

/**
* AddUserToTenant adds a user to an existing tenant using username (email)
* @param ctx context.Context
* @param req *proto.AddUserToTenantRequest
* @return *proto.AddUserToTenantResponse, error
 */
func (s *UserAPI) AddUserToTenant(ctx context.Context, req *proto.AddUserToTenantRequest) (*proto.AddUserToTenantResponse, error) {
	s.log.Info("AddUserToTenant", "tenantID", req.TenantId, "username", req.Username)

	_, err := interceptors.AuthFromContext(ctx)
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
		role = "member"
	}

	// TODO: Add authorization check - only super_admin/admin can add users
	// We will add using OPA
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

	return &proto.AddUserToTenantResponse{
		Message: "User added to tenant successfully",
		Success: true,
	}, nil
}

/**
* GetTenantUsers returns all users in a tenant
* @param ctx context.Context
* @param req *proto.GetTenantUsersRequest
* @return *proto.GetTenantUsersResponse, error
 */
func (s *UserAPI) GetTenantUsers(ctx context.Context, req *proto.GetTenantUsersRequest) (*proto.GetTenantUsersResponse, error) {
	s.log.Info("GetTenantUsers", "tenantID", req.TenantId)

	_, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	// Validate input
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	// TODO: Add authorization check - only tenant members can view users
	// This will be implemented with OPA

	userRows, err := s.dbQueries.GetTenantUsers(ctx, req.TenantId)
	if err != nil {
		s.log.Error("Failed to get tenant users", "error", err)
		return nil, status.Error(codes.Internal, "failed to get tenant users")
	}

	// Convert to proto response
	var tenantUsers []*proto.TenantUser
	for _, row := range userRows {
		tenantUser := &proto.TenantUser{
			Id:        row.ID,
			TenantId:  row.TenantID,
			UserId:    row.UserID,
			Role:      row.Role,
			CreatedAt: timestamppb.New(row.CreatedAt),
			User: &proto.User{
				Id:       row.UserID,
				Username: row.Username,
				Email:    row.Email,
			},
		}
		tenantUsers = append(tenantUsers, tenantUser)
	}

	return &proto.GetTenantUsersResponse{
		Message:     "Tenant users retrieved successfully",
		Success:     true,
		TenantUsers: tenantUsers,
	}, nil
}
