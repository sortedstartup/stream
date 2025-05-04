package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"sortedstartup.com/stream/common/auth"
	"sortedstartup.com/stream/videoservice/api"
	"sortedstartup.com/stream/videoservice/db"
	mockdb "sortedstartup.com/stream/videoservice/db/mock"
	"sortedstartup.com/stream/videoservice/proto"
)

func TestListVideos(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQueries := mockdb.NewMockQueriesInterface(ctrl)

	// Test user ID and pagination parameters
	userID := "user-123"
	pageSize := int32(2)
	pageNumber := int32(1)

	// Expected result from the mock DB call
	expectedVideos := []db.Video{
		{
			ID:             "vid-1",
			Title:          "First Video",
			Description:    "My first video",
			Url:            "https://example.com/vid1.mp4",
			CreatedAt:      time.Now().Add(-2 * time.Hour),
			UploadedUserID: userID,
		},
		{
			ID:             "vid-2",
			Title:          "Second Video",
			Description:    "My second video",
			Url:            "https://example.com/vid2.mp4",
			CreatedAt:      time.Now().Add(-1 * time.Hour),
			UploadedUserID: userID,
		},
	}

	// Setting up mock expectation
	mockQueries.EXPECT().
		GetAllVideoUploadedByUserPaginated(gomock.Any(), db.GetAllVideoUploadedByUserPaginatedParams{
			UserID:     userID,
			PageSize:   int64(pageSize),
			PageNumber: int64(pageNumber),
		}).
		Return(expectedVideos, nil)

	// Injecting fake auth context
	ctx := context.WithValue(context.Background(), auth.AUTH_CONTEXT_KEY, &auth.AuthContext{
		User: &auth.User{ID: userID},
	})

	// Creating VideoAPI instance with mock DB queries
	videoAPI := &api.VideoAPI{
		DBQueries: mockQueries,
	}

	// Setting up request
	req := &proto.ListVideosRequest{
		PageSize:   pageSize,
		PageNumber: pageNumber,
	}

	// Calling the method and checking the response
	resp, err := videoAPI.ListVideos(ctx, req)
	require.NoError(t, err)
	require.Len(t, resp.Videos, 2)

	// Checking each video response
	for i, v := range resp.Videos {
		require.Equal(t, expectedVideos[i].ID, v.Id)
		require.Equal(t, expectedVideos[i].Title, v.Title)
		require.Equal(t, expectedVideos[i].Description, v.Description)
		require.Equal(t, expectedVideos[i].Url, v.Url)
		// Time comparison with tolerance of 1 second
		require.WithinDuration(t, expectedVideos[i].CreatedAt.UTC(), v.CreatedAt.AsTime().UTC(), time.Second)
	}
}
