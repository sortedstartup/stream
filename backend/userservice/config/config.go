package config

type UserServiceConfig struct {
	DB       DBConfig `json:"db" mapstructure:"db"`
	CacheSize int      `json:"cacheSize" mapstructure:"cacheSize"`
}

type DBConfig struct {
	Driver string `json:"driver" mapstructure:"driver"`
	Url    string `json:"url" mapstructure:"url"`
}
