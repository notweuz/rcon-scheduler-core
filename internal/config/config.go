package config

type Config struct {
	Timezone string   `yaml:"time_zone"`
	LogLevel string   `yaml:"log_level"`
	Servers  []Server `yaml:"servers"`
}

type Server struct {
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Type     string `yaml:"type"`
	Jobs     []Job  `yaml:"job"`
}

type Job struct {
	Cron string `mapstructure:"cron"`

}
