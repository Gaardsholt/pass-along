package crypto

import (
	"testing"

	"gotest.tools/assert"
)

// TestEncryptAsExpected tests if a string is encrypted, or tries to
func TestEncryptAsExpected(t *testing.T) {
	// arrange
	data := []byte("mysupersecretvalue")

	// act
	result, err := Encrypt(data, "encryptionkey")
	if err != nil {
		t.Error("encryption failed")
	}

	// assert
	assert.Assert(t, string(result) != string(data))
}
