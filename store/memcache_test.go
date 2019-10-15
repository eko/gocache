package store

import (
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/stretchr/testify/assert"
)

func TestNewMemcache(t *testing.T) {
	// Given
	client := &MockMemcacheClientInterface{}
	options := &Options{Expiration: 3*time.Second}

	// When
	store := NewMemcache(client, options)

	// Then
	assert.IsType(t, new(MemcacheStore), store)
	assert.Equal(t, client, store.client)
	assert.Equal(t, options, store.options)
}

func TestMemcacheGet(t *testing.T) {
	// Given
	options := &Options{Expiration: 3*time.Second}

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := &MockMemcacheClientInterface{}
	client.On("Get", cacheKey).Return(&memcache.Item{
		Value: cacheValue,
	}, nil)

	store := NewMemcache(client, options)

	// When
	value, err := store.Get(cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestMemcacheSet(t *testing.T) {
	// Given
	options := &Options{Expiration: 3*time.Second}

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := &MockMemcacheClientInterface{}
	client.On("Set", &memcache.Item{
		Key:        cacheKey,
		Value:      cacheValue,
		Expiration: int32(5),
	}).Return(nil)

	store := NewMemcache(client, options)

	// When
	err := store.Set(cacheKey, cacheValue, &Options{
		Expiration: 5 * time.Second,
	})

	// Then
	assert.Nil(t, err)
}

func TestMemcacheGetType(t *testing.T) {
	// Given
	client := &MockMemcacheClientInterface{}

	store := NewMemcache(client, nil)

	// When - Then
	assert.Equal(t, MemcacheType, store.GetType())
}
