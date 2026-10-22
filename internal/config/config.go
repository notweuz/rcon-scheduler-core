package config

import (
	"fmt"
	"rcon-scheduler-core/internal/server"

	"github.com/spf13/viper"
)

type Config struct {
	Timezone string   `mapstructure:"time_zone" default:"UTC"`
	LogLevel string   `mapstructure:"log_level" default:"info"`
	Servers  []Server `mapstructure:"servers,omitempty"`
}

type Server struct {
	Address  string      `mapstructure:"address"`
	Port     int         `mapstructure:"port" default:"8080"`
	Password string      `mapstructure:"password"`
	Type     server.Type `mapstructure:"type" default:"source"`
	Jobs     []Job       `mapstructure:"job,omitempty"`
}

type Job struct {
	Cron  string `mapstructure:"cron"`
	Steps []Step `mapstructure:"steps"`
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
	err = v.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("err when unmarshaling config: %w", err)
	}

	return cfg, nil
}
