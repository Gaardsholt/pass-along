package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/gob"
	"io"
)

func Encrypt(data interface{}, encryptionKey string) (encryptedSecret []byte, err error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(data)
	if err != nil {
		return
	}

	key := deriveKey(encryptionKey)

	c, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}

	encryptedSecret = gcm.Seal(nonce, nonce, buf.Bytes(), nil)
	return
}
