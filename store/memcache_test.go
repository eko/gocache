package store

import (
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	mocksStore "github.com/eko/gache/test/mocks/store"
	"github.com/stretchr/testify/assert"
)

func TestNewMemcache(t *testing.T) {
	// Given
	client := &mocksStore.MemcacheClientInterface{}

	// When
	store := NewMemcache(client)

	// Then
	assert.IsType(t, new(MemcacheStore), store)
	assert.Equal(t, client, store.client)
}

func TestMemcacheGet(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := &mocksStore.MemcacheClientInterface{}
	client.On("Get", cacheKey).Return(&memcache.Item{
		Value: cacheValue,
	}, nil)

	store := NewMemcache(client)

	// When
	value, err := store.Get(cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestMemcacheSet(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")
	expiration := 5 * time.Second

	client := &mocksStore.MemcacheClientInterface{}
	client.On("Set", &memcache.Item{
		Key:        cacheKey,
		Value:      cacheValue,
		Expiration: int32(expiration.Seconds()),
	}).Return(nil)

	store := NewMemcache(client)

	// When
	err := store.Set(cacheKey, cacheValue, expiration)

	// Then
	assert.Nil(t, err)
}

func TestMemcacheGetType(t *testing.T) {
	// Given
	client := &mocksStore.MemcacheClientInterface{}

	store := NewMemcache(client)

	// When - Then
	assert.Equal(t, MemcacheType, store.GetType())
}
