// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: queries.sql

package db

import (
	"context"
	"database/sql"
	"time"
)

const createTenant = `-- name: CreateTenant :one
INSERT INTO userservice_tenants (
    id,
    name,
    description,
    is_personal,
    created_at,
    created_by
) VALUES (
    ?1,
    ?2,
    ?3,
    ?4,
    ?5,
    ?6
) RETURNING id, name, description, is_personal, created_at, created_by
`

type CreateTenantParams struct {
	ID          string
	Name        string
	Description sql.NullString
	IsPersonal  bool
	CreatedAt   time.Time
	CreatedBy   string
}

// Tenant queries
func (q *Queries) CreateTenant(ctx context.Context, arg CreateTenantParams) (UserserviceTenant, error) {
	row := q.db.QueryRowContext(ctx, createTenant,
		arg.ID,
		arg.Name,
		arg.Description,
		arg.IsPersonal,
		arg.CreatedAt,
		arg.CreatedBy,
	)
	var i UserserviceTenant
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.IsPersonal,
		&i.CreatedAt,
		&i.CreatedBy,
	)
	return i, err
}

const createTenantUser = `-- name: CreateTenantUser :one
INSERT INTO userservice_tenant_users (
    id,
    tenant_id,
    user_id,
    role,
    created_at
) VALUES (
    ?1,
    ?2,
    ?3,
    ?4,
    ?5
) RETURNING id, tenant_id, user_id, role, created_at
`

type CreateTenantUserParams struct {
	ID        string
	TenantID  string
	UserID    string
	Role      string
	CreatedAt time.Time
}

func (q *Queries) CreateTenantUser(ctx context.Context, arg CreateTenantUserParams) (UserserviceTenantUser, error) {
	row := q.db.QueryRowContext(ctx, createTenantUser,
		arg.ID,
		arg.TenantID,
		arg.UserID,
		arg.Role,
		arg.CreatedAt,
	)
	var i UserserviceTenantUser
	err := row.Scan(
		&i.ID,
		&i.TenantID,
		&i.UserID,
		&i.Role,
		&i.CreatedAt,
	)
	return i, err
}

const createUser = `-- name: CreateUser :one
INSERT INTO userservice_users (
    id,
    username,
    email,
    created_at
) VALUES (
    ?1,
    ?2,
    ?3,
    ?4
) RETURNING id, username, email, created_at
`

type CreateUserParams struct {
	ID        string
	Username  string
	Email     string
	CreatedAt time.Time
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (UserserviceUser, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.ID,
		arg.Username,
		arg.Email,
		arg.CreatedAt,
	)
	var i UserserviceUser
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.CreatedAt,
	)
	return i, err
}

const getTenantUsers = `-- name: GetTenantUsers :many
SELECT 
    tu.role, tu.created_at,
    u.username, u.email,
    t.created_at as tenant_created_at,
    t.name as tenant_name
FROM userservice_tenant_users tu
JOIN userservice_users u ON tu.user_id = u.id
JOIN userservice_tenants t ON tu.tenant_id = t.id
WHERE tu.tenant_id = ?1
ORDER BY tu.created_at ASC
`

type GetTenantUsersRow struct {
	Role            string
	CreatedAt       time.Time
	Username        string
	Email           string
	TenantCreatedAt time.Time
	TenantName      string
}

func (q *Queries) GetTenantUsers(ctx context.Context, tenantID string) ([]GetTenantUsersRow, error) {
	rows, err := q.db.QueryContext(ctx, getTenantUsers, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTenantUsersRow
	for rows.Next() {
		var i GetTenantUsersRow
		if err := rows.Scan(
			&i.Role,
			&i.CreatedAt,
			&i.Username,
			&i.Email,
			&i.TenantCreatedAt,
			&i.TenantName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, username, email, created_at FROM userservice_users 
WHERE email = ?1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (UserserviceUser, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i UserserviceUser
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.CreatedAt,
	)
	return i, err
}

const getUserRoleInTenant = `-- name: GetUserRoleInTenant :one
SELECT role FROM userservice_tenant_users 
WHERE tenant_id = ?1 AND user_id = ?2
`

type GetUserRoleInTenantParams struct {
	TenantID string
	UserID   string
}

func (q *Queries) GetUserRoleInTenant(ctx context.Context, arg GetUserRoleInTenantParams) (string, error) {
	row := q.db.QueryRowContext(ctx, getUserRoleInTenant, arg.TenantID, arg.UserID)
	var role string
	err := row.Scan(&role)
	return role, err
}

const getUserTenants = `-- name: GetUserTenants :many
SELECT 
    tu.id as tenant_user_id,
    t.id as tenant_id, 
    t.name, 
    t.description, 
    t.is_personal, 
    t.created_at, 
    t.created_by,
    tu.role, 
    tu.created_at as joined_at
FROM userservice_tenants t
JOIN userservice_tenant_users tu ON t.id = tu.tenant_id
WHERE tu.user_id = ?1
ORDER BY t.created_at DESC
`

type GetUserTenantsRow struct {
	TenantUserID string
	TenantID     string
	Name         string
	Description  sql.NullString
	IsPersonal   bool
	CreatedAt    time.Time
	CreatedBy    string
	Role         string
	JoinedAt     time.Time
}

func (q *Queries) GetUserTenants(ctx context.Context, userID string) ([]GetUserTenantsRow, error) {
	rows, err := q.db.QueryContext(ctx, getUserTenants, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUserTenantsRow
	for rows.Next() {
		var i GetUserTenantsRow
		if err := rows.Scan(
			&i.TenantUserID,
			&i.TenantID,
			&i.Name,
			&i.Description,
			&i.IsPersonal,
			&i.CreatedAt,
			&i.CreatedBy,
			&i.Role,
			&i.JoinedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
