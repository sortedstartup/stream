-- name: GetAllVideoUploadedByUser :many
SELECT * FROM videos WHERE uploaded_user_id = @id;

-- name: GetAllVideoUploadedByUserPaginated :many
SELECT * FROM videos 
WHERE uploaded_user_id = @user_id
ORDER BY created_at
DESC LIMIT @page_size OFFSET @page_number;