package config

import (
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/spf13/viper"
	s "sortedstartup.com/stream/videoservice/config"
)

type MonolithConfig struct {
	Server       ServerConfig         `json:"server" mapstructure:"server"`
	LogLevel     string               `json:"logLevel" mapstructure:"logLevel"`
	VideoService s.VideoServiceConfig `json:"videoService" mapstructure:"videoService"`
}

type ServerConfig struct {
	Host        string `json:"host" mapstructure:"host"`
	GRPCPort    int    `json:"grpcPort" mapstructure:"grpcPort"`
	GrpcWebPort int    `json:"grpcWebPort" mapstructure:"grpcWebPort"`
}

func (c *ServerConfig) GRPCAddrPortString() string {
	return fmt.Sprintf("%s:%d", c.Host, c.GRPCPort)
}

func (c *ServerConfig) GrpcWebAddrPortString() string {
	return fmt.Sprintf("%s:%d", c.Host, c.GrpcWebPort)
}

func New() (MonolithConfig, error) {
	// Because of config below config.yaml is read first, then environment variables are read
	// environment variables have precendence over config.yaml
	// if you want to force override a config.yaml value, you can set the environment variable to the desired value

	viper.SetConfigName("config") // Name of the config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // Path to look for the config file in
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // Read configuration from environment variables

	// Setting defaults

	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.grpcPort", 50051)
	viper.SetDefault("server.grpcWebPort", 8080)

	viper.SetDefault("videoService.db.driver", "sqlite")
	viper.SetDefault("videoService.db.url", "db.sqlite")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		slog.Warn("Error while reading config file", "err", err)
		//TODO handle it better !
		//return MonolithConfig{}, fmt.Errorf("Error while reading config file: %w", err)
	}

	var config MonolithConfig
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %s", err)
		return MonolithConfig{}, fmt.Errorf("Unable to decode into struct, %w", err)
	}

	return config, nil
}
