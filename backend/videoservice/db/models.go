// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package db

import (
	"time"
)

type Video struct {
	ID             string
	Title          string
	Description    string
	Url            string
	CreatedAt      time.Time
	UploadedUserID string
	UpdatedAt      time.Time
}
