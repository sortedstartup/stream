package api

import (
	"context"
	"database/sql"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"sortedstartup.com/stream/common/auth"
	"sortedstartup.com/stream/common/constants"
	"sortedstartup.com/stream/common/interceptors"
	userProto "sortedstartup.com/stream/userservice/proto"
	"sortedstartup.com/stream/videoservice/db"
	proto "sortedstartup.com/stream/videoservice/proto"
)

// VideoPolicyValidator handles video-related permission checks and validations
type VideoPolicyValidator struct {
	dbQueries         *db.Queries
	userServiceClient userProto.UserServiceClient
	log               *slog.Logger
}

// NewVideoPolicyValidator creates a new video policy validator
func NewVideoPolicyValidator(dbQueries *db.Queries, userServiceClient userProto.UserServiceClient, log *slog.Logger) *VideoPolicyValidator {
	return &VideoPolicyValidator{
		dbQueries:         dbQueries,
		userServiceClient: userServiceClient,
		log:               log,
	}
}

// ValidateBasicRequest performs common validation steps for video requests
func (v *VideoPolicyValidator) ValidateBasicRequest(ctx context.Context) (authContext *auth.AuthContext, tenantID string, err error) {
	// Get auth context
	authContext, err = interceptors.AuthFromContext(ctx)
	if err != nil {
		v.log.Error("Error getting auth from context", "err", err)
		return nil, "", err
	}

	// Get tenant ID from headers/metadata
	tenantID, err = interceptors.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, "", status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	// Validate user has access to this tenant
	err = isUserInTenant(ctx, v.userServiceClient, v.log, tenantID, authContext.User.ID)
	if err != nil {
		return nil, "", err
	}

	return authContext, tenantID, nil
}

// GetAndValidateVideo retrieves video and performs basic validation
func (v *VideoPolicyValidator) GetAndValidateVideo(ctx context.Context, videoID, tenantID string) (*db.VideoserviceVideo, error) {
	video, err := v.dbQueries.GetVideoByVideoIDAndTenantID(ctx, db.GetVideoByVideoIDAndTenantIDParams{
		ID:       videoID,
		TenantID: sql.NullString{String: tenantID, Valid: true},
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "video not found")
		}
		v.log.Error("Error getting video", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &video, nil
}

// ValidateChannelOwnership checks if user is the owner of the specified channel
func (v *VideoPolicyValidator) ValidateChannelOwnership(ctx context.Context, channelAPI *ChannelAPI, channelID, userID, tenantID string) error {
	userRole, err := channelAPI.getUserRoleInChannel(ctx, channelID, userID, tenantID)
	if err != nil {
		return status.Error(codes.PermissionDenied, "access denied: you are not a member of this channel")
	}

	if userRole != constants.ChannelRoleOwner {
		return status.Error(codes.PermissionDenied, "access denied: only channel owners can perform this action")
	}

	return nil
}

// ValidateChannelAccess checks if user has required access level to the channel
func (v *VideoPolicyValidator) ValidateChannelAccess(ctx context.Context, channelAPI *ChannelAPI, channelID, userID, tenantID string, requiredRoles ...string) (string, error) {
	userRole, err := channelAPI.getUserRoleInChannel(ctx, channelID, userID, tenantID)
	if err != nil {
		return "", status.Error(codes.PermissionDenied, "access denied: you are not a member of this channel")
	}

	// Check if user has any of the required roles
	hasRequiredRole := false
	for _, role := range requiredRoles {
		if userRole == role {
			hasRequiredRole = true
			break
		}
	}

	if !hasRequiredRole {
		return "", status.Error(codes.PermissionDenied, "access denied: insufficient permissions for this channel")
	}

	return userRole, nil
}

// ValidateVideoMovePermissions checks permissions for moving a video
func (v *VideoPolicyValidator) ValidateVideoMovePermissions(ctx context.Context, channelAPI *ChannelAPI, video *db.VideoserviceVideo, userID, tenantID, targetChannelID string) error {
	// Validate that target channel exists and user has uploader+ access
	_, err := v.ValidateChannelAccess(ctx, channelAPI, targetChannelID, userID, tenantID,
		constants.ChannelRoleOwner, constants.ChannelRoleUploader)
	if err != nil {
		return status.Error(codes.PermissionDenied, "access denied: you need uploader or owner access to add videos to the target channel")
	}

	// Permission validation for moving videos:
	// 1. For tenant-level videos: User must be the uploader (validated in database)
	// 2. For channel videos: User must be the source channel owner
	if video.ChannelID.Valid && video.ChannelID.String != "" {
		// Video is currently in a channel - validate source channel ownership
		err := v.ValidateChannelOwnership(ctx, channelAPI, video.ChannelID.String, userID, tenantID)
		if err != nil {
			return status.Error(codes.PermissionDenied, "access denied: only the source channel owner can move videos between channels")
		}
	}
	// Note: For tenant-level videos, uploader ownership is validated in the database query

	return nil
}

// ValidateVideoRemovalPermissions checks permissions for removing a video from channel
func (v *VideoPolicyValidator) ValidateVideoRemovalPermissions(ctx context.Context, channelAPI *ChannelAPI, video *db.VideoserviceVideo, userID, tenantID string) error {
	// Check if video is currently in a channel
	if !video.ChannelID.Valid || video.ChannelID.String == "" {
		return status.Error(codes.InvalidArgument, "video is not in any channel")
	}

	// Validate that user is the owner of the channel
	return v.ValidateChannelOwnership(ctx, channelAPI, video.ChannelID.String, userID, tenantID)
}

// ValidateVideoDeletionPermissions checks permissions for deleting a video
func (v *VideoPolicyValidator) ValidateVideoDeletionPermissions(ctx context.Context, channelAPI *ChannelAPI, video *db.VideoserviceVideo, userID, tenantID string) error {
	// Permission check for video deletion:
	// 1. For tenant-level videos: Only video uploader can delete
	// 2. For channel videos: Only channel owner can delete
	if video.ChannelID.Valid && video.ChannelID.String != "" {
		// Video is in a channel - only channel owner can delete it
		err := v.ValidateChannelOwnership(ctx, channelAPI, video.ChannelID.String, userID, tenantID)
		if err != nil {
			return status.Error(codes.PermissionDenied, "access denied: only channel owners can delete videos from channels")
		}
	} else {
		// Video is at tenant-level - only uploader can delete their own videos
		if video.UploadedUserID != userID {
			return status.Error(codes.PermissionDenied, "access denied: you can only delete your own tenant-level videos")
		}
	}

	return nil
}

// ConvertVideoToProto converts a database video to proto format
func (v *VideoPolicyValidator) ConvertVideoToProto(video *db.VideoserviceVideo) *proto.Video {
	return &proto.Video{
		Id:          video.ID,
		Title:       video.Title,
		Description: video.Description,
		Url:         video.Url,
		ChannelId:   video.ChannelID.String,
		Visibility:  proto.Visibility_VISIBILITY_PRIVATE,
		CreatedAt:   timestamppb.New(video.CreatedAt),
	}
}
