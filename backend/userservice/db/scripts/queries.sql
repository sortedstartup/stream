-- name: CreateUser :one
INSERT INTO userservice_users (
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
SELECT * FROM userservice_users 
WHERE email = @email;