package redis

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/Gaardsholt/pass-along/types"
	"github.com/alicebob/miniredis/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog/log"
	"gotest.tools/assert"
)

var secretStore SecretStore

// GetTestRedisServer creates the test server
func GetTestRedisServer(t *testing.T) SecretStore {
	t.Helper()

	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	pool = &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000, // max number of connections
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", s.Addr())
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}

	return SecretStore{
		Data: make(map[string][]byte),
	}

}

// TestAddAsExpected tests if the config is loaded as expected
func TestAddAsExpected(t *testing.T) {
	// arrange
	secretStore := GetTestRedisServer(t)
	entry := types.Entry{
		Content:        "supersecretvalue",
		ExpiresIn:      1,
		UnlimitedViews: false,
	}

	var byteArray bytes.Buffer
	err := gob.NewEncoder(&byteArray).Encode(entry)
	if err != nil {
		log.Fatal().Err(err).Msg("encode error")
	}

	id := "1"

	// act
	err = secretStore.Add(id, byteArray.Bytes(), entry.ExpiresIn)
	if err != nil {
		t.Error(err)
	}

	content, gotData := secretStore.Get(id)
	if gotData == false {
		t.Error("no data recieved from redis")
	}

	// assert
	assert.Equal(t, string(content), byteArray.String())
}
