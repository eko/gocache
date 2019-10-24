package store

import (
	"testing"
	"time"

	mocksStore "github.com/eko/gocache/test/mocks/store/clients"
	"github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/assert"
)

func TestNewRedis(t *testing.T) {
	// Given
	client := &mocksStore.RedisClientInterface{}
	options := &Options{
		Expiration: 6 * time.Second,
	}

	// When
	store := NewRedis(client, options)

	// Then
	assert.IsType(t, new(RedisStore), store)
	assert.Equal(t, client, store.client)
	assert.Equal(t, options, store.options)
}

func TestRedisGet(t *testing.T) {
	// Given
	client := &mocksStore.RedisClientInterface{}
	client.On("Get", "my-key").Return(&redis.StringCmd{})

	store := NewRedis(client, nil)

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
	options := &Options{
		Expiration: 6 * time.Second,
	}

	client := &mocksStore.RedisClientInterface{}
	client.On("Set", "my-key", cacheValue, 5*time.Second).Return(&redis.StatusCmd{})

	store := NewRedis(client, options)

	// When
	err := store.Set(cacheKey, cacheValue, &Options{
		Expiration: 5 * time.Second,
	})

	// Then
	assert.Nil(t, err)
}

func TestRedisSetWhenNoOptionsGiven(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := "my-cache-value"
	options := &Options{
		Expiration: 6 * time.Second,
	}

	client := &mocksStore.RedisClientInterface{}
	client.On("Set", "my-key", cacheValue, 6*time.Second).Return(&redis.StatusCmd{})

	store := NewRedis(client, options)

	// When
	err := store.Set(cacheKey, cacheValue, nil)

	// Then
	assert.Nil(t, err)
}

func TestRedisSetWithTags(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := &mocksStore.RedisClientInterface{}
	client.On("Set", cacheKey, cacheValue, time.Duration(0)).Return(&redis.StatusCmd{})
	client.On("Get", "gocache_tag_tag1").Return(&redis.StringCmd{})
	client.On("Set", "gocache_tag_tag1", []byte("my-key"), 720*time.Hour).Return(&redis.StatusCmd{})

	store := NewRedis(client, nil)

	// When
	err := store.Set(cacheKey, cacheValue, &Options{Tags: []string{"tag1"}})

	// Then
	assert.Nil(t, err)
}

func TestRedisDelete(t *testing.T) {
	// Given
	cacheKey := "my-key"

	client := &mocksStore.RedisClientInterface{}
	client.On("Del", "my-key").Return(&redis.IntCmd{})

	store := NewRedis(client, nil)

	// When
	err := store.Delete(cacheKey)

	// Then
	assert.Nil(t, err)
}

func TestRedisInvalidate(t *testing.T) {
	// Given
	options := InvalidateOptions{
		Tags: []string{"tag1"},
	}

	cacheKeys := &redis.StringCmd{}

	client := &mocksStore.RedisClientInterface{}
	client.On("Get", "gocache_tag_tag1").Return(cacheKeys, nil)

	store := NewRedis(client, nil)

	// When
	err := store.Invalidate(options)

	// Then
	assert.Nil(t, err)
}

func TestRedisClear(t *testing.T) {
	// Given
	client := &mocksStore.RedisClientInterface{}
	client.On("FlushAll").Return(&redis.StatusCmd{})

	store := NewRedis(client, nil)

	// When
	err := store.Clear()

	// Then
	assert.Nil(t, err)
}

func TestRedisGetType(t *testing.T) {
	// Given
	client := &mocksStore.RedisClientInterface{}

	store := NewRedis(client, nil)

	// When - Then
	assert.Equal(t, RedisType, store.GetType())
}
