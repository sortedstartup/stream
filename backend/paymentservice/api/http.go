package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v76"
	"sortedstartup.com/stream/paymentservice/config"
	"sortedstartup.com/stream/paymentservice/db"
	"sortedstartup.com/stream/paymentservice/providers"
)

type HTTPServer struct {
	db     *db.Queries
	config *config.PaymentServiceConfig
	stripe *providers.StripeProvider
}

func NewHTTPServer(database *db.Queries, cfg *config.PaymentServiceConfig) *HTTPServer {
	var stripeProvider *providers.StripeProvider
	if cfg.PaymentProvider == "stripe" {
		stripeProvider = providers.NewStripeProvider(cfg.StripeConfig)
	}

	return &HTTPServer{
		db:     database,
		config: cfg,
		stripe: stripeProvider,
	}
}

// StripeWebhook handles Stripe webhook events
func (s *HTTPServer) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Verify webhook signature
	signature := r.Header.Get("Stripe-Signature")
	if signature == "" {
		log.Printf("Missing Stripe-Signature header")
		http.Error(w, "Missing signature", http.StatusBadRequest)
		return
	}

	event, err := s.stripe.VerifyWebhook(payload, signature)
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		http.Error(w, "Invalid signature", http.StatusBadRequest)
		return
	}

	// Handle the event
	switch event.Type {
	case "checkout.session.completed":
		err = s.handleCheckoutSessionCompleted(event)
	case "customer.subscription.created":
		err = s.handleSubscriptionCreated(event)
	case "customer.subscription.updated":
		err = s.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		err = s.handleSubscriptionDeleted(event)
	case "invoice.payment_succeeded":
		err = s.handleInvoicePaymentSucceeded(event)
	case "invoice.payment_failed":
		err = s.handleInvoicePaymentFailed(event)
	default:
		log.Printf("Unhandled event type: %s", event.Type)
	}

	if err != nil {
		log.Printf("Error handling webhook event %s: %v", event.Type, err)
		http.Error(w, "Webhook handler failed", http.StatusInternalServerError)
		return
	}

	// Respond to Stripe
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleCheckoutSessionCompleted processes successful checkout sessions
func (s *HTTPServer) handleCheckoutSessionCompleted(event stripe.Event) error {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return fmt.Errorf("failed to parse checkout session: %w", err)
	}

	// Extract user ID and plan ID from metadata
	userID, exists := session.Metadata["user_id"]
	if !exists {
		return fmt.Errorf("user_id not found in session metadata")
	}

	planID, exists := session.Metadata["plan_id"]
	if !exists {
		return fmt.Errorf("plan_id not found in session metadata")
	}

	log.Printf("Processing checkout completion for user: %s, plan: %s, session: %s", userID, planID, session.ID)

	// Get the subscription from Stripe
	if session.Subscription == nil {
		return fmt.Errorf("no subscription found in checkout session")
	}

	now := time.Now().Unix()
	customerID := sql.NullString{String: session.Customer.ID, Valid: session.Customer.ID != ""}
	subscriptionID := sql.NullString{String: session.Subscription.ID, Valid: session.Subscription.ID != ""}

	// Update provider details first
	err := s.db.UpdateUserSubscriptionProvider(context.Background(), db.UpdateUserSubscriptionProviderParams{
		UserID:                 userID,
		ProviderCustomerID:     customerID,
		ProviderSubscriptionID: subscriptionID,
		UpdatedAt:              now,
	})
	if err != nil {
		return fmt.Errorf("failed to update subscription provider details: %w", err)
	}

	// Update the user's plan
	err = s.db.UpdateUserSubscriptionPlan(context.Background(), db.UpdateUserSubscriptionPlanParams{
		UserID:    userID,
		PlanID:    planID,
		UpdatedAt: now,
	})
	if err != nil {
		return fmt.Errorf("failed to update plan to %s for user %s: %w", planID, userID, err)
	}

	log.Printf("Successfully updated user %s to plan %s via checkout completion", userID, planID)

	log.Printf("Successfully updated subscription provider details for user %s", userID)
	return nil
}

// handleSubscriptionCreated processes new subscription creation
func (s *HTTPServer) handleSubscriptionCreated(event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	// Determine plan ID from price ID
	planID := s.getPlanIDFromPriceID(subscription.Items.Data[0].Price.ID)
	if planID == "" {
		return fmt.Errorf("unknown price ID: %s", subscription.Items.Data[0].Price.ID)
	}

	log.Printf("Processing subscription creation: %s with plan %s", subscription.ID, planID)

	// Try to find user by provider_subscription_id
	userID, err := s.findUserBySubscriptionID(subscription.ID)
	if err != nil {
		// This is expected if checkout.session.completed hasn't run yet
		// We'll just log and continue - the plan will be set by checkout.session.completed
		log.Printf("Could not find user for subscription %s (checkout may not be complete yet): %v", subscription.ID, err)
		return nil
	}

	now := time.Now().Unix()

	// Update the user's plan (backup in case checkout.session.completed missed it)
	err = s.db.UpdateUserSubscriptionPlan(context.Background(), db.UpdateUserSubscriptionPlanParams{
		UserID:    userID,
		PlanID:    planID,
		UpdatedAt: now,
	})
	if err != nil {
		log.Printf("Warning: failed to update user plan to %s: %v", planID, err)
	} else {
		log.Printf("Successfully updated user %s to plan %s via subscription.created", userID, planID)
	}

	// Update subscription status and billing period
	periodStart := sql.NullInt64{Int64: subscription.CurrentPeriodStart, Valid: true}
	periodEnd := sql.NullInt64{Int64: subscription.CurrentPeriodEnd, Valid: true}

	err = s.db.UpdateUserSubscriptionStatus(context.Background(), db.UpdateUserSubscriptionStatusParams{
		UserID:             userID,
		Status:             string(subscription.Status),
		CurrentPeriodStart: periodStart,
		CurrentPeriodEnd:   periodEnd,
		UpdatedAt:          now,
	})
	if err != nil {
		log.Printf("Warning: Could not update subscription status for %s: %v", subscription.ID, err)
	}

	log.Printf("Successfully updated user %s to plan %s via subscription %s", userID, planID, subscription.ID)
	return nil
}

// handleSubscriptionUpdated processes subscription changes
func (s *HTTPServer) handleSubscriptionUpdated(event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	log.Printf("Processing subscription update: %s, status: %s", subscription.ID, subscription.Status)

	// Try to find user by provider_subscription_id
	userID, err := s.findUserBySubscriptionID(subscription.ID)
	if err != nil {
		log.Printf("Could not find user for subscription %s: %v", subscription.ID, err)
		return nil // Don't fail webhook for this
	}

	now := time.Now().Unix()

	// Update subscription status and billing period based on Stripe data
	periodStart := sql.NullInt64{Int64: subscription.CurrentPeriodStart, Valid: true}
	periodEnd := sql.NullInt64{Int64: subscription.CurrentPeriodEnd, Valid: true}

	err = s.db.UpdateUserSubscriptionStatus(context.Background(), db.UpdateUserSubscriptionStatusParams{
		UserID:             userID,
		Status:             string(subscription.Status),
		CurrentPeriodStart: periodStart,
		CurrentPeriodEnd:   periodEnd,
		UpdatedAt:          now,
	})
	if err != nil {
		return fmt.Errorf("failed to update subscription status for user %s: %w", userID, err)
	}

	log.Printf("Successfully updated subscription %s for user %s to status: %s", subscription.ID, userID, subscription.Status)
	return nil
}

// handleSubscriptionDeleted processes subscription cancellation
func (s *HTTPServer) handleSubscriptionDeleted(event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	log.Printf("Processing subscription deletion: %s", subscription.ID)

	// Try to find user by provider_subscription_id
	userID, err := s.findUserBySubscriptionID(subscription.ID)
	if err != nil {
		log.Printf("Could not find user for subscription %s: %v", subscription.ID, err)
		return nil // Don't fail webhook for this
	}

	now := time.Now().Unix()

	// Update subscription status to canceled
	err = s.db.UpdateUserSubscriptionStatus(context.Background(), db.UpdateUserSubscriptionStatusParams{
		UserID:             userID,
		Status:             "canceled",
		CurrentPeriodStart: sql.NullInt64{},
		CurrentPeriodEnd:   sql.NullInt64{},
		UpdatedAt:          now,
	})
	if err != nil {
		return fmt.Errorf("failed to update subscription status to canceled for user %s: %w", userID, err)
	}

	log.Printf("Successfully canceled subscription %s for user %s", subscription.ID, userID)
	return nil
}

// handleInvoicePaymentSucceeded processes successful payments
func (s *HTTPServer) handleInvoicePaymentSucceeded(event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to parse invoice: %w", err)
	}

	log.Printf("Successfully processed payment success for invoice %s", invoice.ID)
	return nil
}

// handleInvoicePaymentFailed processes failed payments
func (s *HTTPServer) handleInvoicePaymentFailed(event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to parse invoice: %w", err)
	}

	log.Printf("Successfully processed payment failure for invoice %s", invoice.ID)
	return nil
}

// getPlanIDFromPriceID maps Stripe price IDs to our plan IDs
func (s *HTTPServer) getPlanIDFromPriceID(priceID string) string {
	if priceID == s.config.StripeConfig.StandardPriceID {
		return "standard"
	}
	if priceID == s.config.StripeConfig.PremiumPriceID {
		return "premium"
	}
	return ""
}

// findUserBySubscriptionID finds a user by their Stripe subscription ID
func (s *HTTPServer) findUserBySubscriptionID(subscriptionID string) (string, error) {
	userID, err := s.db.GetUserByProviderSubscriptionID(context.Background(), sql.NullString{
		String: subscriptionID,
		Valid:  subscriptionID != "",
	})
	if err != nil {
		return "", fmt.Errorf("user not found for subscription ID %s: %w", subscriptionID, err)
	}
	return userID, nil
}
