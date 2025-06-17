package configuration

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type RabbitMQConf struct {
	URI   string `mapstructure:"uri"`
	Queue string `mapstructure:"queue"`
}

type SchedulerConf struct {
	Interval time.Duration `mapstructure:"interval"`
}

type SystemConf struct {
	Database DatabaseConf `mapstructure:"database"`
}

type SchedulerConfig struct {
	Logger    LoggerConf    `mapstructure:"logger"`
	RabbitMQ  RabbitMQConf  `mapstructure:"rabbitmq"`
	System    SystemConf    `mapstructure:"system"`
	Scheduler SchedulerConf `mapstructure:"scheduler"`
}

func LoadSchedulerConfig(configPath string) (*SchedulerConfig, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	v.RegisterAlias("scheduler.interval", "scheduler.interval")
	v.SetDefault("scheduler.interval", "10s")

	var cfg SchedulerConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	if v.IsSet("scheduler.interval") {
		str := v.GetString("scheduler.interval")
		dur, err := time.ParseDuration(str)
		if err != nil {
			return nil, fmt.Errorf("invalid scheduler.interval: %w", err)
		}
		cfg.Scheduler.Interval = dur
	}

	return &cfg, nil
}
