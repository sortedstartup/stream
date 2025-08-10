package api

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/metadata"

	userProto "sortedstartup.com/stream/userservice/proto"
	db "sortedstartup.com/stream/videoservice/db"
	dbMock "sortedstartup.com/stream/videoservice/db/mocks"
	"sortedstartup.com/stream/videoservice/proto"
	"sortedstartup.com/stream/common/auth"
	"sortedstartup.com/stream/common/interceptors"
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

	api := createTestVideoAPI()
	api.dbQueries = mockDB
	api.userServiceClient = mockUser

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