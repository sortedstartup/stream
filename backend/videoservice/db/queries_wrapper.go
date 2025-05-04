package db

import (
	"context"
)

// QueriesWrapper wraps *Queries to implement QueriesInterface
type QueriesWrapper struct {
	*Queries
}

// CreateVideoUploaded implements QueriesInterface
func (qw *QueriesWrapper) CreateVideoUploaded(ctx context.Context, arg CreateVideoUploadedParams) (Video, error) {
	err := qw.Queries.CreateVideoUploaded(ctx, arg)
	if err != nil {
		return Video{}, err
	}

	return Video{
		ID:             arg.ID,
		Title:          arg.Title,
		Description:    arg.Description,
		Url:            arg.Url,
		UploadedUserID: arg.UploadedUserID,
	}, nil
}
