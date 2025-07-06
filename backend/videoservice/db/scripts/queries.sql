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

