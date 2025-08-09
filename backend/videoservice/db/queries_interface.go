package db

import (
	"context"
	"database/sql"
)

type DBQuerier interface {
	CreateVideoUploaded(ctx context.Context, params CreateVideoUploadedParams) error
	GetVideoByVideoIDAndTenantID(ctx context.Context, params GetVideoByVideoIDAndTenantIDParams) (VideoserviceVideo, error)
	GetVideosByTenantID(ctx context.Context, tenantID sql.NullString) ([]VideoserviceVideo, error)
	GetVideosByTenantIDAndChannelID(ctx context.Context, params GetVideosByTenantIDAndChannelIDParams) ([]VideoserviceVideo, error)
    GetAllAccessibleVideosByTenantID(ctx context.Context, params GetAllAccessibleVideosByTenantIDParams) ([]VideoserviceVideo, error)
    UpdateVideoChannel(ctx context.Context, params UpdateVideoChannelParams) error
    RemoveVideoFromChannel(ctx context.Context, params RemoveVideoFromChannelParams) error
    SoftDeleteVideo(ctx context.Context, params SoftDeleteVideoParams) error
}

var _ DBQuerier = (*Queries)(nil)
