package memory

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/Gaardsholt/pass-along/metrics"
	"github.com/Gaardsholt/pass-along/types"
)

type SecretStore struct {
	Data map[string][]byte
	Lock *sync.RWMutex
}

func NewStore(lock *sync.RWMutex) SecretStore {
	return SecretStore{
		Data: make(map[string][]byte),
		Lock: lock,
	}
}

func (ss SecretStore) Add(id string, secret []byte, expiresIn int) error {
	ss.Lock.Lock()
	defer ss.Lock.Unlock()
	ss.Data[id] = secret

	metrics.SecretsCreated.Inc()
	return nil
}

func (ss SecretStore) Get(id string) (secret []byte, gotData bool) {
	ss.Lock.RLock()
	secret, gotData = ss.Data[id]
	ss.Lock.RUnlock()
	return
}

func (ss SecretStore) Delete(id string) {
	ss.Lock.Lock()
	defer ss.Lock.Unlock()

	delete(ss.Data, id)
	metrics.SecretsDeleted.Inc()
}

func (ss SecretStore) DeleteExpiredSecrets() {
	for {
		time.Sleep(5 * time.Minute)
		ss.Lock.RLock()
		for k, v := range ss.Data {
			s, err := types.Decrypt(v, k)
			if err != nil {
				continue
			}

			isNotExpired := s.Expires.UTC().After(time.Now().UTC())
			if !isNotExpired {
				log.Debug().Msg("Found expired secret, deleting...")
				ss.Lock.RUnlock()
				ss.Delete(k)
				ss.Lock.RLock()
			}
		}
		ss.Lock.RUnlock()
	}
}
