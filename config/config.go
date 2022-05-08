package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

// GlobalConfig holds config parameters
type GlobalConfig struct {
	ServerPort   *int    `required:"false" split_words:"true"`
	HealthPort   *int    `required:"false" split_words:"true"`
	ServerSalt   string  `required:"false" split_words:"true"`
	DatabaseType *string `required:"false" split_words:"true" default:"in-memory"`
	RedisServer  *string `required:"false" split_words:"true"`
	RedisPort    *int    `required:"false" split_words:"true"`
}

var Config GlobalConfig

//LoadConfig Loads config from env
func LoadConfig() {
	err := envconfig.Process("", &Config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	if Config.GetServerPort() == Config.GetHealthPort() {
		log.Fatal().Err(nil).Msg("SERVER_PORT and HEALTH_PORT must be different")
	}

}

// GetDatabaseType determines if a correct db is set
func (c GlobalConfig) GetDatabaseType() (string, error) {
	switch *c.DatabaseType {
	case "in-memory":
		return *c.DatabaseType, nil
	case "redis":
		return *c.DatabaseType, nil
	default:
		return "", fmt.Errorf("unknown database type")
	}
}

func (c GlobalConfig) GetRedisServer() string {
	if c.RedisServer != nil {
		return *c.RedisServer
	}
	return "localhost"
}

func (c GlobalConfig) GetRedisPort() int {
	if c.RedisPort != nil {
		return *c.RedisPort
	}
	return 6379
}

func (c GlobalConfig) GetServerPort() int {
	if c.ServerPort != nil {
		return *c.ServerPort
	}
	return 8080
}

func (c GlobalConfig) GetHealthPort() int {
	if c.HealthPort != nil {
		return *c.HealthPort
	}
	return 8888
}
