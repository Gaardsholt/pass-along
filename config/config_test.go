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
	os.Setenv("SERVERSALT", "somesalt")
	os.Setenv("DATABASETYPE", "redis")

	// act
	LoadConfig()

	// assert
	assert.Equal(t, "redis", *Config.DatabaseType)
	assert.Equal(t, "somesalt", Config.ServerSalt)
}

// TestLoadConfigDefaultDB tests if defaults work when no db env set
func TestLoadConfigDefaultDB(t *testing.T) {
	// arrange
	os.Clearenv()

	// act
	LoadConfig()

	// assert
	assert.Equal(t, "in-memory", *Config.DatabaseType)
}
