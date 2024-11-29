package config

type VideoServiceConfig struct {
	DB DBConfig `json:"db" mapstructure:"db"`
}

type DBConfig struct {
	Driver string `json:"driver" mapstructure:"driver"`
	Url    string `json:"url" mapstructure:"url"`
}