package types

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Gaardsholt/pass-along/config"
	"github.com/Gaardsholt/pass-along/metrics"
	"golang.org/x/crypto/pbkdf2"
)

type Secret struct {
	Content        string    `json:"content"`
	Expires        time.Time `json:"expires"`
	TimeAdded      time.Time `json:"time_added"`
	UnlimitedViews bool      `json:"unlimited_views"`
}

func new(content string, expires time.Time) Secret {
	return Secret{
		Content:   content,
		Expires:   expires,
		TimeAdded: time.Now(),
	}
}

func (s Secret) hash() string {
	checksum := sha512.Sum512([]byte(fmt.Sprintf("%v", s)))
	hash := base64.RawURLEncoding.EncodeToString(checksum[:])
	return hash
}

func deriveKey(passphrase string) []byte {
	return pbkdf2.Key([]byte(passphrase), []byte(config.Config.ServerSalt), 1000, 32, sha512.New)
}

func (s Secret) encrypt(encryptionKey string) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(s)
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

	encryptedSecret := gcm.Seal(nonce, nonce, buf.Bytes(), nil)

	return encryptedSecret, nil
}
func Decrypt(ciphertext []byte, encryptionKey string) (*Secret, error) {
	key := deriveKey(encryptionKey)

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, err
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	p := Secret{}
	dec := gob.NewDecoder(bytes.NewReader(plaintext))
	err = dec.Decode(&p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (ss SecretStore) Add(entry Entry) (id string, err error) {
	expires := time.Now().Add(
		time.Hour*time.Duration(0) +
			time.Minute*time.Duration(0) +
			time.Second*time.Duration(entry.ExpiresIn),
	)

	mySecret := new(entry.Content, expires)
	mySecret.UnlimitedViews = entry.UnlimitedViews
	id = mySecret.hash()

	baah, err := mySecret.encrypt(id)
	if err != nil {
		metrics.SecretsCreatedWithError.Inc()
		return
	}

	ss.Lock.Lock()
	defer ss.Lock.Unlock()
	ss.Data[id] = baah

	metrics.SecretsCreated.Inc()
	return
}
func (ss SecretStore) Get(id string) (content string, gotData bool) {
	ss.Lock.RLock()
	value, gotData := ss.Data[id]
	ss.Lock.RUnlock()
	if gotData {
		s, err := Decrypt(value, id)
		if err != nil {
			log.Fatal(err)
		}

		isNotExpired := s.Expires.UTC().After(time.Now().UTC())
		if isNotExpired {
			content = s.Content
			metrics.SecretsRead.Inc()
		} else {
			gotData = false
			metrics.ExpiredSecretsRead.Inc()
		}

		if !isNotExpired || !s.UnlimitedViews {
			ss.Delete(id)
		}
		return
	}
	metrics.NonExistentSecretsRead.Inc()

	return
}
func (ss SecretStore) Delete(id string) {
	ss.Lock.Lock()
	defer ss.Lock.Unlock()

	delete(ss.Data, id)
	metrics.SecretsDeleted.Inc()
}
