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
	key, err := deriveKey(encryptionKey, config.Config.ServerSalt)
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

	encryptedSecret = gcm.Seal(nonce, nonce, byteArray, nil)
	return
}
