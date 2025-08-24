-- name: CreateUser :one
INSERT INTO userservice_users (
    id,
    username,
    email,
    created_at
) VALUES (
    @id,
    @username,
    @email,
    @created_at
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM userservice_users 
WHERE email = @email;

-- Tenant queries
-- name: GetTenantByName :one
SELECT * FROM userservice_tenants 
WHERE name = @name AND created_by = @created_by;

-- name: CreateTenant :one
INSERT INTO userservice_tenants (
    id,
    name,
    description,
    is_personal,
    created_at,
    created_by
) VALUES (
    @id,
    @name,
    @description,
    @is_personal,
    @created_at,
    @created_by
) RETURNING id, name, description, is_personal, created_at, created_by;

-- name: GetTenantByID :one
SELECT * FROM userservice_tenants WHERE id = @id;

-- name: GetUserTenants :many
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
WHERE tu.user_id = @user_id
ORDER BY t.created_at DESC;

-- name: CreateTenantUser :one
INSERT INTO userservice_tenant_users (
    id,
    tenant_id,
    user_id,
    role,
    created_at
) VALUES (
    @id,
    @tenant_id,
    @user_id,
    @role,
    @created_at
) RETURNING id, tenant_id, user_id, role, created_at;

-- name: GetTenantUsers :many
SELECT 
    tu.role, tu.created_at,
    u.id as user_id, u.username, u.email,
    t.created_at as tenant_created_at,
    t.name as tenant_name
FROM userservice_tenant_users tu
JOIN userservice_users u ON tu.user_id = u.id
JOIN userservice_tenants t ON tu.tenant_id = t.id
WHERE tu.tenant_id = @tenant_id
ORDER BY tu.created_at ASC;

-- name: GetUserRoleInTenant :one
SELECT role FROM userservice_tenant_users 
WHERE tenant_id = @tenant_id AND user_id = @user_id;

-- User limits queries (moved from payment service)
-- name: GetUserLimits :one
SELECT * FROM userservice_user_limits WHERE user_id = ? LIMIT 1;

-- name: CreateUserLimits :one
INSERT INTO userservice_user_limits (
    user_id, users_count, last_calculated_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: AddUserCount :exec
UPDATE userservice_user_limits 
SET users_count = users_count + ?, last_calculated_at = ?, updated_at = ?
WHERE user_id = ?;

-- name: UpsertUserLimits :exec
INSERT INTO userservice_user_limits (
    user_id, users_count, last_calculated_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?)
ON CONFLICT(user_id) DO UPDATE SET
    users_count = excluded.users_count,
    last_calculated_at = excluded.last_calculated_at,
    updated_at = excluded.updated_at;