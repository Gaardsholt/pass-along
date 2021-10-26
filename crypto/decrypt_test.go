package crypto

import (
	"testing"

	"gotest.tools/assert"
)

// TestDecryptAsExpected tests if a string is decrypted correct
func TestDecryptAsExpected(t *testing.T) {

	secretValue := "mysupersecretvalue"

	// arrange
	byteArray, err := GetBytes(secretValue)
	if err != nil {
		t.Error("encode error")
	}
	encryptedResult, err := Encrypt(secretValue, "encryptionkey")
	if err != nil {
		t.Error("encryption failed")
	}

	// act
	result, err := Decrypt(encryptedResult, "encryptionkey")
	if err != nil {
		t.Error("decryption failed")
	}
	t.Log(string(byteArray))
	t.Logf("%d", result)

	// assert
	assert.Equal(t, string(byteArray), string(result))
}
