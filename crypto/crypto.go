package crypto

import (
	"bytes"
	"crypto/sha512"
	"encoding/gob"
	"fmt"

	"github.com/Gaardsholt/pass-along/config"
	"golang.org/x/crypto/pbkdf2"
)

func deriveKey(passphrase string) []byte {
	salt := fmt.Sprintf("pass-along-v2:%s", config.Config.ServerSecret)
	return pbkdf2.Key([]byte(passphrase), []byte(salt), config.Config.KDFIterations, 32, sha512.New)
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
