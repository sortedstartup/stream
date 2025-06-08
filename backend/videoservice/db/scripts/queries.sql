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

-- name: GetVideoByIDWithAccess :one
SELECT v.* FROM videos v
WHERE v.id = @video_id 
    AND (
        v.uploaded_user_id = @user_id  -- User owns the video
        OR EXISTS (  -- Or user has access through a shared space
            SELECT 1 FROM video_spaces vs
            INNER JOIN spaces s ON vs.space_id = s.id
            LEFT JOIN user_spaces us ON s.id = us.space_id
            WHERE vs.video_id = v.id
                AND (s.user_id = @user_id OR us.user_id = @user_id)
        )
    )
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
SELECT s.*, 'owner' as access_level FROM spaces s
WHERE s.user_id = @user_id 
UNION ALL
SELECT s.*, us.access_level FROM spaces s
INNER JOIN user_spaces us ON s.id = us.space_id
WHERE us.user_id = @user_id
ORDER BY created_at DESC;

-- name: GetOwnedSpacesByUser :many
SELECT * FROM spaces 
WHERE user_id = @user_id 
ORDER BY created_at DESC;

-- name: GetSharedSpacesByUser :many
SELECT s.*, us.access_level FROM spaces s
INNER JOIN user_spaces us ON s.id = us.space_id
WHERE us.user_id = @user_id
ORDER BY s.created_at DESC;

-- name: GetSpaceByID :one
SELECT * FROM spaces 
WHERE id = @id AND user_id = @user_id
LIMIT 1;

-- name: GetSpaceByIDWithAccess :one
SELECT s.*, 
    CASE 
        WHEN s.user_id = @user_id THEN 'owner'
        ELSE us.access_level
    END as access_level
FROM spaces s
LEFT JOIN user_spaces us ON s.id = us.space_id AND us.user_id = @user_id
WHERE s.id = @space_id AND (s.user_id = @user_id OR us.user_id = @user_id)
LIMIT 1;

-- name: GetVideosInSpace :many
SELECT v.* FROM videos v
INNER JOIN video_spaces vs ON v.id = vs.video_id
WHERE vs.space_id = @space_id AND v.uploaded_user_id = @user_id
ORDER BY v.created_at DESC;

-- name: GetVideosInSpaceWithAccess :many
SELECT v.* FROM videos v
INNER JOIN video_spaces vs ON v.id = vs.video_id
INNER JOIN spaces s ON vs.space_id = s.id
LEFT JOIN user_spaces us ON s.id = us.space_id AND us.user_id = @user_id
WHERE vs.space_id = @space_id 
    AND (s.user_id = @user_id OR us.user_id = @user_id)
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

-- Space sharing queries

-- name: AddUserToSpace :exec
INSERT INTO user_spaces (
    user_id,
    space_id,
    access_level,
    created_at,
    updated_at
) VALUES (
    @user_id,
    @space_id,
    @access_level,
    @created_at,
    @updated_at
);

-- name: RemoveUserFromSpace :exec
DELETE FROM user_spaces 
WHERE user_id = @user_id AND space_id = @space_id;

-- name: UpdateUserSpaceAccess :exec
UPDATE user_spaces 
SET access_level = @access_level, updated_at = @updated_at
WHERE user_id = @user_id AND space_id = @space_id;

-- name: GetSpaceMembers :many
SELECT us.user_id, us.access_level, us.created_at, us.updated_at
FROM user_spaces us
WHERE us.space_id = @space_id
ORDER BY us.created_at ASC;

-- name: CheckUserSpaceAccess :one
SELECT us.access_level FROM user_spaces us
WHERE us.user_id = @user_id AND us.space_id = @space_id
LIMIT 1;

-- name: IsSpaceOwner :one
SELECT COUNT(*) as is_owner FROM spaces
WHERE id = @space_id AND user_id = @user_id;

-- User queries for sharing

-- name: GetAllUsers :many
SELECT DISTINCT uploaded_user_id as user_id FROM videos
WHERE uploaded_user_id != @exclude_user_id
ORDER BY uploaded_user_id;

