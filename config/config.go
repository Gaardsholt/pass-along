package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type GlobalConfig struct {
	ServerSalt string `required:"false"`
}

var Config GlobalConfig

//LoadConfig Loads config from env
func LoadConfig() {
	configErr := envconfig.Process("", &Config)
	if configErr != nil {
		log.Fatal(configErr)
	}
}
