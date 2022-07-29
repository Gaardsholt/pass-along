package types

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"math/rand"
	"time"

	"github.com/Gaardsholt/pass-along/crypto"
)

type Secret struct {
	Content        string            `json:"content"`
	Files          map[string][]byte `json:"files"`
	Expires        time.Time         `json:"expires"`
	TimeAdded      time.Time         `json:"time_added"`
	UnlimitedViews bool              `json:"unlimited_views"`
}

func NewSecret(content string, expires time.Time) Secret {
	return Secret{
		Content:   content,
		Expires:   expires,
		TimeAdded: time.Now(),
	}
}

func (s Secret) GenerateID() string {
	random := randomString(30)
	checksum := sha512.Sum512([]byte(fmt.Sprintf("%v%v", s, random)))
	hash := base64.RawURLEncoding.EncodeToString(checksum[:])
	return hash
}

func (s Secret) Encrypt(encryptionKey string) ([]byte, error) {
	return crypto.Encrypt(s, encryptionKey)
}

func Decrypt(encryptedData []byte, encryptionKey string) (*Secret, error) {
	decryptedData, err := crypto.Decrypt(encryptedData, encryptionKey)
	if err != nil {
		return nil, err
	}

	var secret Secret
	dec := gob.NewDecoder(bytes.NewReader(decryptedData))
	err = dec.Decode(&secret)
	if err != nil {
		return nil, err
	}

	return &secret, nil
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}
