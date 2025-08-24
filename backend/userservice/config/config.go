package config

type UserServiceConfig struct {
	DB         DBConfig               `json:"db" mapstructure:"db"`
	CacheSize  int                    `json:"cacheSize" mapstructure:"cacheSize"`
	PlanLimits []ApplicationPlanLimit `json:"planLimits" mapstructure:"planLimits"`
}

type DBConfig struct {
	Driver string `json:"driver" mapstructure:"driver"`
	Url    string `json:"url" mapstructure:"url"`
}

// ApplicationPlanLimit defines application-specific limits for subscription plans
// Now includes both user limits and storage limits (moved from videoservice)
type ApplicationPlanLimit struct {
	PlanID         string `json:"planId" mapstructure:"planId"`
	UsersLimit     int64  `json:"usersLimit" mapstructure:"usersLimit"`
	StorageLimitMB int64  `json:"storageLimitMB" mapstructure:"storageLimitMB"` // In MB for easier config
}

// StorageLimitBytes converts MB to bytes for calculations
func (p *ApplicationPlanLimit) StorageLimitBytes() int64 {
	return p.StorageLimitMB * 1024 * 1024
}

// GetDefaultPlanLimits returns default application-specific plan limits
func GetDefaultPlanLimits() []ApplicationPlanLimit {
	return []ApplicationPlanLimit{
		{
			PlanID:         "free",
			UsersLimit:     5,
			StorageLimitMB: 100, // 100 MB
		},
		{
			PlanID:         "standard",
			UsersLimit:     10,
			StorageLimitMB: 102400, // 100 GB (100 * 1024 MB)
		},
		{
			PlanID:         "premium",
			UsersLimit:     25,
			StorageLimitMB: 204800, // 200 GB (200 * 1024 MB)
		},
	}
}

// GetPlanLimitByID returns plan limits for a specific plan ID
func (c *UserServiceConfig) GetPlanLimitByID(planID string) *ApplicationPlanLimit {
	// Use configured limits if available
	for _, limit := range c.PlanLimits {
		if limit.PlanID == planID {
			return &limit
		}
	}

	// Fall back to defaults
	defaultLimits := GetDefaultPlanLimits()
	for _, limit := range defaultLimits {
		if limit.PlanID == planID {
			return &limit
		}
	}

	return nil
}
