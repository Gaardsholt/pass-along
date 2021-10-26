package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

func Decrypt(data []byte, encryptionKey string) (decryptedData []byte, err error) {
	key := deriveKey(encryptionKey)

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("nonce mismatch")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	decryptedData, err = gcm.Open(nil, nonce, ciphertext, nil)

	return
}
