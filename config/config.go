package config

import (
	"fmt"
	"slices"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// GlobalConfig holds config parameters
type GlobalConfig struct {
	ServerPort        *int    `required:"false" split_words:"true"`
	HealthPort        *int    `required:"false" split_words:"true"`
	ServerSalt        string  `required:"false" split_words:"true"`
	DatabaseType      *string `required:"false" split_words:"true" default:"in-memory"`
	RedisServer       *string `required:"false" split_words:"true"`
	RedisPort         *int    `required:"false" split_words:"true"`
	LogLevel          string  `required:"false" split_words:"true"`
	ValidForOptions   []int   `required:"false" split_words:"true" default:"3600,7200,43200,86400"`
	MaxSecretBytes    int     `required:"false" split_words:"true" default:"10485760"`
	MaxFiles          int     `required:"false" split_words:"true" default:"20"`
	MaxFileSizeBytes  int64   `required:"false" split_words:"true" default:"104857600"`
	EnableHSTS        bool    `required:"false" split_words:"true" default:"false"`
	HSTSMaxAgeSeconds int     `required:"false" split_words:"true" default:"31536000"`
}

var Config GlobalConfig

// LoadConfig Loads config from env
func LoadConfig() {
	err := envconfig.Process("", &Config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	if Config.GetServerPort() == Config.GetHealthPort() {
		log.Fatal().Err(nil).Msg("SERVER_PORT and HEALTH_PORT must be different")
	}

	validateConfig()
	setupLogLevel()

}

func validateConfig() {
	if Config.MaxSecretBytes <= 0 {
		log.Fatal().Err(nil).Msg("MAX_SECRET_BYTES must be > 0")
	}

	if Config.MaxFiles <= 0 {
		log.Fatal().Err(nil).Msg("MAX_FILES must be > 0")
	}

	if Config.MaxFileSizeBytes <= 0 {
		log.Fatal().Err(nil).Msg("MAX_FILE_SIZE_BYTES must be > 0")
	}

	if Config.EnableHSTS && Config.HSTSMaxAgeSeconds <= 0 {
		log.Fatal().Err(nil).Msg("HSTS_MAX_AGE_SECONDS must be > 0 when ENABLE_HSTS=true")
	}

	if len(Config.ValidForOptions) == 0 {
		log.Fatal().Err(nil).Msg("VALID_FOR_OPTIONS must not be empty")
	}

	seen := map[int]struct{}{}
	for _, v := range Config.ValidForOptions {
		if v <= 0 {
			log.Fatal().Err(nil).Msg("VALID_FOR_OPTIONS values must be > 0")
		}
		if _, exists := seen[v]; exists {
			log.Fatal().Err(nil).Msg("VALID_FOR_OPTIONS must be unique")
		}
		seen[v] = struct{}{}
	}
}

func (c GlobalConfig) IsValidExpiration(seconds int) bool {
	return slices.Contains(c.ValidForOptions, seconds)
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

func setupLogLevel() {
	// default is info
	switch strings.ToLower(Config.LogLevel) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
