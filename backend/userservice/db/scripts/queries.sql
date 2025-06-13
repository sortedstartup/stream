-- name: CreateUser :one
INSERT INTO users (
    id,
    username,
    email,
    created_at
) VALUES (
    @id,
    @username,
    @email,
    CURRENT_TIMESTAMP
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users 
WHERE id = @id;

-- name: GetUserByEmail :one
SELECT * FROM users 
WHERE email = @email;

-- name: UpdateUser :one
UPDATE users 
SET 
    username = COALESCE(@username, username),
    email = COALESCE(@email, email)
WHERE id = @id
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users 
WHERE id = @id;

-- name: ValidateUser :one
SELECT COUNT(*) as count FROM users 
WHERE id = @id; 