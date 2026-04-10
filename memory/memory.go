package memory

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/Gaardsholt/pass-along/metrics"
)

type SecretStore struct {
	Data    map[string][]byte
	Expires map[string]time.Time
	Lock    *sync.RWMutex
}

func New(lock *sync.RWMutex) (SecretStore, error) {
	return SecretStore{
		Data:    make(map[string][]byte),
		Expires: make(map[string]time.Time),
		Lock:    lock,
	}, nil
}

func (ss SecretStore) Add(id string, secret []byte, expiresIn int) error {
	ss.Lock.Lock()
	defer ss.Lock.Unlock()
	ss.Data[id] = secret
	ss.Expires[id] = time.Now().UTC().Add(time.Duration(expiresIn) * time.Second)

	go metrics.SecretsCreated.Inc()
	return nil
}

func (ss SecretStore) Get(id string) (secret []byte, gotData bool) {
	ss.Lock.Lock()
	defer ss.Lock.Unlock()

	expiration, hasExpiration := ss.Expires[id]
	if hasExpiration && !expiration.After(time.Now().UTC()) {
		delete(ss.Data, id)
		delete(ss.Expires, id)
		go metrics.SecretsDeleted.Inc()
		return nil, false
	}
	secret, gotData = ss.Data[id]
	return
}

func (ss SecretStore) Delete(id string) {
	ss.Lock.Lock()
	defer ss.Lock.Unlock()

	delete(ss.Data, id)
	delete(ss.Expires, id)
	go metrics.SecretsDeleted.Inc()
}

func (ss SecretStore) DeleteExpiredSecrets() {
	for {
		time.Sleep(5 * time.Minute)
		now := time.Now().UTC()
		ss.Lock.RLock()
		expired := []string{}
		for k, expiresAt := range ss.Expires {
			if !expiresAt.After(now) {
				expired = append(expired, k)
			}
		}
		ss.Lock.RUnlock()

		for _, k := range expired {
			log.Debug().Msg("Found expired secret, deleting...")
			ss.Delete(k)
		}
	}
}
