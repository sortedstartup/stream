package db

import (
	"context"
	"database/sql"
	"time"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

type Queries struct {
	db DBTX
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db: tx,
	}
}

type Comment struct {
	ID        string
	Content   string
	VideoID   string
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

const createComment = `-- name: CreateComment :exec
INSERT INTO comments (
    id,
    content,
    video_id,
    user_id,
    created_at,
    updated_at
) VALUES (
    ?1,
    ?2,
    ?3,
    ?4,
    ?5,
    ?6
)
`

type CreateCommentParams struct {
	ID        string
	Content   string
	VideoID   string
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (q *Queries) CreateComment(ctx context.Context, arg CreateCommentParams) error {
	_, err := q.db.ExecContext(ctx, createComment,
		arg.ID,
		arg.Content,
		arg.VideoID,
		arg.UserID,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}

const getCommentByID = `-- name: GetCommentByID :one
SELECT id, content, video_id, user_id, created_at, updated_at FROM comments 
WHERE id = ?1 AND user_id = ?2
LIMIT 1
`

type GetCommentByIDParams struct {
	ID     string
	UserID string
}

func (q *Queries) GetCommentByID(ctx context.Context, arg GetCommentByIDParams) (Comment, error) {
	row := q.db.QueryRowContext(ctx, getCommentByID, arg.ID, arg.UserID)
	var i Comment
	err := row.Scan(
		&i.ID,
		&i.Content,
		&i.VideoID,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getAllCommentsByUserPaginated = `-- name: GetAllCommentsByUserPaginated :many
SELECT id, content, video_id, user_id, created_at, updated_at FROM comments 
WHERE user_id = ?1
ORDER BY created_at DESC
LIMIT ?3 OFFSET ?2
`

type GetAllCommentsByUserPaginatedParams struct {
	UserID     string
	PageNumber int64
	PageSize   int64
}

func (q *Queries) GetAllCommentsByUserPaginated(ctx context.Context, arg GetAllCommentsByUserPaginatedParams) ([]Comment, error) {
	rows, err := q.db.QueryContext(ctx, getAllCommentsByUserPaginated, arg.UserID, arg.PageNumber, arg.PageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Comment
	for rows.Next() {
		var i Comment
		if err := rows.Scan(
			&i.ID,
			&i.Content,
			&i.VideoID,
			&i.UserID,
			&i.CreatedAt,
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
