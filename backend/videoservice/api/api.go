package api

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	_ "modernc.org/sqlite"
	"sortedstartup.com/stream/common/auth"
	"sortedstartup.com/stream/common/constants"
	"sortedstartup.com/stream/common/interceptors"
	userProto "sortedstartup.com/stream/userservice/proto"
	"sortedstartup.com/stream/videoservice/config"
	"sortedstartup.com/stream/videoservice/db"
	"sortedstartup.com/stream/videoservice/proto"
)

type VideoAPI struct {
	config        config.VideoServiceConfig
	HTTPServerMux *http.ServeMux
	db            *sql.DB

	log       *slog.Logger
	DbQueries  db.DBQuerier 

	// gRPC clients for other services
	UserServiceClient userProto.UserServiceClient

	// Policy validator for common video operations
	PolicyValidator PolicyValidator

	//implemented proto server
	proto.UnimplementedVideoServiceServer
	ChannelAPI ChannelAPIInterface    
}

type ChannelAPI struct {
	config        config.VideoServiceConfig
	HTTPServerMux *http.ServeMux
	db            *sql.DB

	Log       *slog.Logger
	DbQueries ChannelDB

	// gRPC clients for other services
	UserServiceClient   userProto.UserServiceClient
	tenantServiceClient userProto.TenantServiceClient

	//implemented proto server
	proto.UnimplementedChannelServiceServer
}

func NewVideoAPIProduction(config config.VideoServiceConfig, userServiceClient userProto.UserServiceClient, tenantServiceClient userProto.TenantServiceClient) (*VideoAPI, *ChannelAPI, error) {
	slog.Info("NewVideoAPIProduction")

	fbAuth, err := auth.NewFirebase()
	if err != nil {
		return nil, nil, err
	}

	childLogger := slog.With("service", "VideoAPI")

	_db, err := sql.Open(config.DB.Driver, config.DB.Url)
	if err != nil {
		return nil, nil, err
	}

	dbQueries := db.New(_db)

	ServerMux := http.NewServeMux()

	channelAPI := &ChannelAPI{
		HTTPServerMux:       ServerMux,
		config:              config,
		db:                  _db,
		Log:                 childLogger,
		DbQueries:           dbQueries,
		UserServiceClient:   userServiceClient,
		tenantServiceClient: tenantServiceClient,
	}

	// Create policy validator
	policyValidator := NewVideoPolicyValidator(dbQueries, userServiceClient, childLogger)

	videoAPI := &VideoAPI{
		HTTPServerMux:     ServerMux,
		config:            config,
		db:                _db,
		log:               childLogger,
		DbQueries:         dbQueries,
		UserServiceClient: userServiceClient,
		PolicyValidator:   policyValidator,
		ChannelAPI:        channelAPI,
	}

	// The authentication is handled in mono/main.go
	ServerMux.Handle("/upload", interceptors.FirebaseHTTPHeaderAuthMiddleware(fbAuth, http.HandlerFunc(videoAPI.uploadHandler)))
	//the cookie auth middleware is just to allow if the user is logged in
	ServerMux.Handle("/video/", interceptors.FirebaseCookieAuthMiddleware(fbAuth, http.HandlerFunc(videoAPI.serveVideoHandler)))

	return videoAPI, channelAPI, nil
}

func (s *VideoAPI) Start() error {
	return nil
}

func (s *VideoAPI) Init() error {
	s.log.Info("Migrating database", "dbDriver", s.config.DB.Driver, "dbURL", s.config.DB.Url)
	err := db.MigrateDB(s.config.DB.Driver, s.config.DB.Url)
	if err != nil {
		return err
	}
	s.log.Info("Migrating database done")
	return nil
}

// ===== SHARED HELPER FUNCTIONS =====

// isUserInTenant checks if the user is part of the specified tenant
// by calling the userservice to get user's tenants and checking if the tenant is in the list
func isUserInTenant(ctx context.Context, userServiceClient userProto.UserServiceClient, log *slog.Logger, tenantID, userID string) error {
	if tenantID == "" {
		return status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	// Call userservice to get user's tenants
	resp, err := userServiceClient.GetTenants(ctx, &userProto.GetTenantsRequest{})
	if err != nil {
		log.Error("Failed to get user tenants from userservice", "error", err, "userID", userID)
		return status.Error(codes.Internal, "failed to check tenant access")
	}

	// Check if the requested tenant is in the user's tenant list
	for _, tenantUser := range resp.TenantUsers {
		if tenantUser.Tenant.Id == tenantID {
			// User is a member of this tenant
			return nil
		}
	}

	// User is not a member of this tenant
	return status.Error(codes.PermissionDenied, "access denied: you are not a member of this tenant")
}

// getUserTenantInfo gets user's role and tenant info for a specific tenant
func getUserTenantInfo(ctx context.Context, userServiceClient userProto.UserServiceClient, log *slog.Logger, tenantID, userID string) (string, bool, error) {
	if tenantID == "" {
		return "", false, status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	// Call userservice to get user's tenants
	resp, err := userServiceClient.GetTenants(ctx, &userProto.GetTenantsRequest{})
	if err != nil {
		log.Error("Failed to get user tenants from userservice", "error", err, "userID", userID)
		return "", false, status.Error(codes.Internal, "failed to check tenant access")
	}

	// Find the requested tenant and get user's role and tenant type
	for _, tenantUser := range resp.TenantUsers {
		if tenantUser.Tenant.Id == tenantID {
			userRole := tenantUser.Role.Role
			isPersonal := tenantUser.Tenant.IsPersonal
			return userRole, isPersonal, nil
		}
	}

	// User is not a member of this tenant
	return "", false, status.Error(codes.PermissionDenied, "access denied: you are not a member of this tenant")
}

func (s *VideoAPI) ListVideos(ctx context.Context, req *proto.ListVideosRequest) (*proto.ListVideosResponse, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		s.log.Error("Error getting auth from context", "err", err)
		return nil, status.Error(codes.Unauthenticated, "auth context not found")
	}

	// Get tenant ID from headers/metadata
	tenantID, err := interceptors.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	// Validate user has access to this tenant
	err = isUserInTenant(ctx, s.UserServiceClient, s.log, tenantID, authContext.User.ID)
	if err != nil {
		return nil, err
	}

	var videos []db.VideoserviceVideo

	// Filter videos based on channel_id parameter
	if req.ChannelId != "" {
		// Get videos for specific channel
		videos, err = s.DbQueries.GetVideosByTenantIDAndChannelID(ctx, db.GetVideosByTenantIDAndChannelIDParams{
			TenantID:  sql.NullString{String: tenantID, Valid: true},
			ChannelID: sql.NullString{String: req.ChannelId, Valid: true},
		})
	} else {
		// Get all videos user has access to (their private videos + channel videos they're member of)
		videos, err = s.DbQueries.GetAllAccessibleVideosByTenantID(ctx, db.GetAllAccessibleVideosByTenantIDParams{
			TenantID: sql.NullString{String: tenantID, Valid: true},
			UserID:   authContext.User.ID,
		})
	}

	if err != nil {
		s.log.Error("Error getting videos", "err", err, "tenantID", tenantID, "channelID", req.ChannelId)
		return nil, status.Error(codes.Internal, "failed to get videos")
	}

	protoVideos := make([]*proto.Video, 0, len(videos))

	for _, video := range videos {
		protoVideos = append(protoVideos, &proto.Video{
			Id:          video.ID,
			Title:       video.Title,
			Description: video.Description,
			Url:         video.Url,
			ChannelId:   video.ChannelID.String,              // Include channel_id in response
			Visibility:  proto.Visibility_VISIBILITY_PRIVATE, // All videos are private for now
			CreatedAt:   timestamppb.New(video.CreatedAt),
		})
	}

	return &proto.ListVideosResponse{Videos: protoVideos}, nil
}

func (s *VideoAPI) GetVideo(ctx context.Context, req *proto.GetVideoRequest) (*proto.Video, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		s.log.Error("Error getting auth from context", "err", err)
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	// Get tenant ID from headers/metadata
	tenantID, err := interceptors.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	// Validate user has access to this tenant
	err = isUserInTenant(ctx, s.UserServiceClient, s.log, tenantID, authContext.User.ID)
	if err != nil {
		return nil, err
	}

	// Get video from database with tenant validation
	video, err := s.DbQueries.GetVideoByVideoIDAndTenantID(ctx, db.GetVideoByVideoIDAndTenantIDParams{
		ID: req.VideoId,
		TenantID: sql.NullString{
			String: tenantID,
			Valid:  true,
		},
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "video not found")
		}
		s.log.Error("Error getting video", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	// Convert to proto message
	return &proto.Video{
		Id:          video.ID,
		Title:       video.Title,
		Description: video.Description,
		Url:         video.Url,
		Visibility:  proto.Visibility_VISIBILITY_PRIVATE, // All videos are private for now
		CreatedAt:   timestamppb.New(video.CreatedAt),
	}, nil
}

// ===== VIDEO-CHANNEL MANAGEMENT METHODS =====

func (s *VideoAPI) MoveVideoToChannel(ctx context.Context, req *proto.MoveVideoToChannelRequest) (*proto.MoveVideoToChannelResponse, error) {
	// Common validation
	authContext, tenantID, err := s.PolicyValidator.ValidateBasicRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Get and validate video
	video, err := s.PolicyValidator.GetAndValidateVideo(ctx, req.VideoId, tenantID)
	if err != nil {
		return nil, err
	}

	// Validate permissions for moving the video
	err = s.PolicyValidator.ValidateVideoMovePermissions(ctx, s.ChannelAPI, video, authContext.User.ID, tenantID, req.ChannelId)
	if err != nil {
		return nil, err
	}

	// Move the video to the target channel using single query
	err = s.DbQueries.UpdateVideoChannel(ctx, db.UpdateVideoChannelParams{
		VideoID:          req.VideoId,
		ChannelID:        sql.NullString{String: req.ChannelId, Valid: true},
		TenantID:         sql.NullString{String: tenantID, Valid: true},
		UploadedUserID:   video.UploadedUserID, // For tenant-level video validation
		CurrentChannelID: video.ChannelID,      // For channel video validation
		UpdatedAt:        time.Now(),
	})
	if err != nil {
		s.log.Error("Error moving video to channel", "err", err, "videoID", req.VideoId, "channelID", req.ChannelId)
		return nil, status.Error(codes.Internal, "failed to move video to channel")
	}

	// Verify the video was actually updated (check if WHERE clause matched)
	updatedVideo, err := s.DbQueries.GetVideoByVideoIDAndTenantID(ctx, db.GetVideoByVideoIDAndTenantIDParams{
		ID:       req.VideoId,
		TenantID: sql.NullString{String: tenantID, Valid: true},
	})
	if err != nil {
		s.log.Error("Error verifying video update", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	// Check if the video was actually moved to the target channel
	if !updatedVideo.ChannelID.Valid || updatedVideo.ChannelID.String != req.ChannelId {
		// Database didn't update the row, meaning permission/ownership validation failed
		if !video.ChannelID.Valid || video.ChannelID.String == "" {
			return nil, status.Error(codes.PermissionDenied, "access denied: you can only move your own tenant-level videos")
		} else {
			return nil, status.Error(codes.PermissionDenied, "access denied: video move failed due to permission or state validation")
		}
	}

	return &proto.MoveVideoToChannelResponse{
		Message: "Video moved to channel successfully",
		Video:   s.PolicyValidator.ConvertVideoToProto(&updatedVideo),
	}, nil
}

func (s *VideoAPI) RemoveVideoFromChannel(ctx context.Context, req *proto.RemoveVideoFromChannelRequest) (*proto.RemoveVideoFromChannelResponse, error) {
	// Common validation
	authContext, tenantID, err := s.PolicyValidator.ValidateBasicRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Get and validate video
	video, err := s.PolicyValidator.GetAndValidateVideo(ctx, req.VideoId, tenantID)
	if err != nil {
		return nil, err
	}

	// Validate permissions for removing video from channel
	err = s.PolicyValidator.ValidateVideoRemovalPermissions(ctx, s.ChannelAPI, video, authContext.User.ID, tenantID)
	if err != nil {
		return nil, err
	}

	// Remove the video from the channel
	err = s.DbQueries.RemoveVideoFromChannel(ctx, db.RemoveVideoFromChannelParams{
		VideoID:   req.VideoId,
		TenantID:  sql.NullString{String: tenantID, Valid: true},
		ChannelID: video.ChannelID,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		s.log.Error("Error removing video from channel", "err", err, "videoID", req.VideoId, "channelID", video.ChannelID.String)
		return nil, status.Error(codes.Internal, "failed to remove video from channel")
	}

	// Get updated video
	updatedVideo, err := s.DbQueries.GetVideoByVideoIDAndTenantID(ctx, db.GetVideoByVideoIDAndTenantIDParams{
		ID:       req.VideoId,
		TenantID: sql.NullString{String: tenantID, Valid: true},
	})
	if err != nil {
		s.log.Error("Error getting updated video", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &proto.RemoveVideoFromChannelResponse{
		Message: "Video removed from channel successfully",
		Video:   s.PolicyValidator.ConvertVideoToProto(&updatedVideo),
	}, nil
}

func (s *VideoAPI) DeleteVideo(ctx context.Context, req *proto.DeleteVideoRequest) (*proto.DeleteVideoResponse, error) {
	// Common validation
	authContext, tenantID, err := s.PolicyValidator.ValidateBasicRequest(ctx)
	if err != nil {
		return nil, err
	}

	// Get and validate video
	video, err := s.PolicyValidator.GetAndValidateVideo(ctx, req.VideoId, tenantID)
	if err != nil {
		return nil, err
	}

	// Validate permissions for deleting the video
	err = s.PolicyValidator.ValidateVideoDeletionPermissions(ctx, s.ChannelAPI, video, authContext.User.ID, tenantID)
	if err != nil {
		return nil, err
	}

	// Soft delete the video
	err = s.DbQueries.SoftDeleteVideo(ctx, db.SoftDeleteVideoParams{
		VideoID:   req.VideoId,
		TenantID:  sql.NullString{String: tenantID, Valid: true},
		UpdatedAt: time.Now(),
	})
	if err != nil {
		s.log.Error("Error soft deleting video", "err", err, "videoID", req.VideoId)
		return nil, status.Error(codes.Internal, "failed to delete video")
	}

	return &proto.DeleteVideoResponse{
		Message: "Video deleted successfully",
	}, nil
}

// ===== CHANNEL API METHODS =====

// ValidateChannelRole validates that the role is one of the allowed channel roles
func (s *ChannelAPI) ValidateChannelRole(role string) error {
	if !constants.IsValidChannelRole(role) {
		return status.Error(codes.InvalidArgument, "invalid role. Valid roles are: owner, uploader, viewer")
	}
	return nil
}

// getUserRoleInChannel checks if user has access to channel and returns their role
func (s *ChannelAPI) GetUserRoleInChannel(ctx context.Context, channelID, userID, tenantID string) (string, error) {
	role, err := s.DbQueries.GetUserRoleInChannel(ctx, db.GetUserRoleInChannelParams{
		ChannelID: channelID,
		UserID:    userID,
		TenantID:  tenantID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return "", status.Error(codes.PermissionDenied, "access denied: you are not a member of this channel")
		}
		return "", status.Error(codes.Internal, "failed to check channel access")
	}
	return role, nil
}

// getChannelMemberCount returns the number of members in a channel
func (s *ChannelAPI) GetChannelMemberCount(ctx context.Context, channelID, tenantID string) (int32, error) {
	members, err := s.DbQueries.GetChannelMembersByChannelIDAndTenantID(ctx, db.GetChannelMembersByChannelIDAndTenantIDParams{
		ChannelID: channelID,
		TenantID:  tenantID,
	})
	if err != nil {
		return 0, err
	}
	return int32(len(members)), nil
}

func (s *ChannelAPI) CreateChannel(ctx context.Context, req *proto.CreateChannelRequest) (*proto.CreateChannelResponse, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	// Get tenant ID from headers/metadata
	tenantID, err := interceptors.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	// Get user's role and tenant type
	userRole, isPersonal, err := getUserTenantInfo(ctx, s.UserServiceClient, s.Log, tenantID, authContext.User.ID)
	if err != nil {
		return nil, err
	}

	// Check channel creation permissions:
	// - Personal tenants: Any member can create channels
	// - Organizational tenants: Only super_admin can create channels
	if !isPersonal && userRole != constants.TenantRoleSuperAdmin {
		return nil, status.Error(codes.PermissionDenied, "access denied: only tenant admins can create channels in organizational tenants")
	}

	// Validate input
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "channel name is required")
	}

	if len(req.Name) > 50 {
    	return nil, status.Errorf(codes.InvalidArgument, "channel name cannot exceed 50 characters")
	}

	// Create channel
	channelID := uuid.New().String()
	now := time.Now()

	channel, err := s.DbQueries.CreateChannel(ctx, db.CreateChannelParams{
		ID:          channelID,
		TenantID:    tenantID,
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		CreatedBy:   authContext.User.ID,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		s.Log.Error("Failed to create channel", "error", err)
		return nil, status.Error(codes.Internal, "failed to create channel")
	}

	// Add creator as owner
	_, err = s.DbQueries.CreateChannelMember(ctx, db.CreateChannelMemberParams{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		UserID:    authContext.User.ID,
		Role:      constants.ChannelRoleOwner,
		AddedBy:   authContext.User.ID,
		CreatedAt: now,
	})
	if err != nil {
		s.Log.Error("Failed to add creator as channel owner", "error", err)
		return nil, status.Error(codes.Internal, "failed to create channel")
	}

	// Create response channel proto
	channelProto := &proto.Channel{
		Id:          channel.ID,
		TenantId:    channel.TenantID,
		Name:        channel.Name,
		Description: channel.Description.String,
		CreatedBy:   channel.CreatedBy,
		CreatedAt:   timestamppb.New(channel.CreatedAt),
		UpdatedAt:   timestamppb.New(channel.UpdatedAt),
		UserRole:    constants.ChannelRoleOwner, // Creator is always owner
		MemberCount: 1,                          // Creator is the first member
	}

	return &proto.CreateChannelResponse{
		Message: "Channel created successfully",
		Channel: channelProto,
	}, nil
}

func (s *ChannelAPI) GetChannels(ctx context.Context, req *proto.GetChannelsRequest) (*proto.GetChannelsResponse, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	// Get tenant ID from headers/metadata
	tenantID, err := interceptors.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	// Validate user has access to this tenant
	err = isUserInTenant(ctx, s.UserServiceClient, s.Log, tenantID, authContext.User.ID)
	if err != nil {
		return nil, err
	}

	// Get all channels in tenant
	channels, err := s.DbQueries.GetChannelsByTenantID(ctx, tenantID)
	if err != nil {
		s.Log.Error("Failed to get channels", "error", err)
		return nil, status.Error(codes.Internal, "failed to get channels")
	}

	tenantIDParam := sql.NullString{String: tenantID, Valid: true}
	videoCounts, err := s.dbQueries.GetVideoCountsPerChannelByTenantID(ctx, tenantIDParam)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get video counts: %v", err)
	}

	
	// Build map for lookup
	countMap := make(map[string]int32)
	for _, vc := range videoCounts {
		if vc.ChannelID.Valid {
			countMap[vc.ChannelID.String] = int32(vc.VideoCount)
		}
	}

	// Filter channels where user is a member and include their role
	var userChannels []*proto.Channel
	for _, channel := range channels {
		// Check if user is a member of this channel and get their role
		userRole, err := s.GetUserRoleInChannel(ctx, channel.ID, authContext.User.ID, tenantID)
		if err == nil {
			// Create channel proto
			channelProto := &proto.Channel{
				Id:          channel.ID,
				TenantId:    channel.TenantID,
				Name:        channel.Name,
				Description: channel.Description.String,
				CreatedBy:   channel.CreatedBy,
				CreatedAt:   timestamppb.New(channel.CreatedAt),
				UpdatedAt:   timestamppb.New(channel.UpdatedAt),
				UserRole:    userRole, // Include the user's role in this channel
				VideoCount:  countMap[channel.ID],
			}

			// Only include member count for channel owners
			if userRole == constants.ChannelRoleOwner {
				memberCount, err := s.GetChannelMemberCount(ctx, channel.ID, tenantID)
				if err != nil {
					s.Log.Warn("Failed to get member count for channel", "channel_id", channel.ID, "error", err)
					memberCount = 0 // Default to 0 if we can't get the count
				}
				channelProto.MemberCount = memberCount
			}

			// User is a member, include this channel
			userChannels = append(userChannels, channelProto)
		}
	}

	return &proto.GetChannelsResponse{
		Message:  "Channels retrieved successfully",
		Channels: userChannels,
	}, nil
}

func (s *ChannelAPI) UpdateChannel(ctx context.Context, req *proto.UpdateChannelRequest) (*proto.UpdateChannelResponse, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	// Get tenant ID from headers/metadata
	tenantID, err := interceptors.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	// Validate user has access to this tenant
	err = isUserInTenant(ctx, s.UserServiceClient, s.Log, tenantID, authContext.User.ID)
	if err != nil {
		return nil, err
	}

	// Validate input
	if req.ChannelId == "" {
		return nil, status.Error(codes.InvalidArgument, "channel ID is required")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "channel name is required")
	}

	// Check if user is owner of the channel
	role, err := s.GetUserRoleInChannel(ctx, req.ChannelId, authContext.User.ID, tenantID)
	if err != nil {
		return nil, err
	}
	if role != constants.ChannelRoleOwner {
		return nil, status.Error(codes.PermissionDenied, "access denied: only channel owners can update channels")
	}

	// Update channel
	channel, err := s.DbQueries.UpdateChannel(ctx, db.UpdateChannelParams{
		ID:          req.ChannelId,
		TenantID:    tenantID,
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "channel not found")
		}
		s.Log.Error("Failed to update channel", "error", err)
		return nil, status.Error(codes.Internal, "failed to update channel")
	}

	// Create response channel proto
	channelProto := &proto.Channel{
		Id:          channel.ID,
		TenantId:    channel.TenantID,
		Name:        channel.Name,
		Description: channel.Description.String,
		CreatedBy:   channel.CreatedBy,
		CreatedAt:   timestamppb.New(channel.CreatedAt),
		UpdatedAt:   timestamppb.New(channel.UpdatedAt),
		UserRole:    role, // Include the user's role in the response
	}

	// Include member count since only owners can update channels
	memberCount, err := s.GetChannelMemberCount(ctx, channel.ID, tenantID)
	if err != nil {
		s.Log.Warn("Failed to get member count for updated channel", "channel_id", channel.ID, "error", err)
		memberCount = 0 // Default to 0 if we can't get the count
	}
	channelProto.MemberCount = memberCount

	return &proto.UpdateChannelResponse{
		Message: "Channel updated successfully",
		Channel: channelProto,
	}, nil
}

func (s *ChannelAPI) GetMembers(ctx context.Context, req *proto.GetChannelMembersRequest) (*proto.GetChannelMembersResponse, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	// Get tenant ID from headers/metadata
	tenantID, err := interceptors.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	// Validate user has access to this tenant
	err = isUserInTenant(ctx, s.UserServiceClient, s.Log, tenantID, authContext.User.ID)
	if err != nil {
		return nil, err
	}

	// Get user's role in the channel
	userRole, err := s.GetUserRoleInChannel(ctx, req.ChannelId, authContext.User.ID, tenantID)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	// Only owners can view members
	if userRole != constants.ChannelRoleOwner {
		return nil, status.Error(codes.PermissionDenied, "only channel owners can view members")
	}

	// Get all channel members from database
	members, err := s.DbQueries.GetChannelMembersByChannelIDAndTenantID(ctx, db.GetChannelMembersByChannelIDAndTenantIDParams{
		ChannelID: req.ChannelId,
		TenantID:  tenantID,
	})
	if err != nil {
		s.Log.Error("failed to get channel members", "error", err)
		return nil, status.Error(codes.Internal, "failed to get channel members")
	}

	// Get all tenant users to lookup user details for the channel members
	tenantUsersResp, err := s.tenantServiceClient.GetUsers(ctx, &userProto.GetUsersRequest{
		TenantId: tenantID,
	})
	if err != nil {
		s.Log.Error("failed to get tenant users", "error", err)
		return nil, status.Error(codes.Internal, "failed to get user details")
	}

	// Create a map for quick user lookup
	userMap := make(map[string]*userProto.User)
	for _, tenantUser := range tenantUsersResp.TenantUsers {
		userMap[tenantUser.User.Id] = tenantUser.User
	}

	// Convert to proto format, excluding current user
	var protoMembers []*proto.ChannelMember
	for _, member := range members {
		// Get user details from the map
		user, exists := userMap[member.UserID]
		if !exists {
			s.Log.Warn("user not found in tenant users", "user_id", member.UserID)
			continue // Skip members whose user details we can't find
		}

		protoMembers = append(protoMembers, &proto.ChannelMember{
			User:      user,
			Role:      member.Role,
			AddedBy:   member.AddedBy,
			CreatedAt: timestamppb.New(member.CreatedAt),
		})
	}

	return &proto.GetChannelMembersResponse{
		Message:        "Channel members retrieved successfully",
		ChannelMembers: protoMembers,
	}, nil
}

func (s *ChannelAPI) AddMember(ctx context.Context, req *proto.AddChannelMemberRequest) (*proto.AddChannelMemberResponse, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	// Get tenant ID from headers/metadata
	tenantID, err := interceptors.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	// Validate user has access to this tenant
	err = isUserInTenant(ctx, s.UserServiceClient, s.Log, tenantID, authContext.User.ID)
	if err != nil {
		return nil, err
	}

	// Validate input
	if req.ChannelId == "" {
		return nil, status.Error(codes.InvalidArgument, "channel ID is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	// Validate role
	err = s.ValidateChannelRole(req.Role)
	if err != nil {
		return nil, err
	}

	// Check if user is owner of the channel
	role, err := s.GetUserRoleInChannel(ctx, req.ChannelId, authContext.User.ID, tenantID)
	if err != nil {
		return nil, err
	}
	if role != constants.ChannelRoleOwner {
		return nil, status.Error(codes.PermissionDenied, "access denied: only channel owners can add members")
	}

	// Validate that the user being added is a member of the tenant
	err = isUserInTenant(ctx, s.UserServiceClient, s.Log, tenantID, req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user is not a member of this tenant")
	}

	// Add member to channel
	_, err = s.DbQueries.CreateChannelMember(ctx, db.CreateChannelMemberParams{
		ID:        uuid.New().String(),
		ChannelID: req.ChannelId,
		UserID:    req.UserId,
		Role:      req.Role,
		AddedBy:   authContext.User.ID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		s.Log.Error("Failed to add channel member", "error", err)
		return nil, status.Error(codes.Internal, "failed to add channel member")
	}

	return &proto.AddChannelMemberResponse{
		Message: "Member added successfully",
	}, nil
}

func (s *ChannelAPI) RemoveMember(ctx context.Context, req *proto.RemoveChannelMemberRequest) (*proto.RemoveChannelMemberResponse, error) {
	authContext, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	// Get tenant ID from headers/metadata
	tenantID, err := interceptors.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	// Validate user has access to this tenant
	err = isUserInTenant(ctx, s.UserServiceClient, s.Log, tenantID, authContext.User.ID)
	if err != nil {
		return nil, err
	}

	// Validate input
	if req.ChannelId == "" {
		return nil, status.Error(codes.InvalidArgument, "channel ID is required")
	}
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	// Check if user is owner of the channel
	role, err := s.GetUserRoleInChannel(ctx, req.ChannelId, authContext.User.ID, tenantID)
	if err != nil {
		return nil, err
	}
	if role != constants.ChannelRoleOwner {
		return nil, status.Error(codes.PermissionDenied, "access denied: only channel owners can remove members")
	}

	// Don't allow removing the channel owner
	memberRole, err := s.GetUserRoleInChannel(ctx, req.ChannelId, req.UserId, tenantID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user is not a member of this channel")
	}
	if memberRole == constants.ChannelRoleOwner {
		return nil, status.Error(codes.InvalidArgument, "cannot remove channel owner")
	}

	// Remove member from channel
	err = s.DbQueries.DeleteChannelMember(ctx, db.DeleteChannelMemberParams{
		ChannelID: req.ChannelId,
		UserID:    req.UserId,
	})
	if err != nil {
		s.Log.Error("Failed to remove channel member", "error", err)
		return nil, status.Error(codes.Internal, "failed to remove channel member")
	}

	return &proto.RemoveChannelMemberResponse{
		Message: "Member removed successfully",
	}, nil
}
