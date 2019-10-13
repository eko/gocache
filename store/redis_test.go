package store

import (
	"testing"
	"time"

	mocksStore "github.com/eko/gache/test/mocks/store"
	"github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/assert"
)

func TestNewRedis(t *testing.T) {
	// Given
	client := &mocksStore.RedisClientInterface{}

	// When
	store := NewRedis(client)

	// Then
	assert.IsType(t, new(RedisStore), store)
	assert.Equal(t, client, store.client)
}

func TestRedisGet(t *testing.T) {
	// Given
	client := &mocksStore.RedisClientInterface{}
	client.On("Get", "my-key").Return(&redis.StringCmd{})

	store := NewRedis(client)

	// When
	value, err := store.Get("my-key")

	// Then
	assert.Nil(t, err)
	assert.NotNil(t, value)
}

func TestRedisSet(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := "my-cache-value"
	expiration := 5 * time.Second

	client := &mocksStore.RedisClientInterface{}
	client.On("Set", "my-key", cacheValue, expiration).Return(&redis.StatusCmd{})

	store := NewRedis(client)

	// When
	err := store.Set(cacheKey, cacheValue, expiration)

	// Then
	assert.Nil(t, err)
}

func TestRedisGetType(t *testing.T) {
	// Given
	client := &mocksStore.RedisClientInterface{}

	store := NewRedis(client)

	// When - Then
	assert.Equal(t, RedisType, store.GetType())
}
