package db

import (
	"context"
	"database/sql"
)

type DBQuerier interface {
	CreateVideoUploaded(ctx context.Context, params CreateVideoUploadedParams) error
	GetVideoByVideoIDAndTenantID(ctx context.Context, params GetVideoByVideoIDAndTenantIDParams) (Video, error)
	GetVideosByTenantID(ctx context.Context, tenantID sql.NullString) ([]Video, error)
}

var _ DBQuerier = (*Queries)(nil)
