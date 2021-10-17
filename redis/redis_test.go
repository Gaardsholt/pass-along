package redis

import (
	"sync"
	"testing"

	"github.com/Gaardsholt/pass-along/datastore"
	"github.com/Gaardsholt/pass-along/types"
	"github.com/alicebob/miniredis/v2"
	"github.com/gomodule/redigo/redis"
	"gotest.tools/assert"
)

var secretStore datastore.SecretStore
var lock = sync.RWMutex{}

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
		Lock: &lock,
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

	// act
	id, err := secretStore.Add(entry)
	if err != nil {
		t.Error(err)
	}

	content, gotData := secretStore.Get(id)
	if gotData == false {
		t.Error("no data recieved from redis")
	}

	// assert
	assert.Equal(t, content, entry.Content)
}
