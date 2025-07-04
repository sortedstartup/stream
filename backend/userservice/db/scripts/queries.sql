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
    tu.id, tu.tenant_id, tu.user_id, tu.role, tu.created_at,
    u.username, u.email
FROM userservice_tenant_users tu
JOIN userservice_users u ON tu.user_id = u.id
WHERE tu.tenant_id = @tenant_id
ORDER BY tu.created_at ASC;