package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"sortedstartup.com/stream/common/interceptors"
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

// Init initializes the payment service, including setting up plans from configuration
func (s *PaymentServer) Init(ctx context.Context) error {
	return s.initializePlans(ctx)
}

// initializePlans creates or updates plans based on configuration
func (s *PaymentServer) initializePlans(ctx context.Context) error {
	// Use configured plans or fall back to defaults
	plans := s.config.Plans
	if len(plans) == 0 {
		log.Printf("No plans configured, using defaults")
		plans = config.GetDefaultPlans()
	}

	for _, planConfig := range plans {
		// Check if plan already exists
		_, err := s.db.GetPlan(ctx, planConfig.ID)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to check existing plan %s: %w", planConfig.ID, err)
		}

		now := time.Now().Unix()

		if err == sql.ErrNoRows {
			// Plan doesn't exist, create it
			log.Printf("Creating plan: %s (%s)", planConfig.ID, planConfig.Name)
			_, err = s.db.CreatePlan(ctx, db.CreatePlanParams{
				ID:         planConfig.ID,
				Name:       planConfig.Name,
				PriceCents: sql.NullInt64{Int64: planConfig.PriceCents, Valid: true},
				CreatedAt:  now,
				UpdatedAt:  now,
			})
			if err != nil {
				return fmt.Errorf("failed to create plan %s: %w", planConfig.ID, err)
			}
		} else {
			// Plan exists, update it with current config
			log.Printf("Updating plan: %s (%s)", planConfig.ID, planConfig.Name)
			_, err = s.db.UpdatePlan(ctx, db.UpdatePlanParams{
				ID:         planConfig.ID,
				Name:       planConfig.Name,
				PriceCents: sql.NullInt64{Int64: planConfig.PriceCents, Valid: true},
				UpdatedAt:  now,
			})
			if err != nil {
				return fmt.Errorf("failed to update plan %s: %w", planConfig.ID, err)
			}
		}
	}

	log.Printf("Successfully initialized %d plans", len(plans))
	return nil
}

// CreateUserSubscription creates a subscription for a user (called by application services)
func (s *PaymentServer) CreateUserSubscription(ctx context.Context, req *pb.CreateUserSubscriptionRequest) (*pb.CreateUserSubscriptionResponse, error) {
	if req.UserId == "" || req.PlanId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and plan_id are required")
	}

	// Check if user already has subscription
	_, err := s.db.GetUserSubscription(ctx, req.UserId)
	if err == nil {
		// User already has subscription
		return &pb.CreateUserSubscriptionResponse{
			Success:        true,
			SubscriptionId: "", // Existing subscription
		}, nil
	} else if err != sql.ErrNoRows {
		return nil, status.Error(codes.Internal, "failed to check existing subscription")
	}

	// Validate plan exists
	_, err = s.db.GetPlan(ctx, req.PlanId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "plan not found")
	}

	now := time.Now().Unix()

	// Create subscription
	subscription, err := s.db.CreateUserSubscription(ctx, db.CreateUserSubscriptionParams{
		ID:        fmt.Sprintf("sub_%s_%d", req.UserId, now),
		UserID:    req.UserId,
		PlanID:    req.PlanId,
		Provider:  s.config.PaymentProvider,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create subscription")
	}

	return &pb.CreateUserSubscriptionResponse{
		Success:        true,
		SubscriptionId: subscription.ID,
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

	// Get plan details
	plan, err := s.db.GetPlan(ctx, subscription.PlanID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get plan details")
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
			Plan: &pb.Plan{
				Id:         plan.ID,
				Name:       plan.Name,
				PriceCents: plan.PriceCents.Int64,
				IsActive:   plan.IsActive.Bool,
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
			req.PlanId,
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

// GetPlans returns all available subscription plans (authentication required)
func (s *PaymentServer) GetPlans(ctx context.Context, req *pb.GetPlansRequest) (*pb.GetPlansResponse, error) {
	// Verify user is authenticated
	_, err := interceptors.AuthFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	// Get all active plans from database
	dbPlans, err := s.db.GetActivePlans(ctx)
	if err != nil {
		log.Printf("GetPlans: Database error: %v", err)
		return &pb.GetPlansResponse{
			Success:      false,
			ErrorMessage: "Failed to retrieve plans",
		}, nil
	}

	// Convert database plans to proto plans
	var plans []*pb.Plan
	for _, dbPlan := range dbPlans {
		plans = append(plans, &pb.Plan{
			Id:         dbPlan.ID,
			Name:       dbPlan.Name,
			PriceCents: dbPlan.PriceCents.Int64,
			IsActive:   dbPlan.IsActive.Bool,
		})
	}

	return &pb.GetPlansResponse{
		Plans:   plans,
		Success: true,
	}, nil
}
