package config

type PaymentServiceConfig struct {
	DB              DBConfig     `json:"db" mapstructure:"db"`
	PaymentProvider string       `json:"paymentProvider" mapstructure:"paymentProvider"`
	StripeConfig    StripeConfig `json:"stripe" mapstructure:"stripe"`
}

type DBConfig struct {
	Driver string `json:"driver" mapstructure:"driver"`
	Url    string `json:"url" mapstructure:"url"`
}

type StripeConfig struct {
	SecretKey       string `json:"secretKey" mapstructure:"secretKey"`
	PublishableKey  string `json:"publishableKey" mapstructure:"publishableKey"`
	WebhookSecret   string `json:"webhookSecret" mapstructure:"webhookSecret"`
	StandardPriceID string `json:"standardPriceId" mapstructure:"standardPriceId"`
	PremiumPriceID  string `json:"premiumPriceId" mapstructure:"premiumPriceId"`
}

// GetStripePriceID maps plan ID to Stripe price ID
func (c *PaymentServiceConfig) GetStripePriceID(planID string) string {
	switch planID {
	case "standard":
		return c.StripeConfig.StandardPriceID
	case "premium":
		return c.StripeConfig.PremiumPriceID
	case "free":
		return "" // Free plan has no Stripe price
	default:
		return ""
	}
}
