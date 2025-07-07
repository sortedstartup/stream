-- name: GetAllVideoUploadedByUserPaginated :many
SELECT * FROM videos 
WHERE uploaded_user_id = @user_id
ORDER BY created_at DESC
LIMIT @page_size OFFSET @page_number;

-- name: CreateVideoUploaded :exec
INSERT INTO videos (
    id,
    title,
    description,
    url,
    uploaded_user_id,
    tenant_id,
    is_private,
    created_at,
    updated_at
) VALUES (
    @id,
    @title,
    @description, 
    @url,
    @uploaded_user_id,
    @tenant_id,
    @is_private,
    @created_at,
    @updated_at
);

-- name: GetVideoByVideoIDAndTenantID :one
SELECT * FROM videos 
WHERE id = @id AND tenant_id = @tenant_id
LIMIT 1;

-- name: GetVideosByTenantID :many
SELECT * FROM videos 
WHERE tenant_id = @tenant_id 
ORDER BY created_at DESC;

-- name: GetVideosByTenantIDAndChannelID :many
SELECT * FROM videos 
WHERE tenant_id = @tenant_id AND channel_id = @channel_id
ORDER BY created_at DESC;

-- Channel queries
-- name: CreateChannel :one
INSERT INTO channels (
    id,
    tenant_id,
    name,
    description,
    created_by,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @name,
    @description,
    @created_by,
    @created_at,
    @updated_at
) RETURNING *;

-- name: GetChannelsByTenantID :many
SELECT * FROM channels 
WHERE tenant_id = @tenant_id
ORDER BY created_at DESC;

-- name: GetChannelByIDAndTenantID :one
SELECT * FROM channels 
WHERE id = @id AND tenant_id = @tenant_id;

-- name: UpdateChannel :one
UPDATE channels 
SET 
    name = @name,
    description = @description,
    updated_at = @updated_at
WHERE id = @id AND tenant_id = @tenant_id
RETURNING *;

-- name: CreateChannelMember :one
INSERT INTO channel_members (
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
    c.name as channel_name,
    c.tenant_id
FROM channel_members cm
JOIN channels c ON cm.channel_id = c.id
WHERE cm.channel_id = @channel_id AND c.tenant_id = @tenant_id
ORDER BY cm.created_at ASC;

-- name: GetUserRoleInChannel :one
SELECT cm.role FROM channel_members cm
JOIN channels c ON cm.channel_id = c.id
WHERE cm.channel_id = @channel_id AND cm.user_id = @user_id AND c.tenant_id = @tenant_id;

-- name: DeleteChannelMember :exec
DELETE FROM channel_members 
WHERE channel_id = @channel_id AND user_id = @user_id;

