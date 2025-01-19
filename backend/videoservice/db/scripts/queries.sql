-- name: GetAllVideoUploadedByUser :many
SELECT * FROM videos WHERE uploaded_user_id = @id;

-- name: GetAllVideoUploadedByUserPaginated :many
SELECT * FROM videos 
WHERE uploaded_user_id = @user_id
ORDER BY created_at
DESC LIMIT @page_size OFFSET @page_number;

-- name: CreateVideoUploaded :exec
INSERT INTO videos (
    id,
    title,
    description,
    url,
    uploaded_user_id,
    created_at,
    updated_at
) VALUES (
    @id,
    @title,
    @description, 
    @url,
    @uploaded_user_id,
    @created_at,
    @updated_at
);

-- name: GetVideoByID :one
SELECT * FROM videos 
WHERE id = @id AND uploaded_user_id = @user_id
LIMIT 1;

