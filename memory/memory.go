package memory

import (
	"bytes"
	"encoding/gob"
	"log"
	"sync"
	"time"

	"github.com/Gaardsholt/pass-along/crypto"
	"github.com/Gaardsholt/pass-along/metrics"
	"github.com/Gaardsholt/pass-along/types"
)

type SecretStore struct {
	Data map[string][]byte
	Lock *sync.RWMutex
}

func new(content string, expires time.Time) types.Secret {
	return types.Secret{
		Content:   content,
		Expires:   expires,
		TimeAdded: time.Now(),
	}
}

func Decrypt(encryptedData []byte, encryptionKey string) (*types.Secret, error) {
	decryptedData, err := crypto.Decrypt(encryptedData, encryptionKey)
	if err != nil {
		return nil, err
	}

	var secret types.Secret
	dec := gob.NewDecoder(bytes.NewReader(decryptedData))
	err = dec.Decode(&secret)
	if err != nil {
		return nil, err
	}

	return &secret, nil
}

func (ss SecretStore) Add(entry types.Entry) (id string, err error) {
	expires := time.Now().Add(
		time.Hour*time.Duration(0) +
			time.Minute*time.Duration(0) +
			time.Second*time.Duration(entry.ExpiresIn),
	)

	mySecret := new(entry.Content, expires)
	mySecret.UnlimitedViews = entry.UnlimitedViews
	id = mySecret.GenerateID()

	baah, err := mySecret.Encrypt(id)
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
