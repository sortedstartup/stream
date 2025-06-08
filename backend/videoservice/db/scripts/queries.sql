-- name: GetAllVideoUploadedByUser :many
SELECT * FROM videos WHERE uploaded_user_id = @id ORDER BY created_at DESC;

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

-- Space related queries

-- name: CreateSpace :exec
INSERT INTO spaces (
    id,
    name,
    description,
    user_id,
    created_at,
    updated_at
) VALUES (
    @id,
    @name,
    @description,
    @user_id,
    @created_at,
    @updated_at
);

-- name: GetSpacesByUser :many
SELECT * FROM spaces 
WHERE user_id = @user_id 
ORDER BY created_at DESC;

-- name: GetSpaceByID :one
SELECT * FROM spaces 
WHERE id = @id AND user_id = @user_id
LIMIT 1;

-- name: GetVideosInSpace :many
SELECT v.* FROM videos v
INNER JOIN video_spaces vs ON v.id = vs.video_id
WHERE vs.space_id = @space_id AND v.uploaded_user_id = @user_id
ORDER BY v.created_at DESC;

-- name: AddVideoToSpace :exec
INSERT INTO video_spaces (
    video_id,
    space_id,
    created_at,
    updated_at
) VALUES (
    @video_id,
    @space_id,
    @created_at,
    @updated_at
);

-- name: RemoveVideoFromSpace :exec
DELETE FROM video_spaces 
WHERE video_id = @video_id AND space_id = @space_id;

