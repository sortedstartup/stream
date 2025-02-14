-- name: test :many
select * from comments;

-- name: CreateComment :exec
INSERT INTO comments (
    id,
    content,
    video_id,
    user_id,
    parent_comment_id,
    created_at,
    updated_at
) VALUES (
    @id,
    @content,
    @video_id,
    @user_id,
    @parent_comment_id,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- name: GetCommentByID :one
SELECT * FROM comments 
WHERE id = @id AND user_id = @user_id
LIMIT 1;

-- name: GetAllCommentsByUserPaginated :many
SELECT * FROM comments 
WHERE user_id = @user_id
ORDER BY created_at DESC
LIMIT @page_size OFFSET (@page_number * @page_size);

-- name: GetCommentsByVideo :many
SELECT * FROM comments 
WHERE video_id = @video_id
ORDER BY created_at DESC;

-- name: GetCommentsByVideoPaginated :many
SELECT * FROM comments 
WHERE video_id = @video_id
ORDER BY created_at DESC
LIMIT @page_size OFFSET (@page_number * @page_size);

-- name: GetRepliesByCommentID :many
SELECT * FROM comments 
WHERE parent_comment_id = @comment_id
ORDER BY created_at ASC;

-- name: UpdateComment :exec
UPDATE comments 
SET content = @content, updated_at = CURRENT_TIMESTAMP
WHERE id = @id AND user_id = @user_id;

-- name: DeleteComment :exec
DELETE FROM comments 
WHERE id = @id AND user_id = @user_id;

-- name: GetCommentCount :one
SELECT COUNT(*) FROM comments WHERE video_id = @video_id;

-- name: LikeComment :exec
INSERT INTO comment_likes (
    id,
    user_id,
    comment_id,
    created_at
) VALUES (
    @id,
    @user_id,
    @comment_id,
    CURRENT_TIMESTAMP
);

-- name: UnlikeComment :exec
DELETE FROM comment_likes 
WHERE user_id = @user_id AND comment_id = @comment_id;

-- name: GetCommentLikesCount :one
SELECT COUNT(*) FROM comment_likes 
WHERE comment_id = @comment_id;

-- name: CheckUserLikedComment :one
SELECT COUNT(*) FROM comment_likes 
WHERE user_id = @user_id AND comment_id = @comment_id;