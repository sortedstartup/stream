package api_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"sortedstartup.com/stream/common/auth"
	"sortedstartup.com/stream/common/interceptors"
	userProto "sortedstartup.com/stream/userservice/proto"
	api "sortedstartup.com/stream/videoservice/api"
	mockApi "sortedstartup.com/stream/videoservice/api/mocks"
	db "sortedstartup.com/stream/videoservice/db"
	dbMock "sortedstartup.com/stream/videoservice/db/mocks"
	proto "sortedstartup.com/stream/videoservice/proto"
)

// createAuthContextWithTenant creates a context with auth and tenant metadata
func createAuthContextWithTenant(tenantID string) context.Context {
	authUser := &auth.AuthContext{
		User: &auth.User{
			ID:    "test-user-id",
			Name:  "Test User",
			Email: "test@example.com",
		},
	}
	baseCtx := context.Background()
	ctxWithAuth := context.WithValue(baseCtx, auth.AUTH_CONTEXT_KEY, authUser)

	// Create metadata with tenant ID
	md := metadata.Pairs("x-tenant-id", tenantID)
	ctxWithMetadata := metadata.NewIncomingContext(ctxWithAuth, md)

	// Now run your TenantInterceptor manually to put tenant ID in context key
	interceptor := interceptors.TenantInterceptor()
	// the interceptor has signature: func(ctx, req, info, handler) (resp, err)
	// we can call handler that just returns ctx so we can get new context with tenant id key

	var newCtx context.Context
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		newCtx = ctx
		return nil, nil
	}

	_, err := interceptor(ctxWithMetadata, nil, nil, handler)
	if err != nil {
		panic("failed to run tenant interceptor in test context setup: " + err.Error())
	}

	return newCtx
}

func TestListVideos(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := dbMock.NewMockDBQuerier(ctrl)
	mockUser := userProto.NewMockUserServiceClient(ctrl)

	api := api.CreateTestVideoAPI()
	api.DbQueries = mockDB
	api.UserServiceClient = mockUser

	tenantID := "tenant-1"
	ctx := createAuthContextWithTenant(tenantID)

	t.Run("Positive GetAllAccessibleVideosByTenantID", func(t *testing.T) {
		mockUser.EXPECT().
			GetTenants(gomock.Any(), gomock.Any()).
			Return(&userProto.GetTenantsResponse{
				TenantUsers: []*userProto.TenantUser{
					{Tenant: &userProto.Tenant{Id: tenantID}},
				},
			}, nil)

		mockDB.EXPECT().
			GetAllAccessibleVideosByTenantID(gomock.Any(), gomock.Any()).
			Return([]db.VideoserviceVideo{
				{
					ID:        "v1",
					Title:     "Video 1",
					Url:       "http://example.com/v1",
					CreatedAt: time.Now(),
				},
			}, nil)

		resp, err := api.ListVideos(ctx, &proto.ListVideosRequest{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(resp.Videos) != 1 {
			t.Errorf("Expected 1 video, got %d", len(resp.Videos))
		}
	})

	t.Run("Positive With ChannelID", func(t *testing.T) {
		mockUser.EXPECT().
			GetTenants(gomock.Any(), gomock.Any()).
			Return(&userProto.GetTenantsResponse{
				TenantUsers: []*userProto.TenantUser{
					{Tenant: &userProto.Tenant{Id: tenantID}},
				},
			}, nil)

		mockDB.EXPECT().
			GetVideosByTenantIDAndChannelID(gomock.Any(), gomock.Any()).
			Return([]db.VideoserviceVideo{
				{
					ID:        "v2",
					Title:     "Video 2",
					Url:       "http://example.com/v2",
					CreatedAt: time.Now(),
				},
			}, nil)

		resp, err := api.ListVideos(ctx, &proto.ListVideosRequest{ChannelId: "ch1"})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(resp.Videos) != 1 {
			t.Errorf("Expected 1 video, got %d", len(resp.Videos))
		}
	})

	t.Run("Negative Missing Auth", func(t *testing.T) {
		_, err := api.ListVideos(context.Background(), &proto.ListVideosRequest{})
		if err == nil {
			t.Errorf("Expected error for missing auth, got nil")
		}
		if st, ok := status.FromError(err); !ok || st.Code() != codes.Unauthenticated {
			t.Errorf("Expected Unauthenticated, got %v", err)
		}
	})

	t.Run("Negative User Not In Tenant", func(t *testing.T) {
		mockUser.EXPECT().
			GetTenants(gomock.Any(), gomock.Any()).
			Return(&userProto.GetTenantsResponse{}, nil)

		_, err := api.ListVideos(ctx, &proto.ListVideosRequest{})
		if err == nil {
			t.Errorf("Expected error for user not in tenant, got nil")
		}
		if st, _ := status.FromError(err); st.Code() != codes.PermissionDenied {
			t.Errorf("Expected PermissionDenied, got %v", err)
		}
	})

	t.Run("Negative DB Error", func(t *testing.T) {
		mockUser.EXPECT().
			GetTenants(gomock.Any(), gomock.Any()).
			Return(&userProto.GetTenantsResponse{
				TenantUsers: []*userProto.TenantUser{
					{Tenant: &userProto.Tenant{Id: tenantID}},
				},
			}, nil)

		mockDB.EXPECT().
			GetAllAccessibleVideosByTenantID(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("db failure"))

		_, err := api.ListVideos(ctx, &proto.ListVideosRequest{})
		if err == nil {
			t.Errorf("Expected DB error, got nil")
		}
		if st, _ := status.FromError(err); st.Code() != codes.Internal {
			t.Errorf("Expected Internal code, got %v", err)
		}
	})
}

func TestGetVideo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := dbMock.NewMockDBQuerier(ctrl)
	mockUser := userProto.NewMockUserServiceClient(ctrl)

	api := api.CreateTestVideoAPI()
	api.DbQueries = mockDB
	api.UserServiceClient = mockUser

	tenantID := "tenant-1"
	videoID := "video-123"
	ctx := createAuthContextWithTenant(tenantID)

	t.Run("Positive_Success", func(t *testing.T) {
		// Mock user tenant validation
		mockUser.EXPECT().
			GetTenants(gomock.Any(), gomock.Any()).
			Return(&userProto.GetTenantsResponse{
				TenantUsers: []*userProto.TenantUser{
					{Tenant: &userProto.Tenant{Id: tenantID}},
				},
			}, nil)

		// Mock DB returning a video
		mockDB.EXPECT().
			GetVideoByVideoIDAndTenantID(gomock.Any(), db.GetVideoByVideoIDAndTenantIDParams{
				ID: videoID,
				TenantID: sql.NullString{
					String: tenantID,
					Valid:  true,
				},
			}).
			Return(db.VideoserviceVideo{
				ID:          videoID,
				Title:       "Test Video",
				Description: "Test Desc",
				Url:         "http://example.com/video.mp4",
				CreatedAt:   time.Now(),
			}, nil)

		resp, err := api.GetVideo(ctx, &proto.GetVideoRequest{VideoId: videoID})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if resp.Id != videoID {
			t.Errorf("Expected video ID %s, got %s", videoID, resp.Id)
		}
	})

	t.Run("Negative_MissingAuth", func(t *testing.T) {
		_, err := api.GetVideo(context.Background(), &proto.GetVideoRequest{VideoId: videoID})
		if err == nil {
			t.Fatalf("Expected error for missing auth, got nil")
		}
		if st, ok := status.FromError(err); !ok || st.Code() != codes.Unauthenticated {
			t.Errorf("Expected Unauthenticated error, got %v", err)
		}
	})

	t.Run("Negative_MissingTenantID", func(t *testing.T) {
		// Context with auth but no tenant ID
		authCtx := createAuthContextWithTenant("")
		_, err := api.GetVideo(authCtx, &proto.GetVideoRequest{VideoId: videoID})
		if err == nil {
			t.Fatalf("Expected error for missing tenant ID, got nil")
		}
		if st, ok := status.FromError(err); !ok || st.Code() != codes.InvalidArgument {
			t.Errorf("Expected InvalidArgument error, got %v", err)
		}
	})

	t.Run("Negative_UserNotInTenant", func(t *testing.T) {
		mockUser.EXPECT().
			GetTenants(gomock.Any(), gomock.Any()).
			Return(&userProto.GetTenantsResponse{}, nil) // empty tenants

		_, err := api.GetVideo(ctx, &proto.GetVideoRequest{VideoId: videoID})
		if err == nil {
			t.Fatalf("Expected error for user not in tenant, got nil")
		}
		if st, ok := status.FromError(err); !ok || st.Code() != codes.PermissionDenied {
			t.Errorf("Expected PermissionDenied error, got %v", err)
		}
	})

	t.Run("Negative_VideoNotFound", func(t *testing.T) {
		mockUser.EXPECT().
			GetTenants(gomock.Any(), gomock.Any()).
			Return(&userProto.GetTenantsResponse{
				TenantUsers: []*userProto.TenantUser{{Tenant: &userProto.Tenant{Id: tenantID}}},
			}, nil)

		mockDB.EXPECT().
			GetVideoByVideoIDAndTenantID(gomock.Any(), gomock.Any()).
			Return(db.VideoserviceVideo{}, sql.ErrNoRows)

		_, err := api.GetVideo(ctx, &proto.GetVideoRequest{VideoId: videoID})
		if err == nil {
			t.Fatalf("Expected error for video not found, got nil")
		}
		if st, ok := status.FromError(err); !ok || st.Code() != codes.NotFound {
			t.Errorf("Expected NotFound error, got %v", err)
		}
	})

	t.Run("Negative_InternalDBError", func(t *testing.T) {
		mockUser.EXPECT().
			GetTenants(gomock.Any(), gomock.Any()).
			Return(&userProto.GetTenantsResponse{
				TenantUsers: []*userProto.TenantUser{{Tenant: &userProto.Tenant{Id: tenantID}}},
			}, nil)

		mockDB.EXPECT().
			GetVideoByVideoIDAndTenantID(gomock.Any(), gomock.Any()).
			Return(db.VideoserviceVideo{}, errors.New("db error"))

		_, err := api.GetVideo(ctx, &proto.GetVideoRequest{VideoId: videoID})
		if err == nil {
			t.Fatalf("Expected error for internal DB error, got nil")
		}
		if st, ok := status.FromError(err); !ok || st.Code() != codes.Internal {
			t.Errorf("Expected Internal error, got %v", err)
		}
	})
}

func TestMoveVideoToChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := dbMock.NewMockDBQuerier(ctrl)
	mockPolicyValidator := mockApi.NewMockPolicyValidator(ctrl)
	mockChannelAPI := mockApi.NewMockChannelAPIInterface(ctrl)

	api := api.CreateTestVideoAPI()
	api.DbQueries = mockDB
	api.PolicyValidator = mockPolicyValidator
	api.ChannelAPI = mockChannelAPI

	ctx := context.Background()
	req := &proto.MoveVideoToChannelRequest{
		VideoId:   "vid-1",
		ChannelId: "ch-1",
	}

	tenantID := "tenant-1"
	authCtx := &auth.AuthContext{
		User: &auth.User{ID: "user-1"},
	}
	video := db.VideoserviceVideo{
		ID:             "vid-1",
		UploadedUserID: "user-1",
		ChannelID:      sql.NullString{String: "old-channel", Valid: true},
	}

	updatedVideo := db.VideoserviceVideo{
		ID:        "vid-1",
		ChannelID: sql.NullString{String: "ch-1", Valid: true},
	}

	t.Run("Success", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(authCtx, tenantID, nil)

		// Return pointer here
		mockPolicyValidator.EXPECT().
			GetAndValidateVideo(ctx, req.VideoId, tenantID).
			Return(&video, nil)

		// Pass pointer to ValidateVideoMovePermissions
		mockPolicyValidator.EXPECT().
			ValidateVideoMovePermissions(ctx, mockChannelAPI, &video, "user-1", tenantID, req.ChannelId).
			Return(nil)

		mockDB.EXPECT().
			UpdateVideoChannel(ctx, gomock.Any()).
			Return(nil)

		mockDB.EXPECT().
			GetVideoByVideoIDAndTenantID(ctx, gomock.Any()).
			Return(updatedVideo, nil)

		mockPolicyValidator.EXPECT().
			ConvertVideoToProto(&updatedVideo).
			Return(&proto.Video{Id: "vid-1"})

		resp, err := api.MoveVideoToChannel(ctx, req)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if resp.Video.Id != "vid-1" {
			t.Errorf("Expected vid-1, got %s", resp.Video.Id)
		}
	})

	t.Run("Fail_ValidateBasicRequest", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(nil, "", status.Error(codes.Unauthenticated, "no auth"))

		_, err := api.MoveVideoToChannel(ctx, req)
		if st, _ := status.FromError(err); st.Code() != codes.Unauthenticated {
			t.Errorf("Expected Unauthenticated, got %v", err)
		}
	})

	t.Run("Fail_GetAndValidateVideo", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(authCtx, tenantID, nil)

		// Return pointer here as well
		mockPolicyValidator.EXPECT().
			GetAndValidateVideo(ctx, req.VideoId, tenantID).
			Return(nil, status.Error(codes.NotFound, "video not found"))

		_, err := api.MoveVideoToChannel(ctx, req)
		if st, _ := status.FromError(err); st.Code() != codes.NotFound {
			t.Errorf("Expected NotFound, got %v", err)
		}
	})

	t.Run("Fail_ValidateVideoMovePermissions", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(authCtx, tenantID, nil)

		// Return pointer here
		mockPolicyValidator.EXPECT().
			GetAndValidateVideo(ctx, req.VideoId, tenantID).
			Return(&video, nil)

		mockPolicyValidator.EXPECT().
			ValidateVideoMovePermissions(ctx, mockChannelAPI, &video, "user-1", tenantID, req.ChannelId).
			Return(status.Error(codes.PermissionDenied, "no permission"))

		_, err := api.MoveVideoToChannel(ctx, req)
		if st, _ := status.FromError(err); st.Code() != codes.PermissionDenied {
			t.Errorf("Expected PermissionDenied, got %v", err)
		}
	})

	t.Run("Fail_UpdateVideoChannel", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(authCtx, tenantID, nil)
		mockPolicyValidator.EXPECT().
			GetAndValidateVideo(ctx, req.VideoId, tenantID).
			Return(&video, nil)
		mockPolicyValidator.EXPECT().
			ValidateVideoMovePermissions(ctx, mockChannelAPI, &video, "user-1", tenantID, req.ChannelId).
			Return(nil)

		mockDB.EXPECT().
			UpdateVideoChannel(ctx, gomock.Any()).
			Return(errors.New("db error"))

		_, err := api.MoveVideoToChannel(ctx, req)
		if st, _ := status.FromError(err); st.Code() != codes.Internal {
			t.Errorf("Expected Internal, got %v", err)
		}
	})

	t.Run("Fail_GetVideoAfterUpdate", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(authCtx, tenantID, nil)
		mockPolicyValidator.EXPECT().
			GetAndValidateVideo(ctx, req.VideoId, tenantID).
			Return(&video, nil)
		mockPolicyValidator.EXPECT().
			ValidateVideoMovePermissions(ctx, mockChannelAPI, &video, "user-1", tenantID, req.ChannelId).
			Return(nil)
		mockDB.EXPECT().
			UpdateVideoChannel(ctx, gomock.Any()).
			Return(nil)
		mockDB.EXPECT().
			GetVideoByVideoIDAndTenantID(ctx, gomock.Any()).
			Return(db.VideoserviceVideo{}, errors.New("db error"))

		_, err := api.MoveVideoToChannel(ctx, req)
		if st, _ := status.FromError(err); st.Code() != codes.Internal {
			t.Errorf("Expected Internal, got %v", err)
		}
	})

	t.Run("Fail_ChannelNotUpdated_TenantLevelVideo", func(t *testing.T) {
		videoWithoutChannel := video
		videoWithoutChannel.ChannelID = sql.NullString{}

		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(authCtx, tenantID, nil)
		mockPolicyValidator.EXPECT().
			GetAndValidateVideo(ctx, req.VideoId, tenantID).
			Return(&videoWithoutChannel, nil)
		mockPolicyValidator.EXPECT().
			ValidateVideoMovePermissions(ctx, mockChannelAPI, &videoWithoutChannel, "user-1", tenantID, req.ChannelId).
			Return(nil)
		mockDB.EXPECT().
			UpdateVideoChannel(ctx, gomock.Any()).
			Return(nil)
		mockDB.EXPECT().
			GetVideoByVideoIDAndTenantID(ctx, gomock.Any()).
			Return(videoWithoutChannel, nil)

		_, err := api.MoveVideoToChannel(ctx, req)
		if st, _ := status.FromError(err); st.Code() != codes.PermissionDenied {
			t.Errorf("Expected PermissionDenied, got %v", err)
		}
	})

	t.Run("Fail_ChannelNotUpdated_ChannelVideo", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(authCtx, tenantID, nil)
		mockPolicyValidator.EXPECT().
			GetAndValidateVideo(ctx, req.VideoId, tenantID).
			Return(&video, nil)
		mockPolicyValidator.EXPECT().
			ValidateVideoMovePermissions(ctx, mockChannelAPI, &video, "user-1", tenantID, req.ChannelId).
			Return(nil)
		mockDB.EXPECT().
			UpdateVideoChannel(ctx, gomock.Any()).
			Return(nil)
		mockDB.EXPECT().
			GetVideoByVideoIDAndTenantID(ctx, gomock.Any()).
			Return(video, nil)

		_, err := api.MoveVideoToChannel(ctx, req)
		if st, _ := status.FromError(err); st.Code() != codes.PermissionDenied {
			t.Errorf("Expected PermissionDenied, got %v", err)
		}
	})
}

func TestRemoveVideoFromChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := dbMock.NewMockDBQuerier(ctrl)
	mockPolicyValidator := mockApi.NewMockPolicyValidator(ctrl)
	mockChannelAPI := mockApi.NewMockChannelAPIInterface(ctrl)

	api := api.CreateTestVideoAPI()
	api.DbQueries = mockDB
	api.PolicyValidator = mockPolicyValidator
	api.ChannelAPI = mockChannelAPI

	ctx := context.Background()
	req := &proto.RemoveVideoFromChannelRequest{
		VideoId: "vid-1",
	}

	tenantID := "tenant-1"
	authCtx := &auth.AuthContext{
		User: &auth.User{ID: "user-1"},
	}

	video := db.VideoserviceVideo{
		ID:        "vid-1",
		ChannelID: sql.NullString{String: "ch-1", Valid: true},
	}

	updatedVideo := db.VideoserviceVideo{
		ID:        "vid-1",
		ChannelID: sql.NullString{Valid: false}, // Channel removed
	}

	t.Run("Success", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(authCtx, tenantID, nil)

		mockPolicyValidator.EXPECT().
			GetAndValidateVideo(ctx, req.VideoId, tenantID).
			Return(&video, nil)

		mockPolicyValidator.EXPECT().
			ValidateVideoRemovalPermissions(ctx, mockChannelAPI, &video, authCtx.User.ID, tenantID).
			Return(nil)

		mockDB.EXPECT().
			RemoveVideoFromChannel(ctx, gomock.AssignableToTypeOf(db.RemoveVideoFromChannelParams{})).
			DoAndReturn(func(ctx context.Context, params db.RemoveVideoFromChannelParams) error {
				if params.VideoID != req.VideoId {
					return errors.New("unexpected VideoID")
				}
				if !params.TenantID.Valid || params.TenantID.String != tenantID {
					return errors.New("unexpected TenantID")
				}
				if params.ChannelID != video.ChannelID {
					return errors.New("unexpected ChannelID")
				}
				return nil
			})

		mockDB.EXPECT().
			GetVideoByVideoIDAndTenantID(ctx, db.GetVideoByVideoIDAndTenantIDParams{
				ID:       req.VideoId,
				TenantID: sql.NullString{String: tenantID, Valid: true},
			}).
			Return(updatedVideo, nil)

		mockPolicyValidator.EXPECT().
			ConvertVideoToProto(&updatedVideo).
			Return(&proto.Video{Id: "vid-1"})

		resp, err := api.RemoveVideoFromChannel(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Message != "Video removed from channel successfully" {
			t.Errorf("unexpected message: %s", resp.Message)
		}
		if resp.Video.Id != "vid-1" {
			t.Errorf("expected video id 'vid-1', got %s", resp.Video.Id)
		}
	})

	t.Run("Fail_ValidateBasicRequest", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(nil, "", status.Error(codes.Unauthenticated, "no auth"))

		_, err := api.RemoveVideoFromChannel(ctx, req)
		if st, _ := status.FromError(err); st.Code() != codes.Unauthenticated {
			t.Errorf("expected Unauthenticated error, got %v", err)
		}
	})

	t.Run("Fail_GetAndValidateVideo", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(authCtx, tenantID, nil)

		mockPolicyValidator.EXPECT().
			GetAndValidateVideo(ctx, req.VideoId, tenantID).
			Return(nil, status.Error(codes.NotFound, "video not found"))

		_, err := api.RemoveVideoFromChannel(ctx, req)
		if st, _ := status.FromError(err); st.Code() != codes.NotFound {
			t.Errorf("expected NotFound error, got %v", err)
		}
	})

	t.Run("Fail_ValidateVideoRemovalPermissions", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(authCtx, tenantID, nil)

		mockPolicyValidator.EXPECT().
			GetAndValidateVideo(ctx, req.VideoId, tenantID).
			Return(&video, nil)

		mockPolicyValidator.EXPECT().
			ValidateVideoRemovalPermissions(ctx, mockChannelAPI, &video, authCtx.User.ID, tenantID).
			Return(status.Error(codes.PermissionDenied, "no permission"))

		_, err := api.RemoveVideoFromChannel(ctx, req)
		if st, _ := status.FromError(err); st.Code() != codes.PermissionDenied {
			t.Errorf("expected PermissionDenied error, got %v", err)
		}
	})

	t.Run("Fail_RemoveVideoFromChannel_DBError", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(authCtx, tenantID, nil)
		mockPolicyValidator.EXPECT().
			GetAndValidateVideo(ctx, req.VideoId, tenantID).
			Return(&video, nil)
		mockPolicyValidator.EXPECT().
			ValidateVideoRemovalPermissions(ctx, mockChannelAPI, &video, authCtx.User.ID, tenantID).
			Return(nil)

		mockDB.EXPECT().
			RemoveVideoFromChannel(ctx, gomock.Any()).
			Return(errors.New("db error"))

		_, err := api.RemoveVideoFromChannel(ctx, req)
		if st, _ := status.FromError(err); st.Code() != codes.Internal {
			t.Errorf("expected Internal error, got %v", err)
		}
	})

	t.Run("Fail_GetUpdatedVideo_DBError", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(authCtx, tenantID, nil)
		mockPolicyValidator.EXPECT().
			GetAndValidateVideo(ctx, req.VideoId, tenantID).
			Return(&video, nil)
		mockPolicyValidator.EXPECT().
			ValidateVideoRemovalPermissions(ctx, mockChannelAPI, &video, authCtx.User.ID, tenantID).
			Return(nil)

		mockDB.EXPECT().
			RemoveVideoFromChannel(ctx, gomock.Any()).
			Return(nil)

		mockDB.EXPECT().
			GetVideoByVideoIDAndTenantID(ctx, gomock.Any()).
			Return(db.VideoserviceVideo{}, errors.New("db error"))

		_, err := api.RemoveVideoFromChannel(ctx, req)
		if st, _ := status.FromError(err); st.Code() != codes.Internal {
			t.Errorf("expected Internal error, got %v", err)
		}
	})
}

func TestDeleteVideo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := dbMock.NewMockDBQuerier(ctrl)
	mockPolicyValidator := mockApi.NewMockPolicyValidator(ctrl)
	mockChannelAPI := mockApi.NewMockChannelAPIInterface(ctrl)

	api := api.CreateTestVideoAPI()
	api.DbQueries = mockDB
	api.PolicyValidator = mockPolicyValidator
	api.ChannelAPI = mockChannelAPI

	ctx := context.Background()
	req := &proto.DeleteVideoRequest{
		VideoId: "vid-1",
	}

	tenantID := "tenant-1"
	authCtx := &auth.AuthContext{
		User: &auth.User{ID: "user-1"},
	}

	video := db.VideoserviceVideo{
		ID:             "vid-1",
		UploadedUserID: "user-1",
		ChannelID:      sql.NullString{String: "ch-1", Valid: true},
	}

		t.Run("Success", func(t *testing.T) {
			mockPolicyValidator.EXPECT().
				ValidateBasicRequest(ctx).
				Return(authCtx, tenantID, nil)

			mockPolicyValidator.EXPECT().
				GetAndValidateVideo(ctx, req.VideoId, tenantID).
				Return(&video, nil) 

			mockPolicyValidator.EXPECT().
				ValidateVideoDeletionPermissions(ctx, mockChannelAPI, &video, authCtx.User.ID, tenantID).
				Return(nil)

			mockDB.EXPECT().
				SoftDeleteVideo(ctx, gomock.AssignableToTypeOf(db.SoftDeleteVideoParams{})).
				DoAndReturn(func(ctx context.Context, params db.SoftDeleteVideoParams) error {
					if params.VideoID != req.VideoId {
						return errors.New("unexpected VideoID")
					}
					if !params.TenantID.Valid || params.TenantID.String != tenantID {
						return errors.New("unexpected TenantID")
					}
					return nil
				})

			resp, err := api.DeleteVideo(ctx, req)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if resp.Message != "Video deleted successfully" {
				t.Errorf("Expected success message, got %s", resp.Message)
			}
		})

	t.Run("Fail_ValidateBasicRequest", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(nil, "", status.Error(codes.Unauthenticated, "no auth"))

		_, err := api.DeleteVideo(ctx, req)
		if st, _ := status.FromError(err); st.Code() != codes.Unauthenticated {
			t.Errorf("Expected Unauthenticated, got %v", err)
		}
	})

	t.Run("Fail_GetAndValidateVideo", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(authCtx, tenantID, nil)

		mockPolicyValidator.EXPECT().
			GetAndValidateVideo(ctx, req.VideoId, tenantID).
			Return(nil, status.Error(codes.NotFound, "video not found")) // Return nil pointer + error

		_, err := api.DeleteVideo(ctx, req)
		if st, _ := status.FromError(err); st.Code() != codes.NotFound {
			t.Errorf("Expected NotFound, got %v", err)
		}
	})

	t.Run("Fail_ValidateVideoDeletionPermissions", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(authCtx, tenantID, nil)

		mockPolicyValidator.EXPECT().
			GetAndValidateVideo(ctx, req.VideoId, tenantID).
			Return(&video, nil)

		mockPolicyValidator.EXPECT().
			ValidateVideoDeletionPermissions(ctx, mockChannelAPI, &video, authCtx.User.ID, tenantID).
			Return(status.Error(codes.PermissionDenied, "no permission"))

		_, err := api.DeleteVideo(ctx, req)
		if st, _ := status.FromError(err); st.Code() != codes.PermissionDenied {
			t.Errorf("Expected PermissionDenied, got %v", err)
		}
	})

	t.Run("Fail_SoftDeleteVideo", func(t *testing.T) {
		mockPolicyValidator.EXPECT().
			ValidateBasicRequest(ctx).
			Return(authCtx, tenantID, nil)

		mockPolicyValidator.EXPECT().
			GetAndValidateVideo(ctx, req.VideoId, tenantID).
			Return(&video, nil)

		mockPolicyValidator.EXPECT().
			ValidateVideoDeletionPermissions(ctx, mockChannelAPI, &video, authCtx.User.ID, tenantID).
			Return(nil)

		mockDB.EXPECT().
			SoftDeleteVideo(ctx, gomock.AssignableToTypeOf(db.SoftDeleteVideoParams{})).
			Return(errors.New("db error"))

		_, err := api.DeleteVideo(ctx, req)
		if st, _ := status.FromError(err); st.Code() != codes.Internal {
			t.Errorf("Expected Internal error, got %v", err)
		}
	})
}

func TestValidateChannelRole(t *testing.T) {
	chAPI := &api.ChannelAPI{}

	validRoles := []string{"owner", "uploader", "viewer"}
	for _, role := range validRoles {
		t.Run("ValidRole_"+role, func(t *testing.T) {
			err := chAPI.ValidateChannelRole(role)
			if err != nil {
				t.Errorf("Expected no error for valid role %q, got %v", role, err)
			}
		})
	}

	invalidRoles := []string{"admin", "guest", "", "Owner", "UPLOADER", "viewer1", "user", " "}

	for _, role := range invalidRoles {
		t.Run("InvalidRole_"+role, func(t *testing.T) {
			err := chAPI.ValidateChannelRole(role)
			if err == nil {
				t.Errorf("Expected error for invalid role %q, got nil", role)
				return
			}
			st, ok := status.FromError(err)
			if !ok {
				t.Errorf("Expected gRPC status error for invalid role %q, got %v", role, err)
				return
			}
			if st.Code() != codes.InvalidArgument {
				t.Errorf("Expected InvalidArgument code for invalid role %q, got %v", role, st.Code())
			}
			expectedMsg := "invalid role. Valid roles are: owner, uploader, viewer"
			if st.Message() != expectedMsg {
				t.Errorf("Expected error message %q, got %q", expectedMsg, st.Message())
			}
		})
	}
}

func TestGetUserRoleInChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockApi.NewMockChannelDB(ctrl)

	api := &api.ChannelAPI{
		DbQueries: mockDB,
	}

	ctx := context.Background()
	channelID := "ch-1"
	userID := "user-1"
	tenantID := "tenant-1"

	t.Run("Success - user is member", func(t *testing.T) {
		mockDB.EXPECT().
			GetUserRoleInChannel(ctx, db.GetUserRoleInChannelParams{
				ChannelID: channelID,
				UserID:    userID,
				TenantID:  tenantID,
			}).
			Return("uploader", nil)

		role, err := api.GetUserRoleInChannel(ctx, channelID, userID, tenantID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if role != "uploader" {
			t.Errorf("expected role 'uploader', got %s", role)
		}
	})

	t.Run("Fail - user not a member (sql.ErrNoRows)", func(t *testing.T) {
		mockDB.EXPECT().
			GetUserRoleInChannel(ctx, db.GetUserRoleInChannelParams{
				ChannelID: channelID,
				UserID:    userID,
				TenantID:  tenantID,
			}).
			Return("", sql.ErrNoRows)

		_, err := api.GetUserRoleInChannel(ctx, channelID, userID, tenantID)
		st, _ := status.FromError(err)
		if st.Code() != codes.PermissionDenied {
			t.Errorf("expected PermissionDenied, got %v", err)
		}
	})

	t.Run("Fail - db error", func(t *testing.T) {
		mockDB.EXPECT().
			GetUserRoleInChannel(ctx, db.GetUserRoleInChannelParams{
				ChannelID: channelID,
				UserID:    userID,
				TenantID:  tenantID,
			}).
			Return("", errors.New("db error"))

		_, err := api.GetUserRoleInChannel(ctx, channelID, userID, tenantID)
		st, _ := status.FromError(err)
		if st.Code() != codes.Internal {
			t.Errorf("expected Internal error, got %v", err)
		}
	})
}

func TestGetChannelMemberCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockApi.NewMockChannelDB(ctrl)
	api := &api.ChannelAPI{
		DbQueries: mockDB,
	}

	ctx := context.Background()
	channelID := "ch-1"
	tenantID := "tenant-1"

	t.Run("Success - multiple members", func(t *testing.T) {
		mockDB.EXPECT().
			GetChannelMembersByChannelIDAndTenantID(ctx, db.GetChannelMembersByChannelIDAndTenantIDParams{
				ChannelID: channelID,
				TenantID:  tenantID,
			}).
			Return([]db.GetChannelMembersByChannelIDAndTenantIDRow{
				{UserID: "user1", Role: "owner"},
				{UserID: "user2", Role: "viewer"},
				{UserID: "user3", Role: "uploader"},
			}, nil)

		count, err := api.GetChannelMemberCount(ctx, channelID, tenantID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 3 {
			t.Errorf("expected count 3, got %d", count)
		}
	})

	t.Run("Success - zero members", func(t *testing.T) {
		mockDB.EXPECT().
			GetChannelMembersByChannelIDAndTenantID(ctx, db.GetChannelMembersByChannelIDAndTenantIDParams{
				ChannelID: channelID,
				TenantID:  tenantID,
			}).
			Return([]db.GetChannelMembersByChannelIDAndTenantIDRow{}, nil)

		count, err := api.GetChannelMemberCount(ctx, channelID, tenantID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 0 {
			t.Errorf("expected count 0, got %d", count)
		}
	})

	t.Run("Fail - database error", func(t *testing.T) {
		mockDB.EXPECT().
			GetChannelMembersByChannelIDAndTenantID(ctx, db.GetChannelMembersByChannelIDAndTenantIDParams{
				ChannelID: channelID,
				TenantID:  tenantID,
			}).
			Return(nil, errors.New("db error"))

		_, err := api.GetChannelMemberCount(ctx, channelID, tenantID)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestCreateChannel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockApi.NewMockChannelDB(ctrl)
	mockUser := userProto.NewMockUserServiceClient(ctrl)

	channelAPI := &api.ChannelAPI{
		DbQueries:         mockDB,
		UserServiceClient: mockUser,
		Log: slog.New(
        slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}),
    	),
	}

	t.Run("Positive - personal tenant", func(t *testing.T) {
		userID := "user-123"
		tenantID := "tenant-456"

		// Authenticated context with tenant ID
		ctx := context.Background()
		ctx = context.WithValue(ctx, interceptors.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, auth.AUTH_CONTEXT_KEY, &auth.AuthContext{
			User: &auth.User{
				ID:    userID,
				Name:  "Test User",
				Email: "test@example.com",
				Roles: []auth.Role{"admin"},
			},
			IsAuthenticated: true,
		})

		// Mock GetTenants response
		mockUser.EXPECT().
			GetTenants(gomock.Any(), gomock.Any()).
			Return(&userProto.GetTenantsResponse{
				TenantUsers: []*userProto.TenantUser{
					{
						Tenant: &userProto.Tenant{
							Id:        tenantID,
							IsPersonal: true, // <- mark tenant as personal
						},
						User:   &userProto.User{Id: userID},
						Role:   &userProto.Role{Role: "admin"},
					},
				},
			}, nil)

		// Mock CreateChannel
		mockDB.EXPECT().
			CreateChannel(gomock.Any(), gomock.Any()).
			Return(db.VideoserviceChannel{
				ID:       "channel-789",
				TenantID: tenantID,
				Name:     "My Test Channel",
				CreatedBy: userID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil)

		// Mock CreateChannelMember
		mockDB.EXPECT().
			CreateChannelMember(gomock.Any(), gomock.Any()).
			Return(db.VideoserviceChannelMember{
				ID:        "member-001",
				ChannelID: "channel-789",
				UserID:    userID,
				Role:      "owner",
				AddedBy:   userID,
				CreatedAt: time.Now(),
			}, nil)

		// Run API call
		req := &proto.CreateChannelRequest{
			Name:        "My Test Channel",
			Description: "Test description",
		}

		resp, err := channelAPI.CreateChannel(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Channel.Id != "channel-789" {
			t.Errorf("expected channel ID 'channel-789', got %v", resp.Channel.Id)
		}
		if resp.Channel.UserRole != "owner" {
			t.Errorf("expected user role 'owner', got %v", resp.Channel.UserRole)
		}
		if resp.Channel.MemberCount != 1 {
			t.Errorf("expected member count 1, got %v", resp.Channel.MemberCount)
		}
	})

	t.Run("Negative - missing auth", func(t *testing.T) {
		ctx := context.Background() // no auth in context
		req := &proto.CreateChannelRequest{Name: "Test"}

		_, err := channelAPI.CreateChannel(ctx, req)
		if status.Code(err) != codes.Unauthenticated {
			t.Errorf("expected Unauthenticated error, got %v", err)
		}
	})

	t.Run("Negative - missing tenant ID", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), auth.AUTH_CONTEXT_KEY, &auth.AuthContext{
			User: &auth.User{
				ID: "user-123",
			},
			IsAuthenticated: true,
		})
		req := &proto.CreateChannelRequest{Name: "Test"}

		_, err := channelAPI.CreateChannel(ctx, req)
		if status.Code(err) != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument error, got %v", err)
		}
	})

	t.Run("Negative - org tenant, insufficient role", func(t *testing.T) {
		userID := "user-123"
		tenantID := "tenant-456"

		ctx := context.Background()
		ctx = context.WithValue(ctx, interceptors.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, auth.AUTH_CONTEXT_KEY, &auth.AuthContext{
			User: &auth.User{ID: userID},
			IsAuthenticated: true,
		})

		// Mock GetTenants for organizational tenant
		mockUser.EXPECT().
			GetTenants(gomock.Any(), gomock.Any()).
			Return(&userProto.GetTenantsResponse{
				TenantUsers: []*userProto.TenantUser{
					{
						Tenant: &userProto.Tenant{Id: tenantID, IsPersonal: false},
						User:   &userProto.User{Id: userID},
						Role:   &userProto.Role{Role: "member"}, // not super admin
					},
				},
			}, nil)

		req := &proto.CreateChannelRequest{Name: "Test"}

		_, err := channelAPI.CreateChannel(ctx, req)
		if status.Code(err) != codes.PermissionDenied {
			t.Errorf("expected PermissionDenied error, got %v", err)
		}
	})

	t.Run("Negative - DB CreateChannel fails", func(t *testing.T) {
		userID := "user-123"
		tenantID := "tenant-456"

		ctx := context.Background()
		ctx = context.WithValue(ctx, interceptors.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, auth.AUTH_CONTEXT_KEY, &auth.AuthContext{
			User: &auth.User{ID: userID},
			IsAuthenticated: true,
		})

		// Mock GetTenants
		mockUser.EXPECT().
			GetTenants(gomock.Any(), gomock.Any()).
			Return(&userProto.GetTenantsResponse{
				TenantUsers: []*userProto.TenantUser{
					{
						Tenant: &userProto.Tenant{Id: tenantID, IsPersonal: true},
						User:   &userProto.User{Id: userID},
						Role:   &userProto.Role{Role: "admin"},
					},
				},
			}, nil)

		// Simulate DB failure
		mockDB.EXPECT().
			CreateChannel(gomock.Any(), gomock.Any()).
			Return(db.VideoserviceChannel{}, fmt.Errorf("db error"))

		req := &proto.CreateChannelRequest{
			Name: "Test Channel",
		}

		_, err := channelAPI.CreateChannel(ctx, req)
		if status.Code(err) != codes.Internal {
			t.Errorf("expected Internal error, got %v", err)
		}
	})

}

func TestGetChannels(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockApi.NewMockChannelDB(ctrl)
	mockUser := userProto.NewMockUserServiceClient(ctrl)

	// Create ChannelAPI with a test logger to avoid nil pointer panics
	channelAPI := &api.ChannelAPI{
		DbQueries:         mockDB,
		UserServiceClient: mockUser,
		Log: slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
	}

	userID := "user-123"
	tenantID := "tenant-456"

	// common authenticated context
	authCtx := &auth.AuthContext{
		User: &auth.User{
			ID:    userID,
			Name:  "Test User",
			Email: "test@example.com",
			Roles: []auth.Role{"admin"},
		},
		IsAuthenticated: true,
	}

	t.Run("Positive - multiple channels with membership", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, interceptors.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, auth.AUTH_CONTEXT_KEY, authCtx)

		channels := []db.VideoserviceChannel{
			{
				ID:        "channel-1",
				TenantID:  tenantID,
				Name:      "Channel One",
				CreatedBy: userID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:        "channel-2",
				TenantID:  tenantID,
				Name:      "Channel Two",
				CreatedBy: "other-user",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		mockUser.EXPECT().
			GetTenants(gomock.Any(), gomock.Any()).
			Return(&userProto.GetTenantsResponse{
				TenantUsers: []*userProto.TenantUser{
					{
						Tenant: &userProto.Tenant{
							Id:       tenantID,
							IsPersonal: true,
						},
						User: &userProto.User{Id: userID},
						Role: &userProto.Role{Role: "admin"},
					},
				},
			}, nil).AnyTimes()

		// Mock DB call
		mockDB.EXPECT().GetChannelsByTenantID(gomock.Any(), tenantID).Return(channels, nil)

		// Mock GetUserRoleInChannel: user is owner of channel-1, not a member of channel-2
		mockDB.EXPECT().GetUserRoleInChannel(gomock.Any(), db.GetUserRoleInChannelParams{
			ChannelID: "channel-1",
			UserID:    userID,
			TenantID:  tenantID,
		}).Return("owner", nil)
		mockDB.EXPECT().GetChannelMembersByChannelIDAndTenantID(gomock.Any(), db.GetChannelMembersByChannelIDAndTenantIDParams{
			ChannelID: "channel-1",
			TenantID:  tenantID,
		}).Return([]db.GetChannelMembersByChannelIDAndTenantIDRow{
			{UserID: userID, Role: "owner"},
			{UserID: "user-2", Role: "uploader"},
		}, nil)

		// For channel-2, return error to simulate not a member
		mockDB.EXPECT().GetUserRoleInChannel(gomock.Any(), db.GetUserRoleInChannelParams{
			ChannelID: "channel-2",
			UserID:    userID,
			TenantID:  tenantID,
		}).Return("", fmt.Errorf("user not in channel"))

		resp, err := channelAPI.GetChannels(ctx, &proto.GetChannelsRequest{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.Channels) != 1 {
			t.Errorf("expected 1 channel, got %d", len(resp.Channels))
		}
		c := resp.Channels[0]
		if c.Id != "channel-1" || c.UserRole != "owner" || c.MemberCount != 2 {
			t.Errorf("unexpected channel data: %+v", c)
		}
	})

	t.Run("Negative - missing auth", func(t *testing.T) {
		ctx := context.Background() // no auth context
		ctx = context.WithValue(ctx, interceptors.TenantIDKey, tenantID)

		_, err := channelAPI.GetChannels(ctx, &proto.GetChannelsRequest{})
		if status.Code(err) != codes.Unauthenticated {
			t.Errorf("expected Unauthenticated, got %v", err)
		}
	})

	t.Run("Negative - missing tenant ID", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, auth.AUTH_CONTEXT_KEY, authCtx)

		_, err := channelAPI.GetChannels(ctx, &proto.GetChannelsRequest{})
		if status.Code(err) != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", err)
		}
	})

	t.Run("Negative - DB GetChannelsByTenantID fails", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, interceptors.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, auth.AUTH_CONTEXT_KEY, authCtx)

		mockDB.EXPECT().GetChannelsByTenantID(gomock.Any(), tenantID).Return(nil, fmt.Errorf("db error"))

		_, err := channelAPI.GetChannels(ctx, &proto.GetChannelsRequest{})
		if status.Code(err) != codes.Internal {
			t.Errorf("expected Internal error, got %v", err)
		}
	})

	t.Run("Negative - GetUserRoleInChannel fails for a channel", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, interceptors.TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, auth.AUTH_CONTEXT_KEY, authCtx)

		channels := []db.VideoserviceChannel{
			{ID: "channel-1", TenantID: tenantID, Name: "Channel One"},
		}

		mockDB.EXPECT().GetChannelsByTenantID(gomock.Any(), tenantID).Return(channels, nil)
		mockDB.EXPECT().GetUserRoleInChannel(gomock.Any(), db.GetUserRoleInChannelParams{
			ChannelID: "channel-1",
			UserID:    userID,
			TenantID:  tenantID,
		}).Return("", fmt.Errorf("db error"))

		resp, err := channelAPI.GetChannels(ctx, &proto.GetChannelsRequest{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Channels) != 0 {
			t.Errorf("expected 0 channels, got %d", len(resp.Channels))
		}
	})
}
