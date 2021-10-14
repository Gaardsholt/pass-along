package redis

import (
	"fmt"
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

func New(lock *sync.RWMutex) SecretStore {

	// conn, err := redis.Dial("tcp", "localhost:6379")
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("cant freaking connect to redis")
	// }
	// defer conn.Close()
	// _, err = conn.Do("HMSET", "id:2", "secret", "Very secret Value")
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("cant freaking connect to redis")
	// }
	// secret, err := redis.String(conn.Do("HGET", "id:2", "secret"))
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("cant freaking connect to redis")
	// }
	// fmt.Println(secret)

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

	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_, err = conn.Do("HMSET", id, "secret", baah)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (ss SecretStore) Get(id string) (content string, gotData bool) {
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		return "", false
	}
	defer conn.Close()
	secret, err := redis.String(conn.Do("HGET", id, "secret"))
	if err != nil {
		log.Fatal().Err(err).Msg("cant freaking connect to redis")
		return "", false
	}
	return secret, true
}

func (ss SecretStore) Delete(id string) {
	fmt.Println("Deleting...")
}
