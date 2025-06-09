package config

type VideoServiceConfig struct {
	DB           DBConfig `json:"db" mapstructure:"db"`
	FileStoreDir string   `json:"fileStoreDir" mapstructure:"fileStoreDir"`
}

type DBConfig struct {
	Driver string `json:"driver" mapstructure:"driver"`
	Url    string `json:"url" mapstructure:"url"`
}
