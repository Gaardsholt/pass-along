package crypto

import (
	"bytes"
	"crypto/sha512"
	"encoding/gob"

	"github.com/Gaardsholt/pass-along/config"
	"golang.org/x/crypto/pbkdf2"
)

func deriveKey(passphrase string) []byte {
	return pbkdf2.Key([]byte(passphrase), []byte(config.Config.ServerSalt), 1000, 32, sha512.New)
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
