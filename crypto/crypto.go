package crypto

import (
	"bytes"
	"crypto/pbkdf2"
	"crypto/sha512"
	"encoding/gob"
)

func deriveKey(passphrase, salt string) ([]byte, error) {
	return pbkdf2.Key(sha512.New, passphrase, []byte(salt), 300000, 32)
}

func getBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
