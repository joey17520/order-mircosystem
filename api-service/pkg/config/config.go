package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type ConsulConfig struct {
	Address string `yaml:"address"`
}

type ServiceConfig struct {
	Order struct {
		Name string `yaml:"name"`
		Port int    `yaml:"port"`
	} `yaml:"order"`
	Inventory struct {
		Name string `yaml:"name"`
		Port int    `yaml:"port"`
	} `yaml:"inventory"`
	Payment struct {
		Name string `yaml:"name"`
		Port int    `yaml:"port"`
	} `yaml:"payment"`
}

type CorsConfig struct {
	AllowOrigins     []string      `mapstructure:"allow_origins" yaml:"allow_origins"`
	AllowMethods     []string      `mapstructure:"allow_methods" yaml:"allow_methods"`
	AllowHeaders     []string      `mapstructure:"allow_headers" yaml:"allow_headers"`
	ExposeHeaders    []string      `mapstructure:"expose_headers" yaml:"expose_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials" yaml:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age" yaml:"max_age"`
}

type JaegerConfig struct {
	AgentHost   string `mapstructure:"agent_host" yaml:"agent_host"`
	AgentPort   int    `mapstructure:"agent_port" yaml:"agent_port"`
	ServiceName string `mapstructure:"service_name" yaml:"service_name"`
}

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Consul  ConsulConfig  `yaml:"consul"`
	Service ServiceConfig `yaml:"proxy"`
	CORS    CorsConfig    `yaml:"cors"`
	Jaeger  JaegerConfig  `yaml:"jaeger"`
}

func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %v", err)
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Printf("Unable to decode into struct: %v", err)
		return nil, err
	}
	return &config, nil
}
