package config

import (
	"os"
	"testing"

	"gotest.tools/assert"
)

// TestLoadConfigAsExpected tests if the config is loaded as expected
func TestLoadConfigAsExpected(t *testing.T) {
	// arrange
	os.Clearenv()
	os.Setenv("SERVER_SECRET", "0123456789abcdef0123456789abcdef")
	os.Setenv("DATABASE_TYPE", "redis")

	// act
	LoadConfig()

	// assert
	assert.Equal(t, "redis", *Config.DatabaseType)
	assert.Equal(t, "0123456789abcdef0123456789abcdef", Config.ServerSecret)
}

// TestLoadConfigDefaultDB tests if defaults work when no db env set
func TestLoadConfigDefaultDB(t *testing.T) {
	// arrange
	os.Clearenv()
	os.Setenv("SERVER_SECRET", "0123456789abcdef0123456789abcdef")

	// act
	LoadConfig()

	// assert
	assert.Equal(t, "in-memory", *Config.DatabaseType)
}
