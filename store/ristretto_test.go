package store

import (
	"testing"
	"time"

	mocksStore "github.com/eko/gache/test/mocks/store"
	"github.com/stretchr/testify/assert"
)

func TestNewRistretto(t *testing.T) {
	// Given
	client := &mocksStore.RistrettoClientInterface{}

	// When
	store := NewRistretto(client)

	// Then
	assert.IsType(t, new(RistrettoStore), store)
	assert.Equal(t, client, store.client)
}

func TestRistrettoGet(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := &mocksStore.RistrettoClientInterface{}
	client.On("Get", cacheKey).Return(cacheValue, true)

	store := NewRistretto(client)

	// When
	value, err := store.Get(cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestRistrettoSet(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := "my-cache-value"
	expiration := 5 * time.Second

	client := &mocksStore.RistrettoClientInterface{}
	client.On("Set", cacheKey, cacheValue, int64(1)).Return(true)

	store := NewRistretto(client)

	// When
	err := store.Set(cacheKey, cacheValue, expiration)

	// Then
	assert.Nil(t, err)
}

func TestRistrettoGetType(t *testing.T) {
	// Given
	client := &mocksStore.RistrettoClientInterface{}

	store := NewRistretto(client)

	// When - Then
	assert.Equal(t, RistrettoType, store.GetType())
}
