-- name: GetAllVideo :many
SELECT * FROM videos;

-- name: CreateVideo :exec
INSERT INTO videos (id, title, description, url)
VALUES (?, ?, ?, ?);