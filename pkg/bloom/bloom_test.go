package bloom

import (
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisBitSet_Add(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	defer client.Close()

	filter := New(client, "test_key", 1024)
	assert.Nil(t, filter.Add([]byte("hello")))
	assert.Nil(t, filter.Add([]byte("world")))
	ok, err := filter.Exists([]byte("hello"))
	assert.Nil(t, err)
	assert.True(t, ok)
}
