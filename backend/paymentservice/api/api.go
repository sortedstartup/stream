package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"sortedstartup.com/stream/paymentservice/config"
	"sortedstartup.com/stream/paymentservice/db"
	pb "sortedstartup.com/stream/paymentservice/proto"
	"sortedstartup.com/stream/paymentservice/providers"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentServer struct {
	pb.UnimplementedPaymentServiceServer
	db     *db.Queries
	config *config.PaymentServiceConfig
	stripe *providers.StripeProvider
}

func NewPaymentServer(database *db.Queries, cfg *config.PaymentServiceConfig) *PaymentServer {
	var stripeProvider *providers.StripeProvider
	if cfg.PaymentProvider == "stripe" {
		stripeProvider = providers.NewStripeProvider(cfg.StripeConfig)
	}

	return &PaymentServer{
		db:     database,
		config: cfg,
		stripe: stripeProvider,
	}
}

// CheckUserAccess checks if user can perform specific action
func (s *PaymentServer) CheckUserAccess(ctx context.Context, req *pb.CheckUserAccessRequest) (*pb.CheckUserAccessResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Get user access info from database
	accessInfo, err := s.db.CheckUserAccess(ctx, req.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			// User doesn't have subscription - need to initialize with free plan
			return &pb.CheckUserAccessResponse{
				HasAccess:      false,
				Reason:         "no_subscription",
				IsNearLimit:    false,
				WarningMessage: "No subscription found. Please upgrade to continue.",
			}, nil
		}
		return nil, status.Error(codes.Internal, "failed to check user access")
	}

	// Determine access based on usage type and requested amount
	hasAccess := false
	reason := ""
	isNearLimit := false
	warningMessage := ""

	switch req.UsageType {
	case "storage":
		wouldExceed := accessInfo.CurrentStorageBytes+req.RequestedUsage > accessInfo.StorageLimitBytes.Int64
		hasAccess = accessInfo.HasStorageAccess == 1 && !wouldExceed
		if !hasAccess {
			if accessInfo.SubscriptionStatus != "active" {
				reason = "subscription_inactive"
			} else {
				reason = "storage_limit_exceeded"
			}
		}
		isNearLimit = accessInfo.StorageUsagePercent > 75

	case "users":
		wouldExceed := accessInfo.CurrentUsersCount+req.RequestedUsage > accessInfo.UsersLimit.Int64
		hasAccess = accessInfo.HasUsersAccess == 1 && !wouldExceed
		if !hasAccess {
			if accessInfo.SubscriptionStatus != "active" {
				reason = "subscription_inactive"
			} else {
				reason = "users_limit_exceeded"
			}
		}
		isNearLimit = accessInfo.UsersUsagePercent > 75
	}

	// Generate warning message
	if isNearLimit && hasAccess {
		if req.UsageType == "storage" {
			warningMessage = fmt.Sprintf("Storage %.1f%% full. Consider upgrading your plan.", float64(accessInfo.StorageUsagePercent))
		} else {
			warningMessage = fmt.Sprintf("Users %.1f%% of limit. Consider upgrading your plan.", float64(accessInfo.UsersUsagePercent))
		}
	}

	return &pb.CheckUserAccessResponse{
		HasAccess: hasAccess,
		Reason:    reason,
		SubscriptionInfo: &pb.UserSubscriptionInfo{
			UserId: req.UserId,
			Usage: &pb.UserUsage{
				UserId:              req.UserId,
				StorageUsedBytes:    accessInfo.CurrentStorageBytes,
				UsersCount:          int32(accessInfo.CurrentUsersCount),
				StorageUsagePercent: float64(accessInfo.StorageUsagePercent),
				UsersUsagePercent:   float64(accessInfo.UsersUsagePercent),
			},
			Plan: &pb.Plan{
				Id:                accessInfo.PlanID,
				Name:              accessInfo.PlanName.String,
				StorageLimitBytes: accessInfo.StorageLimitBytes.Int64,
				UsersLimit:        int32(accessInfo.UsersLimit.Int64),
				PriceCents:        accessInfo.PriceCents.Int64,
				IsActive:          true,
			},
		},
		IsNearLimit:    isNearLimit,
		WarningMessage: warningMessage,
	}, nil
}

// GetUserSubscription gets user subscription and usage details
func (s *PaymentServer) GetUserSubscription(ctx context.Context, req *pb.GetUserSubscriptionRequest) (*pb.GetUserSubscriptionResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Get user subscription
	subscription, err := s.db.GetUserSubscription(ctx, req.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.GetUserSubscriptionResponse{
				Success:      false,
				ErrorMessage: "No subscription found",
			}, nil
		}
		return nil, status.Error(codes.Internal, "failed to get user subscription")
	}

	// Get user usage
	usage, err := s.db.GetUserUsage(ctx, req.UserId)
	if err != nil && err != sql.ErrNoRows {
		return nil, status.Error(codes.Internal, "failed to get user usage")
	}

	// Get plan details
	plan, err := s.db.GetPlan(ctx, subscription.PlanID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get plan details")
	}

	// Calculate usage percentages
	storagePercent := 0.0
	usersPercent := 0.0
	if plan.StorageLimitBytes > 0 {
		storagePercent = float64(usage.StorageUsedBytes.Int64) / float64(plan.StorageLimitBytes) * 100
	}
	if plan.UsersLimit > 0 {
		usersPercent = float64(usage.UsersCount.Int64) / float64(plan.UsersLimit) * 100
	}

	return &pb.GetUserSubscriptionResponse{
		Success: true,
		SubscriptionInfo: &pb.UserSubscriptionInfo{
			UserId: req.UserId,
			Subscription: &pb.Subscription{
				Id:        subscription.ID,
				UserId:    subscription.UserID,
				PlanId:    subscription.PlanID,
				Provider:  subscription.Provider,
				Status:    subscription.Status,
				CreatedAt: timestamppb.New(time.Unix(subscription.CreatedAt, 0)),
				UpdatedAt: timestamppb.New(time.Unix(subscription.UpdatedAt, 0)),
			},
			Usage: &pb.UserUsage{
				UserId:              req.UserId,
				StorageUsedBytes:    usage.StorageUsedBytes.Int64,
				UsersCount:          int32(usage.UsersCount.Int64),
				StorageUsagePercent: storagePercent,
				UsersUsagePercent:   usersPercent,
				LastCalculatedAt:    timestamppb.New(time.Unix(usage.LastCalculatedAt.Int64, 0)),
			},
			Plan: &pb.Plan{
				Id:                plan.ID,
				Name:              plan.Name,
				StorageLimitBytes: plan.StorageLimitBytes,
				UsersLimit:        int32(plan.UsersLimit),
				PriceCents:        plan.PriceCents.Int64,
				IsActive:          plan.IsActive.Bool,
			},
		},
	}, nil
}

// CreateCheckoutSession creates a payment checkout session
func (s *PaymentServer) CreateCheckoutSession(ctx context.Context, req *pb.CreateCheckoutSessionRequest) (*pb.CreateCheckoutSessionResponse, error) {
	if req.UserId == "" || req.PlanId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and plan_id are required")
	}

	// Validate plan exists and is not free
	_, err := s.db.GetPlan(ctx, req.PlanId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "plan not found")
	}

	if req.PlanId == "free" {
		return nil, status.Error(codes.InvalidArgument, "cannot create checkout for free plan")
	}

	// Create checkout session with payment provider
	if s.config.PaymentProvider == "stripe" && s.stripe != nil {
		priceID := s.config.GetStripePriceID(req.PlanId)
		if priceID == "" {
			return nil, status.Error(codes.Internal, "no price ID configured for plan")
		}

		sessionURL, sessionID, err := s.stripe.CreateCheckoutSession(
			req.UserId,
			priceID,
			req.SuccessUrl,
			req.CancelUrl,
		)
		if err != nil {
			log.Printf("Stripe checkout session creation failed: %v", err)
			return nil, status.Error(codes.Internal, "failed to create checkout session")
		}

		return &pb.CreateCheckoutSessionResponse{
			CheckoutUrl: sessionURL,
			SessionId:   sessionID,
			Success:     true,
		}, nil
	}

	return nil, status.Error(codes.Internal, "payment provider not configured")
}

// UpdateUserUsage updates user usage metrics
func (s *PaymentServer) UpdateUserUsage(ctx context.Context, req *pb.UpdateUserUsageRequest) (*pb.UpdateUserUsageResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	now := time.Now().Unix()

	switch req.UsageType {
	case "storage":
		err := s.db.UpdateUserStorageUsage(ctx, db.UpdateUserStorageUsageParams{
			UserID:           req.UserId,
			StorageUsedBytes: sql.NullInt64{Int64: req.UsageChange, Valid: true},
			LastCalculatedAt: sql.NullInt64{Int64: now, Valid: true},
			UpdatedAt:        now,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to update storage usage")
		}

	case "users":
		err := s.db.UpdateUserUsersCount(ctx, db.UpdateUserUsersCountParams{
			UserID:           req.UserId,
			UsersCount:       sql.NullInt64{Int64: req.UsageChange, Valid: true},
			LastCalculatedAt: sql.NullInt64{Int64: now, Valid: true},
			UpdatedAt:        now,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to update users count")
		}

	default:
		return nil, status.Error(codes.InvalidArgument, "invalid usage_type")
	}

	return &pb.UpdateUserUsageResponse{
		Success: true,
	}, nil
}

// InitializeUser creates a free subscription for a new user
func (s *PaymentServer) InitializeUser(ctx context.Context, req *pb.InitializeUserRequest) (*pb.InitializeUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	now := time.Now().Unix()

	// Create free subscription
	subscription, err := s.db.CreateUserSubscription(ctx, db.CreateUserSubscriptionParams{
		ID:        fmt.Sprintf("sub_%s_%d", req.UserId, now),
		UserID:    req.UserId,
		PlanID:    "free",
		Provider:  s.config.PaymentProvider,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create subscription")
	}

	// Create usage tracking
	usage, err := s.db.CreateUserUsage(ctx, db.CreateUserUsageParams{
		UserID:           req.UserId,
		StorageUsedBytes: sql.NullInt64{Int64: 0, Valid: true},
		UsersCount:       sql.NullInt64{Int64: 0, Valid: true},
		LastCalculatedAt: sql.NullInt64{Int64: now, Valid: true},
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create usage tracking")
	}

	// Get plan details
	plan, err := s.db.GetPlan(ctx, "free")
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get free plan")
	}

	return &pb.InitializeUserResponse{
		Success: true,
		SubscriptionInfo: &pb.UserSubscriptionInfo{
			UserId: req.UserId,
			Subscription: &pb.Subscription{
				Id:       subscription.ID,
				UserId:   subscription.UserID,
				PlanId:   subscription.PlanID,
				Provider: subscription.Provider,
				Status:   subscription.Status,
			},
			Usage: &pb.UserUsage{
				UserId:           req.UserId,
				StorageUsedBytes: usage.StorageUsedBytes.Int64,
				UsersCount:       int32(usage.UsersCount.Int64),
			},
			Plan: &pb.Plan{
				Id:                plan.ID,
				Name:              plan.Name,
				StorageLimitBytes: plan.StorageLimitBytes,
				UsersLimit:        int32(plan.UsersLimit),
				PriceCents:        plan.PriceCents.Int64,
				IsActive:          plan.IsActive.Bool,
			},
		},
	}, nil
}
