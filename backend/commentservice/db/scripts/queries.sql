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

-- name: GetComentsAndRepliesForVideoID :many
SELECT 
    c1.id, 
    c1.content, 
    c1.video_id, 
    c1.user_id, 
    c1.parent_comment_id,
    c1.created_at,  
    c1.updated_at,  
    COALESCE(
        json_group_array(
            json_object(
                'id', c2.id,
                'content', c2.content,
                'user_id', c2.user_id,
                'video_id', c2.video_id,
                'parent_comment_id', c2.parent_comment_id,
                'created_at', datetime(c2.created_at, 'unixepoch'), 
                'updated_at', datetime(c2.updated_at, 'unixepoch')   
            )
        ) FILTER (WHERE c2.id IS NOT NULL), 
        '[]'
    ) AS replies
FROM comments c1
LEFT JOIN comments c2 ON c1.id = c2.parent_comment_id
WHERE c1.video_id = @video_id 
AND c1.parent_comment_id IS NULL
GROUP BY c1.id
ORDER BY c1.created_at DESC;

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