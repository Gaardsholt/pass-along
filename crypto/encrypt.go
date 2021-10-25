package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func Encrypt(data interface{}, encryptionKey string) (encryptedSecret []byte, err error) {
	byteArray, err := GetBytes(data)
	if err != nil {
		return nil, err
	}
	key := deriveKey(encryptionKey)

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	encryptedSecret = gcm.Seal(nonce, nonce, byteArray, nil)
	return
}
