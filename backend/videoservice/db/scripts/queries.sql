-- name: GetAllVideoUploadedByUserPaginated :many
SELECT * FROM videoservice_videos 
WHERE uploaded_user_id = @user_id AND is_deleted = FALSE
ORDER BY created_at DESC
LIMIT @page_size OFFSET @page_number;

-- name: CreateVideoUploaded :exec
INSERT INTO videoservice_videos (
    id,
    title,
    description,
    url,
    uploaded_user_id,
    tenant_id,
    channel_id,
    is_private,
    is_deleted,
    created_at,
    updated_at
) VALUES (
    @id,
    @title,
    @description, 
    @url,
    @uploaded_user_id,
    @tenant_id,
    @channel_id,
    @is_private,
    @is_deleted,
    @created_at,
    @updated_at
);

-- name: GetVideoByVideoIDAndTenantID :one
SELECT * FROM videoservice_videos 
WHERE id = @id AND tenant_id = @tenant_id AND is_deleted = FALSE
LIMIT 1;

-- name: GetVideosByTenantID :many
SELECT * FROM videoservice_videos 
WHERE tenant_id = @tenant_id AND is_deleted = FALSE
ORDER BY created_at DESC;

-- name: GetVideosByTenantIDAndChannelID :many
SELECT * FROM videoservice_videos 
WHERE tenant_id = @tenant_id AND channel_id = @channel_id AND is_deleted = FALSE
ORDER BY created_at DESC;

-- name: GetAllAccessibleVideosByTenantID :many
SELECT DISTINCT v.* FROM videoservice_videos v
LEFT JOIN videoservice_channels c ON v.channel_id = c.id
LEFT JOIN videoservice_channel_members cm ON c.id = cm.channel_id
WHERE v.tenant_id = @tenant_id AND v.is_deleted = FALSE
  AND (
    -- User's own videos (private)
    (v.uploaded_user_id = @user_id AND (v.channel_id IS NULL OR v.channel_id = ''))
    OR 
    -- Videos in channels user is member of
    (v.channel_id IS NOT NULL AND v.channel_id != '' AND cm.user_id = @user_id)
  )
ORDER BY v.created_at DESC;

-- name: UpdateVideoTitleDescriptionIsPrivate :exec
UPDATE videoservice_videos
SET title = @title,
    description = @description,
    is_private = @is_private,
    updated_at = CURRENT_TIMESTAMP
WHERE id = @id
  AND is_deleted = FALSE;

-- Channel queries
-- name: CreateChannel :one
INSERT INTO videoservice_channels (
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
SELECT * FROM videoservice_channels 
WHERE tenant_id = @tenant_id
ORDER BY created_at DESC;

-- name: GetChannelByIDAndTenantID :one
SELECT * FROM videoservice_channels 
WHERE id = @id AND tenant_id = @tenant_id;

-- name: UpdateChannel :one
UPDATE videoservice_channels 
SET 
    name = @name,
    description = @description,
    updated_at = @updated_at
WHERE id = @id AND tenant_id = @tenant_id
RETURNING *;

-- name: CreateChannelMember :one
INSERT INTO videoservice_channel_members (
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
FROM videoservice_channel_members cm
JOIN videoservice_channels c ON cm.channel_id = c.id
WHERE cm.channel_id = @channel_id AND c.tenant_id = @tenant_id
ORDER BY cm.created_at ASC;

-- name: GetChannelMembersByChannelIDExcludingUser :many
SELECT 
    cm.id as channel_member_id,
    cm.channel_id,
    cm.user_id,
    cm.role,
    cm.added_by,
    cm.created_at,
    c.name as channel_name,
    c.tenant_id
FROM videoservice_channel_members cm
JOIN videoservice_channels c ON cm.channel_id = c.id
WHERE cm.channel_id = @channel_id AND c.tenant_id = @tenant_id AND cm.user_id != @user_id
ORDER BY cm.created_at ASC;

-- name: GetUserRoleInChannel :one
SELECT cm.role FROM videoservice_channel_members cm
JOIN videoservice_channels c ON cm.channel_id = c.id
WHERE cm.channel_id = @channel_id AND cm.user_id = @user_id AND c.tenant_id = @tenant_id;

-- name: DeleteChannelMember :exec
DELETE FROM videoservice_channel_members 
WHERE channel_id = @channel_id AND user_id = @user_id;

-- Video-Channel Management Queries
-- name: UpdateVideoChannel :exec
UPDATE videoservice_videos 
SET channel_id = @channel_id, updated_at = @updated_at
WHERE id = @video_id AND tenant_id = @tenant_id AND is_deleted = FALSE
  AND (
    -- For tenant-level videos: validate uploader ownership
    (channel_id IS NULL OR channel_id = '') AND uploaded_user_id = @uploaded_user_id
    OR
    -- For channel videos: validate current channel (permission checked in API)
    channel_id = @current_channel_id
  );

-- name: RemoveVideoFromChannel :exec
UPDATE videoservice_videos 
SET channel_id = NULL, updated_at = @updated_at
WHERE id = @video_id AND tenant_id = @tenant_id AND channel_id = @channel_id AND is_deleted = FALSE;

-- name: SoftDeleteVideo :exec
UPDATE videoservice_videos 
SET is_deleted = TRUE, updated_at = @updated_at
WHERE id = @video_id AND tenant_id = @tenant_id AND is_deleted = FALSE;

