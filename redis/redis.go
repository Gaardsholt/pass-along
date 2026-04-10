package redis

import (
	"fmt"
	"sync"

	"github.com/Gaardsholt/pass-along/config"
	"github.com/Gaardsholt/pass-along/metrics"
	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog/log"
)

type SecretStore struct {
	Data map[string][]byte
	Lock *sync.RWMutex
}

var pool *redis.Pool

const logIDPrefixLength = 8

func New() (ss SecretStore, err error) {

	server := config.Config.GetRedisServer()
	port := config.Config.GetRedisPort()

	pool = &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000, // max number of connections
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", fmt.Sprintf("%s:%d", server, port))
		},
	}

	conn := pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		return SecretStore{}, err
	}

	ss = SecretStore{
		Data: make(map[string][]byte),
	}

	return ss, nil
}

func (ss SecretStore) Add(id string, secret []byte, expiresIn int) error {
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("HMSET", id, "secret", secret)
	if err != nil {
		go metrics.SecretsCreatedWithError.Inc()
		return err
	}

	_, err = conn.Do("EXPIRE", id, expiresIn)
	if err != nil {
		go metrics.SecretsCreatedWithError.Inc()
		return err
	}
	go metrics.SecretsCreated.Inc()
	return nil
}

func (ss SecretStore) Get(id string) (secret []byte, gotData bool) {
	conn := pool.Get()
	defer conn.Close()

	secret, err := redis.Bytes(conn.Do("HGET", id, "secret"))
	if err != nil {
		go metrics.NonExistentSecretsRead.Inc()
		return nil, false
	}
	return secret, true
}

func (ss SecretStore) Delete(id string) {
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", id)
	if err != nil {
		idPrefix := id
		if len(idPrefix) > logIDPrefixLength {
			idPrefix = idPrefix[:logIDPrefixLength]
		}
		log.Error().Err(err).Str("id_prefix", idPrefix).Msg("Failed to delete secret")
		return
	}
	go metrics.SecretsDeleted.Inc()
}

func (ss SecretStore) DeleteExpiredSecrets() {
	log.Debug().Msg("Not doing anything as redis will automatically delete expired secrets")
}
