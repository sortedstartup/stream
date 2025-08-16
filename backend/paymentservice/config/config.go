package config

type PaymentServiceConfig struct {
	DB              DBConfig     `json:"db" mapstructure:"db"`
	PaymentProvider string       `json:"paymentProvider" mapstructure:"paymentProvider"`
	StripeConfig    StripeConfig `json:"stripe" mapstructure:"stripe"`
	Plans           []PlanConfig `json:"plans" mapstructure:"plans"`
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

type PlanConfig struct {
	ID             string `json:"id" mapstructure:"id"`
	Name           string `json:"name" mapstructure:"name"`
	StorageLimitMB int64  `json:"storageLimitMB" mapstructure:"storageLimitMB"` // In MB for easier config
	UsersLimit     int64  `json:"usersLimit" mapstructure:"usersLimit"`
	PriceCents     int64  `json:"priceCents" mapstructure:"priceCents"`
}

// StorageLimitBytes converts MB to bytes for database storage
func (p *PlanConfig) StorageLimitBytes() int64 {
	return p.StorageLimitMB * 1024 * 1024
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

// GetDefaultPlans returns default plan configuration if none provided
func GetDefaultPlans() []PlanConfig {
	return []PlanConfig{
		{
			ID:             "free",
			Name:           "Free Plan",
			StorageLimitMB: 100, // 100 MB
			UsersLimit:     5,
			PriceCents:     0,
		},
		{
			ID:             "standard",
			Name:           "Standard Plan",
			StorageLimitMB: 102400, // 100 GB (100 * 1024 MB)
			UsersLimit:     10,     // Reduced from 50 to 10
			PriceCents:     2900,   // $29.00
		},
		{
			ID:             "premium",
			Name:           "Premium Plan",
			StorageLimitMB: 204800, // 200 GB (200 * 1024 MB)
			UsersLimit:     25,
			PriceCents:     9900, // $99.00
		},
	}
}
