package config

import (
	"fmt"
	"log"

	"github.com/kelseyhightower/envconfig"
)

// GlobalConfig holds config parameters
type GlobalConfig struct {
	ServerSalt   string  `required:"false"`
	DatabaseType *string `required:"false" default:"in-memory"`
	RedisServer  *string `required:"false"`
	RedisPort    *int    `required:"false"`
}

var Config GlobalConfig

//LoadConfig Loads config from env
func LoadConfig() {
	err := envconfig.Process("", &Config)
	if err != nil {
		log.Fatal(err)
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
