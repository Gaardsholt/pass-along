package redis

import (
	"sync"
	"time"

	"github.com/Gaardsholt/pass-along/metrics"
	"github.com/Gaardsholt/pass-along/types"
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

func (ss SecretStore) Add(entry types.Entry) (id string, err error) {
	expires := time.Now().Add(
		time.Hour*time.Duration(0) +
			time.Minute*time.Duration(0) +
			time.Second*time.Duration(entry.ExpiresIn),
	)

	mySecret := types.NewSecret(entry.Content, expires)
	mySecret.UnlimitedViews = entry.UnlimitedViews
	id = mySecret.GenerateID()
	baah, err := mySecret.Encrypt(id)
	if err != nil {
		metrics.SecretsCreatedWithError.Inc()
		return
	}

	conn := pool.Get()
	defer conn.Close()

	_, err = conn.Do("HMSET", id, "secret", baah)
	if err != nil {
		return "", err
	}

	_, err = conn.Do("EXPIRE", id, entry.ExpiresIn)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (ss SecretStore) Get(id string) (content string, gotData bool) {
	conn := pool.Get()
	defer conn.Close()

	secret, err := redis.Bytes(conn.Do("HGET", id, "secret"))
	if err != nil {
		return "", false
	}

	decryptedData, err := types.Decrypt(secret, id)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to decrypt secret")
		return "", false
	}

	ss.Delete(id)
	return decryptedData.Content, true
}

func (ss SecretStore) Delete(id string) {
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("HDEL", id, "secret")
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to delete secret with id %s", id)
	}
}
func (ss SecretStore) DeleteExpiredSecrets() {
	log.Debug().Msg("Not doing anything as redis will automatically delete expired secrets")
}
