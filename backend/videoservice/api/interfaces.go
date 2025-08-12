package api

import (
	"context"
	"sortedstartup.com/stream/common/auth"
	"sortedstartup.com/stream/videoservice/db"
	"sortedstartup.com/stream/videoservice/proto"
)

// PolicyValidator defines video-related permission and validation operations.
type PolicyValidator interface {
	ValidateBasicRequest(ctx context.Context) (authContext *auth.AuthContext, tenantID string, err error)
	GetAndValidateVideo(ctx context.Context, videoID, tenantID string) (*db.VideoserviceVideo, error)
	ValidateChannelOwnership(ctx context.Context, channelAPI ChannelAPIInterface, channelID, userID, tenantID string) error
	ValidateChannelAccess(ctx context.Context, channelAPI ChannelAPIInterface, channelID, userID, tenantID string, requiredRoles ...string) (string, error)
	ValidateVideoMovePermissions(ctx context.Context, channelAPI ChannelAPIInterface, video *db.VideoserviceVideo, userID, tenantID, targetChannelID string) error
	ValidateVideoRemovalPermissions(ctx context.Context, channelAPI ChannelAPIInterface, video *db.VideoserviceVideo, userID, tenantID string) error
	ValidateVideoDeletionPermissions(ctx context.Context, channelAPI ChannelAPIInterface, video *db.VideoserviceVideo, userID, tenantID string) error
	ConvertVideoToProto(video *db.VideoserviceVideo) *proto.Video
}

type ChannelAPIInterface interface {
	GetUserRoleInChannel(ctx context.Context, channelID, userID, tenantID string) (string, error)
}
