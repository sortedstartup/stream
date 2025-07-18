// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package db

import (
	"database/sql"
	"time"
)

type UserserviceTenant struct {
	ID          string
	Name        string
	Description sql.NullString
	IsPersonal  bool
	CreatedAt   time.Time
	CreatedBy   string
}

type UserserviceTenantUser struct {
	ID        string
	TenantID  string
	UserID    string
	Role      string
	CreatedAt time.Time
}

type UserserviceUser struct {
	ID        string
	Username  string
	Email     string
	CreatedAt time.Time
}
