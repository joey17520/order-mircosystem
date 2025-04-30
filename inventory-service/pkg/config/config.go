package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	Host        string `mapstructure:"host"`
	MetricsPort int    `mapstructure:"metrics_port"`
}

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type RabbitMQConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Exchange string `mapstructure:"exchange"`
}

type ConsulConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	ServiceName string `mapstructure:"service_name"`
	ServiceID   string `mapstructure:"service_id"`
}

type JaegerConfig struct {
	AgentHost   string `mapstructure:"agent_host"`
	AgentPort   int    `mapstructure:"agent_port"`
	ServiceName string `mapstructure:"service_name"`
}

type Config struct {
	Server   ServerConfig `mapstructure:"server"`
	Database struct {
		MySQL MySQLConfig `mapstructure:"mysql"`
	} `mapstructure:"database"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	Consul   ConsulConfig   `mapstructure:"consul"`
	Jaeger   JaegerConfig   `mapstructure:"jaeger"`
}

func NewConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}
	return &config, nil
}
