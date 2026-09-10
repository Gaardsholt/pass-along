package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/Gaardsholt/pass-along/config"
)

func Encrypt(data interface{}, encryptionKey string) (encryptedSecret []byte, err error) {
	byteArray, err := getBytes(data)
	if err != nil {
		return nil, err
	}

	return encryptWithSalt(byteArray, encryptionKey, config.Config.ServerSalt)
}

func encryptWithSalt(data []byte, encryptionKey, salt string) (encryptedSecret []byte, err error) {
	key, err := deriveKey(encryptionKey, salt)
	if err != nil {
		return nil, err
	}

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

	encryptedSecret = gcm.Seal(nonce, nonce, data, nil)
	return
}
