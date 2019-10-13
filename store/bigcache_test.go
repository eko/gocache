package store

import (
	"testing"
	"time"

	mocksStore "github.com/eko/gache/test/mocks/store"
	"github.com/stretchr/testify/assert"
)

func TestNewBigcache(t *testing.T) {
	// Given
	client := &mocksStore.BigcacheClientInterface{}

	// When
	store := NewBigcache(client)

	// Then
	assert.IsType(t, new(BigcacheStore), store)
	assert.Equal(t, client, store.client)
}

func TestBigcacheGet(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := &mocksStore.BigcacheClientInterface{}
	client.On("Get", cacheKey).Return(cacheValue, nil)

	store := NewBigcache(client)

	// When
	value, err := store.Get(cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestBigcacheSet(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")
	expiration := 5 * time.Second

	client := &mocksStore.BigcacheClientInterface{}
	client.On("Set", cacheKey, cacheValue).Return(nil)

	store := NewBigcache(client)

	// When
	err := store.Set(cacheKey, cacheValue, expiration)

	// Then
	assert.Nil(t, err)
}

func TestBigcacheGetType(t *testing.T) {
	// Given
	client := &mocksStore.BigcacheClientInterface{}

	store := NewBigcache(client)

	// When - Then
	assert.Equal(t, BigcacheType, store.GetType())
}
