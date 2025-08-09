package db

import "context"

type Querier interface {
    GetUserByEmail(ctx context.Context, email string) (UserserviceUser, error)
    CreateUser(ctx context.Context, params CreateUserParams) (UserserviceUser, error)
    GetUserTenants(ctx context.Context, userID string) ([]GetUserTenantsRow, error)
	GetTenantUsers(ctx context.Context, tenantID string) ([]GetTenantUsersRow, error)
    CreateTenant(ctx context.Context, params CreateTenantParams) (UserserviceTenant, error)
    CreateTenantUser(ctx context.Context, params CreateTenantUserParams) (UserserviceTenantUser, error)
    GetUserRoleInTenant(ctx context.Context, params GetUserRoleInTenantParams) (string, error)
}


var _ Querier = (*Queries)(nil)

