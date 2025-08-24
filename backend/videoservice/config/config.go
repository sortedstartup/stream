package config

type VideoServiceConfig struct {
	DB           DBConfig               `json:"db" mapstructure:"db"`
	FileStoreDir string                 `json:"fileStoreDir" mapstructure:"fileStoreDir"`
	PlanLimits   []ApplicationPlanLimit `json:"planLimits" mapstructure:"planLimits"`
}

type DBConfig struct {
	Driver string `json:"driver" mapstructure:"driver"`
	Url    string `json:"url" mapstructure:"url"`
}

// ApplicationPlanLimit defines application-specific limits for subscription plans
type ApplicationPlanLimit struct {
	PlanID         string `json:"planId" mapstructure:"planId"`
	StorageLimitMB int64  `json:"storageLimitMB" mapstructure:"storageLimitMB"` // In MB for easier config
	UsersLimit     int64  `json:"usersLimit" mapstructure:"usersLimit"`
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
			StorageLimitMB: 100, // 100 MB
			UsersLimit:     5,
		},
		{
			PlanID:         "standard",
			StorageLimitMB: 102400, // 100 GB (100 * 1024 MB)
			UsersLimit:     10,
		},
		{
			PlanID:         "premium",
			StorageLimitMB: 204800, // 200 GB (200 * 1024 MB)
			UsersLimit:     25,
		},
	}
}

// GetPlanLimitByID returns plan limits for a specific plan ID
func (c *VideoServiceConfig) GetPlanLimitByID(planID string) *ApplicationPlanLimit {
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
