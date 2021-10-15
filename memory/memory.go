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

func new(content string, expires time.Time) types.Secret {
	return types.Secret{
		Content:   content,
		Expires:   expires,
		TimeAdded: time.Now(),
	}
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
		s, err := types.Decrypt(value, id)
		if err != nil {
			log.Fatal().Err(err).Msg("Unable to decrypt secret")
			return "", false
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
