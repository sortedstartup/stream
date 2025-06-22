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
    @created_at
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users 
WHERE email = @email;