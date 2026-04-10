package types

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"strings"
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

func GenerateToken() (lookupID string, accessKey string, token string, err error) {
	lookupID, err = randomTokenPart(24)
	if err != nil {
		return "", "", "", err
	}

	accessKey, err = randomTokenPart(32)
	if err != nil {
		return "", "", "", err
	}

	return lookupID, accessKey, lookupID + "." + accessKey, nil
}

func ParseToken(token string) (lookupID string, accessKey string, err error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", "", errors.New("invalid token format")
	}

	lookupID = strings.TrimSpace(parts[0])
	accessKey = strings.TrimSpace(parts[1])

	if lookupID == "" || accessKey == "" {
		return "", "", errors.New("invalid token values")
	}

	if _, err = base64.RawURLEncoding.DecodeString(lookupID); err != nil {
		return "", "", errors.New("lookup id is not valid base64url")
	}

	if _, err = base64.RawURLEncoding.DecodeString(accessKey); err != nil {
		return "", "", errors.New("access key is not valid base64url")
	}

	return lookupID, accessKey, nil
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

func randomTokenPart(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
