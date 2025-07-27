package providers

import (
	"fmt"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"
	"sortedstartup.com/stream/paymentservice/config"
)

type StripeProvider struct {
	config config.StripeConfig
}

func NewStripeProvider(cfg config.StripeConfig) *StripeProvider {
	stripe.Key = cfg.SecretKey
	return &StripeProvider{
		config: cfg,
	}
}

// CreateCheckoutSession creates a Stripe checkout session
func (s *StripeProvider) CreateCheckoutSession(userID, priceID, successURL, cancelURL string) (string, string, error) {
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL:        stripe.String(successURL),
		CancelURL:         stripe.String(cancelURL),
		ClientReferenceID: stripe.String(userID), // Store user ID for webhook processing
		Metadata: map[string]string{
			"user_id": userID,
		},
	}

	session, err := session.New(params)
	if err != nil {
		return "", "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	return session.URL, session.ID, nil
}

// VerifyWebhook verifies Stripe webhook signature
func (s *StripeProvider) VerifyWebhook(payload []byte, signature string) (stripe.Event, error) {
	event, err := webhook.ConstructEvent(payload, signature, s.config.WebhookSecret)
	if err != nil {
		return stripe.Event{}, fmt.Errorf("webhook signature verification failed: %w", err)
	}
	return event, nil
}
