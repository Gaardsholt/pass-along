package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/Gaardsholt/pass-along/config"
	"gotest.tools/assert"
)

func TestDeriveKeyMatchesExpectedVector(t *testing.T) {
	originalSalt := config.Config.ServerSalt
	config.Config.ServerSalt = "testsalt"
	defer func() {
		config.Config.ServerSalt = originalSalt
	}()

	key, err := deriveKey("encryptionkey")
	assert.NilError(t, err)
	assert.Equal(t, "0e4951ca36c04f0ef60971be5b0e6cd342bb5922a94869659e960fcebe32b2ec", hex.EncodeToString(key))
}
