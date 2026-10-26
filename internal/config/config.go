package config

import (
	"fmt"

	"github.com/creasty/defaults"
	"github.com/notweuz/rcon-scheduler-core/server"

	"github.com/spf13/viper"
)

type Config struct {
	Timezone string   `mapstructure:"timezone" default:"Europe/Moscow"`
	LogLevel string   `mapstructure:"log_level" default:"info"`
	Servers  []Server `mapstructure:"servers,omitempty"`
}

type Server struct {
	Address  string      `mapstructure:"address"`
	Port     int         `mapstructure:"port" default:"27015"`
	Password string      `mapstructure:"password"`
	Type     server.Type `mapstructure:"type" default:"source"`
	Jobs     []Job       `mapstructure:"jobs,omitempty"`
}

type Job struct {
	Name       string `mapstructure:"name"`
	Cron       string `mapstructure:"cron"`
	LogResults bool   `mapstructure:"log_results"`
	Steps      []Step `mapstructure:"steps"`
}

type Step struct {
	Execute *string `mapstructure:"execute,omitempty"`
	Wait    *string `mapstructure:"wait,omitempty"`
}

func Load(name string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.SetConfigName(name)

	err := v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("err when reading config: %w", err)
	}

	cfg := &Config{}
	err = v.Unmarshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("err when unmarshaling config: %w", err)
	}
	if err := defaults.Set(cfg); err != nil {
		return nil, fmt.Errorf("apply defaults: %w", err)
	}

	return cfg, nil
}
