package crypto

import (
	"crypto/aes"
	"crypto/cipher"
)

func Decrypt(data []byte, encryptionKey string) (decryptedData []byte, err error) {
	key := deriveKey(encryptionKey)

	c, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	decryptedData, err = gcm.Open(nil, nonce, ciphertext, nil)

	return
}
