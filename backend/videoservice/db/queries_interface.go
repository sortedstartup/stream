package db

import "context"

type QueriesInterface interface {
	GetAllVideoUploadedByUserPaginated(ctx context.Context, params GetAllVideoUploadedByUserPaginatedParams) ([]Video, error)
	GetVideoByID(ctx context.Context, params GetVideoByIDParams) (Video, error)
	CreateVideoUploaded(ctx context.Context, params CreateVideoUploadedParams) (Video, error)
}
