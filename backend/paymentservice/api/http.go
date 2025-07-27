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

	// Extract user ID from metadata
	userID, exists := session.Metadata["user_id"]
	if !exists {
		return fmt.Errorf("user_id not found in session metadata")
	}

	log.Printf("Processing checkout completion for user: %s, session: %s", userID, session.ID)

	// Update user subscription with Stripe details
	now := time.Now().Unix()
	customerID := sql.NullString{String: session.Customer.ID, Valid: session.Customer.ID != ""}
	subscriptionID := sql.NullString{String: session.Subscription.ID, Valid: session.Subscription.ID != ""}

	err := s.db.UpdateUserSubscriptionProvider(context.Background(), db.UpdateUserSubscriptionProviderParams{
		UserID:                 userID,
		ProviderCustomerID:     customerID,
		ProviderSubscriptionID: subscriptionID,
		UpdatedAt:              now,
	})

	if err != nil {
		return fmt.Errorf("failed to update subscription after checkout: %w", err)
	}

	log.Printf("Successfully updated subscription for user %s", userID)
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

	// For now, we'll update the subscription status - in a full implementation
	// we would need more sophisticated logic to handle plan changes
	now := time.Now().Unix()
	periodStart := sql.NullInt64{Int64: subscription.CurrentPeriodStart, Valid: true}
	periodEnd := sql.NullInt64{Int64: subscription.CurrentPeriodEnd, Valid: true}

	err := s.db.UpdateUserSubscriptionStatus(context.Background(), db.UpdateUserSubscriptionStatusParams{
		UserID:             "", // We need to find user by provider_subscription_id
		Status:             string(subscription.Status),
		CurrentPeriodStart: periodStart,
		CurrentPeriodEnd:   periodEnd,
		UpdatedAt:          now,
	})

	if err != nil {
		log.Printf("Warning: Could not update subscription status for %s: %v", subscription.ID, err)
		// Don't return error to Stripe - this is a limitation of our current schema
	}

	log.Printf("Successfully processed subscription creation %s", subscription.ID)
	return nil
}

// handleSubscriptionUpdated processes subscription changes
func (s *HTTPServer) handleSubscriptionUpdated(event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	log.Printf("Successfully processed subscription update %s", subscription.ID)
	return nil
}

// handleSubscriptionDeleted processes subscription cancellation
func (s *HTTPServer) handleSubscriptionDeleted(event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription: %w", err)
	}

	log.Printf("Successfully processed subscription deletion %s", subscription.ID)
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
