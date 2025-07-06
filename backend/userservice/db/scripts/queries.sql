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
    tu.role, tu.created_at,
    u.username, u.email,
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

-- Channel queries
-- name: CreateChannel :one
INSERT INTO userservice_channels (
    id,
    tenant_id,
    name,
    description,
    is_private,
    created_by,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @name,
    @description,
    @is_private,
    @created_by,
    @created_at,
    @updated_at
) RETURNING *;

-- name: GetChannelsByTenantID :many
SELECT * FROM userservice_channels 
WHERE tenant_id = @tenant_id
ORDER BY created_at DESC;

-- name: GetChannelByIDAndTenantID :one
SELECT * FROM userservice_channels 
WHERE id = @id AND tenant_id = @tenant_id;

-- name: UpdateChannel :one
UPDATE userservice_channels 
SET 
    name = @name,
    description = @description,
    updated_at = @updated_at
WHERE id = @id AND tenant_id = @tenant_id
RETURNING *;

-- name: CreateChannelMember :one
INSERT INTO userservice_channel_members (
    id,
    channel_id,
    user_id,
    role,
    added_by,
    created_at
) VALUES (
    @id,
    @channel_id,
    @user_id,
    @role,
    @added_by,
    @created_at
) RETURNING *;

-- name: GetChannelMembersByChannelIDAndTenantID :many
SELECT 
    cm.id as channel_member_id,
    cm.channel_id,
    cm.user_id,
    cm.role,
    cm.added_by,
    cm.created_at,
    u.username,
    u.email,
    c.name as channel_name,
    c.tenant_id
FROM userservice_channel_members cm
JOIN userservice_users u ON cm.user_id = u.id
JOIN userservice_channels c ON cm.channel_id = c.id
WHERE cm.channel_id = @channel_id AND c.tenant_id = @tenant_id
ORDER BY cm.created_at ASC;

-- name: GetUserRoleInChannel :one
SELECT cm.role FROM userservice_channel_members cm
JOIN userservice_channels c ON cm.channel_id = c.id
WHERE cm.channel_id = @channel_id AND cm.user_id = @user_id AND c.tenant_id = @tenant_id;

-- name: DeleteChannelMember :exec
DELETE FROM userservice_channel_members 
WHERE channel_id = @channel_id AND user_id = @user_id;