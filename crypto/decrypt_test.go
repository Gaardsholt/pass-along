package crypto

import (
	"testing"

	"gotest.tools/assert"
)

// TestDecryptAsExpected tests if a string is decrypted correct
func TestDecryptAsExpected(t *testing.T) {
	// arrange
	data := []byte("mysupersecretvalue")
	encryptedResult, err := Encrypt(data, "encryptionkey")
	if err != nil {
		t.Error("encryption failed")
	}

	// act
	result, err := Decrypt(encryptedResult, "encryptionkey")
	if err != nil {
		t.Error("decryption failed")
	}
	t.Log(string(data))
	t.Logf("%d", result)

	// assert
	assert.Equal(t, string(data), string(result))
}
