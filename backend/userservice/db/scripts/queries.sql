-- name: CreateUser :one
INSERT INTO users (
    id,
    username,
    email,
    avatar_url,
    created_at
) VALUES (
    @id,
    @username,
    @email,
    @avatar_url,
    @created_at
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users 
WHERE email = @email;

-- name: UpdateUser :one
UPDATE users 
SET 
    username = COALESCE(@username, username),
    email = COALESCE(@email, email),
    avatar_url = COALESCE(@avatar_url, avatar_url)  
WHERE id = @id
RETURNING *;