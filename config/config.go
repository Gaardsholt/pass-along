package config

import (
	"fmt"
	"log"

	"github.com/kelseyhightower/envconfig"
)

type GlobalConfig struct {
	ServerSalt   string  `required:"false"`
	DatabaseType *string `required:"false" default:"in-memory"`
}

var Config GlobalConfig

//LoadConfig Loads config from env
func LoadConfig() {
	configErr := envconfig.Process("", &Config)
	if configErr != nil {
		log.Fatal(configErr)
	}
}

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
