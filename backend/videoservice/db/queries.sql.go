// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package db

import (
	"context"
)

const getAllVideoUploadedByUser = `-- name: GetAllVideoUploadedByUser :many
SELECT id, title, description, url, created_at, uploaded_user_id FROM videos WHERE uploaded_user_id = ?1
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
SELECT id, title, description, url, created_at, uploaded_user_id FROM videos 
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
