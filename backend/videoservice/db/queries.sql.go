// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package db

import (
	"context"
	"time"
)

const createVideoUploaded = `-- name: CreateVideoUploaded :exec
INSERT INTO videos (
    id,
    title,
    description,
    url,
    uploaded_user_id,
    created_at,
    updated_at
) VALUES (
    ?1,
    ?2,
    ?3, 
    ?4,
    ?5,
    ?6,
    ?7
)
`

type CreateVideoUploadedParams struct {
	ID             string
	Title          string
	Description    string
	Url            string
	UploadedUserID string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (q *Queries) CreateVideoUploaded(ctx context.Context, arg CreateVideoUploadedParams) error {
	_, err := q.db.ExecContext(ctx, createVideoUploaded,
		arg.ID,
		arg.Title,
		arg.Description,
		arg.Url,
		arg.UploadedUserID,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}

const getAllVideoUploadedByUser = `-- name: GetAllVideoUploadedByUser :many
SELECT id, title, description, url, created_at, uploaded_user_id, updated_at FROM videos WHERE uploaded_user_id = ?1
`

func (q *Queries) GetAllVideoUploadedByUser(ctx context.Context, id string) ([]Video, error) {
	rows, err := q.db.QueryContext(ctx, getAllVideoUploadedByUser, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Video
	for rows.Next() {
		var i Video
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Description,
			&i.Url,
			&i.CreatedAt,
			&i.UploadedUserID,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllVideoUploadedByUserPaginated = `-- name: GetAllVideoUploadedByUserPaginated :many
SELECT id, title, description, url, created_at, uploaded_user_id, updated_at FROM videos 
WHERE uploaded_user_id = ?1
ORDER BY created_at
DESC LIMIT ?3 OFFSET ?2
`

type GetAllVideoUploadedByUserPaginatedParams struct {
	UserID     string
	PageNumber int64
	PageSize   int64
}

func (q *Queries) GetAllVideoUploadedByUserPaginated(ctx context.Context, arg GetAllVideoUploadedByUserPaginatedParams) ([]Video, error) {
	rows, err := q.db.QueryContext(ctx, getAllVideoUploadedByUserPaginated, arg.UserID, arg.PageNumber, arg.PageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Video
	for rows.Next() {
		var i Video
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Description,
			&i.Url,
			&i.CreatedAt,
			&i.UploadedUserID,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getVideoByID = `-- name: GetVideoByID :one
SELECT id, title, description, url, created_at, uploaded_user_id, updated_at FROM videos 
WHERE id = ?1 AND uploaded_user_id = ?2
LIMIT 1
`

type GetVideoByIDParams struct {
	ID     string
	UserID string
}

func (q *Queries) GetVideoByID(ctx context.Context, arg GetVideoByIDParams) (Video, error) {
	row := q.db.QueryRowContext(ctx, getVideoByID, arg.ID, arg.UserID)
	var i Video
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Description,
		&i.Url,
		&i.CreatedAt,
		&i.UploadedUserID,
		&i.UpdatedAt,
	)
	return i, err
}
