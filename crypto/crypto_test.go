package crypto

import (
	"encoding/hex"
	"testing"

	"gotest.tools/assert"
)

func TestDeriveKeyMatchesExpectedVector(t *testing.T) {
	key, err := deriveKey("encryptionkey", "testsalt")
	assert.NilError(t, err)
	assert.Equal(t, "0e4951ca36c04f0ef60971be5b0e6cd342bb5922a94869659e960fcebe32b2ec", hex.EncodeToString(key))
}
