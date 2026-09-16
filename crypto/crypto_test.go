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

func TestEncryptDecryptRoundTripWithExplicitSalt(t *testing.T) {
	expectedBytes, err := getBytes("mysupersecretvalue")
	assert.NilError(t, err)

	encryptedData, err := encryptWithSalt(expectedBytes, "encryptionkey", "testsalt")
	assert.NilError(t, err)

	decryptedData, err := decryptWithSalt(encryptedData, "encryptionkey", "testsalt")
	assert.NilError(t, err)
	assert.Equal(t, string(expectedBytes), string(decryptedData))
}
