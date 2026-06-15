package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// GlobalConfig holds config parameters
type GlobalConfig struct {
	ServerPort              *int    `required:"false" split_words:"true"`
	HealthPort              *int    `required:"false" split_words:"true"`
	ServerSalt              string  `required:"false" split_words:"true"`
	DatabaseType            *string `required:"false" split_words:"true" default:"in-memory"`
	RedisServer             *string `required:"false" split_words:"true"`
	RedisPort               *int    `required:"false" split_words:"true"`
	LogLevel                string  `required:"false" split_words:"true"`
	ValidForOptions         []int   `required:"false" split_words:"true" default:"3600,7200,43200,86400"`
	MaxSecretBytes          int     `required:"false" split_words:"true" default:"10485760"`
	MaxFiles                int     `required:"false" split_words:"true" default:"20"`
	MaxFileSizeBytes        int64   `required:"false" split_words:"true" default:"104857600"`
	EnableHSTS              bool    `required:"false" split_words:"true" default:"false"`
	HSTSMaxAgeSeconds       int     `required:"false" split_words:"true" default:"31536000"`
	GracefulShutdownSeconds int     `required:"false" split_words:"true" default:"25"`
	ReadinessDrainSeconds   int     `required:"false" split_words:"true" default:"5"`
	ShutdownHardSeconds     int     `required:"false" split_words:"true" default:"3"`
}

var Config GlobalConfig

// LoadConfig Loads config from env
func LoadConfig() {
	err := envconfig.Process("", &Config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	if err := validateConfig(); err != nil {
		errorCount := logValidationErrors(err)
		log.Fatal().Int("error_count", errorCount).Msg("Invalid config")
	}
	setupLogLevel()

}

func validateConfig() error {
	validationErrors := []error{}

	if Config.GetServerPort() == Config.GetHealthPort() {
		validationErrors = append(validationErrors, fmt.Errorf("SERVER_PORT and HEALTH_PORT must be different"))
	}

	if Config.MaxSecretBytes <= 0 {
		validationErrors = append(validationErrors, fmt.Errorf("MAX_SECRET_BYTES must be > 0"))
	}

	if Config.MaxFiles <= 0 {
		validationErrors = append(validationErrors, fmt.Errorf("MAX_FILES must be > 0"))
	}

	if Config.MaxFileSizeBytes <= 0 {
		validationErrors = append(validationErrors, fmt.Errorf("MAX_FILE_SIZE_BYTES must be > 0"))
	}

	if Config.EnableHSTS && Config.HSTSMaxAgeSeconds <= 0 {
		validationErrors = append(validationErrors, fmt.Errorf("HSTS_MAX_AGE_SECONDS must be > 0 when ENABLE_HSTS=true"))
	}

	if len(Config.ValidForOptions) == 0 {
		validationErrors = append(validationErrors, fmt.Errorf("VALID_FOR_OPTIONS must not be empty"))
	}

	seen := map[int]struct{}{}
	for _, v := range Config.ValidForOptions {
		if v <= 0 {
			validationErrors = append(validationErrors, fmt.Errorf("VALID_FOR_OPTIONS values must be > 0"))
		}
		if _, exists := seen[v]; exists {
			validationErrors = append(validationErrors, fmt.Errorf("VALID_FOR_OPTIONS must be unique"))
		}
		seen[v] = struct{}{}
	}

	if Config.GracefulShutdownSeconds <= 0 {
		validationErrors = append(validationErrors, fmt.Errorf("GRACEFUL_SHUTDOWN_SECONDS must be > 0"))
	}

	if Config.ReadinessDrainSeconds <= 0 {
		validationErrors = append(validationErrors, fmt.Errorf("READINESS_DRAIN_SECONDS must be > 0"))
	}

	if Config.ShutdownHardSeconds <= 0 {
		validationErrors = append(validationErrors, fmt.Errorf("SHUTDOWN_HARD_SECONDS must be > 0"))
	}

	if Config.GracefulShutdownSeconds <= Config.ReadinessDrainSeconds+Config.ShutdownHardSeconds {
		validationErrors = append(validationErrors, fmt.Errorf("GRACEFUL_SHUTDOWN_SECONDS must be greater than READINESS_DRAIN_SECONDS plus SHUTDOWN_HARD_SECONDS"))
	}

	return errors.Join(validationErrors...)
}

func logValidationErrors(err error) int {
	type joinedError interface {
		Unwrap() []error
	}

	joined, ok := err.(joinedError)
	if !ok {
		log.Error().Err(err).Msg("Config validation error")
		return 1
	}

	validationErrors := joined.Unwrap()
	for _, validationError := range validationErrors {
		log.Error().Err(validationError).Msg("Config validation error")
	}

	return len(validationErrors)
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

func (c GlobalConfig) GetGracefulShutdownTimeout() time.Duration {
	return time.Duration(c.GracefulShutdownSeconds) * time.Second
}

func (c GlobalConfig) GetHTTPShutdownTimeout() time.Duration {
	return time.Duration(c.GracefulShutdownSeconds-c.ReadinessDrainSeconds-c.ShutdownHardSeconds) * time.Second
}

func (c GlobalConfig) GetReadinessDrainDelay() time.Duration {
	return time.Duration(c.ReadinessDrainSeconds) * time.Second
}

func (c GlobalConfig) GetShutdownHardDelay() time.Duration {
	return time.Duration(c.ShutdownHardSeconds) * time.Second
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
