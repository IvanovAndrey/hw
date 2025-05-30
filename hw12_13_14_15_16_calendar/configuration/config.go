package configuration

import (
	"github.com/caarlos0/env/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Logger LoggerConf `mapstructure:"logger"`
	System struct {
		HTTP struct {
			Address      string `mapstructure:"address" env:"HTTP_ADDRESS"`
			WriteTimeout int    `mapstructure:"write_timeout" env:"HTTP_WRITE_TIMEOUT"`
			ReadTimeout  int    `mapstructure:"read_timeout" env:"HTTP_READ_TIMEOUT"`
		} `mapstructure:"http"`

		Grpc struct {
			Port              uint16 `mapstructure:"port" env:"GRPC_PORT"`
			ConnectionTimeout int    `mapstructure:"connection_timeout" env:"GRPC_CONNECTION_TIMEOUT"`
		} `mapstructure:"grpc"`

		Database struct {
			Enable   bool   `mapstructure:"enable" env:"DB_ENABLE"`
			Host     string `mapstructure:"host" env:"DB_HOST"`
			Port     int    `mapstructure:"port" env:"DB_PORT"`
			DBName   string `mapstructure:"db_name" env:"DB_NAME"`
			Scheme   string `mapstructure:"scheme" env:"DB_SCHEME"`
			User     string `mapstructure:"user" env:"DB_USER"`
			Password string `mapstructure:"password" env:"DB_PASSWORD"`
			Timeout  int    `mapstructure:"timeout" env:"DB_TIMEOUT"`
			SSLMode  string `mapstructure:"ssl_mode" env:"DB_SSL_MODE"`
		} `mapstructure:"database"`
	} `mapstructure:"system"`
}

type LoggerConf struct {
	Level string `mapstructure:"level" env:"LOG_LEVEL" envDefault:"debug"`
}

func LoadConfig(configPath string) (*Config, error) {
	var cfg Config

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.System.HTTP); err != nil {
		return nil, err
	}
	if err := env.Parse(&cfg.System.Grpc); err != nil {
		return nil, err
	}
	if err := env.Parse(&cfg.System.Database); err != nil {
		return nil, err
	}

	return &cfg, nil
}
