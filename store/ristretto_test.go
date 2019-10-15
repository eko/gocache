package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRistretto(t *testing.T) {
	// Given
	client := &MockRistrettoClientInterface{}
	options := &Options{
		Cost: 8,
	}

	// When
	store := NewRistretto(client, options)

	// Then
	assert.IsType(t, new(RistrettoStore), store)
	assert.Equal(t, client, store.client)
	assert.Equal(t, options, store.options)
}

func TestRistrettoGet(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := &MockRistrettoClientInterface{}
	client.On("Get", cacheKey).Return(cacheValue, true)

	store := NewRistretto(client, nil)

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
	options := &Options{
		Cost: 7,
	}

	client := &MockRistrettoClientInterface{}
	client.On("Set", cacheKey, cacheValue, int64(4)).Return(true)

	store := NewRistretto(client, options)

	// When
	err := store.Set(cacheKey, cacheValue, &Options{
		Cost: 4,
	})

	// Then
	assert.Nil(t, err)
}

func TestRistrettoGetType(t *testing.T) {
	// Given
	client := &MockRistrettoClientInterface{}

	store := NewRistretto(client, nil)

	// When - Then
	assert.Equal(t, RistrettoType, store.GetType())
}
