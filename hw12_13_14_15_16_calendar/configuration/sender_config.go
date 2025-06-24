package configuration

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type SenderConfig struct {
	Logger struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"logger"`

	RabbitMQ struct {
		URI   string `mapstructure:"uri"`
		Queue string `mapstructure:"queue"`
	} `mapstructure:"rabbitmq"`
}

func LoadSenderConfig(configPath string) (*SenderConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg SenderConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
