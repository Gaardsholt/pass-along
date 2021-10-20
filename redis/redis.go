package redis

import (
	"sync"

	"github.com/Gaardsholt/pass-along/metrics"
	"github.com/rs/zerolog/log"

	"github.com/gomodule/redigo/redis"
)

type SecretStore struct {
	Data map[string][]byte
	Lock *sync.RWMutex
}

var pool *redis.Pool

func NewStore(lock *sync.RWMutex) SecretStore {
	pool = &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "localhost:6379")
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}

	return SecretStore{
		Data: make(map[string][]byte),
		Lock: lock,
	}
}

func (ss SecretStore) Add(id string, secret []byte, expiresIn int) error {
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("HMSET", id, "secret", secret)
	if err != nil {
		metrics.SecretsCreatedWithError.Inc()
		return err
	}

	_, err = conn.Do("EXPIRE", id, expiresIn)
	if err != nil {
		metrics.SecretsCreatedWithError.Inc()
		return err
	}
	metrics.SecretsCreated.Inc()
	return nil
}

func (ss SecretStore) Get(id string) (secret []byte, gotData bool) {
	conn := pool.Get()
	defer conn.Close()

	secret, err := redis.Bytes(conn.Do("HGET", id, "secret"))
	if err != nil {
		metrics.NonExistentSecretsRead.Inc()
		return nil, false
	}
	return secret, true
}

func (ss SecretStore) Delete(id string) {
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("HDEL", id, "secret")
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to delete secret with id %s", id)
	}
	metrics.SecretsDeleted.Inc()
}

func (ss SecretStore) DeleteExpiredSecrets() {
	log.Debug().Msg("Not doing anything as redis will automatically delete expired secrets")
}
