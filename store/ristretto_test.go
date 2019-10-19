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

func TestRistrettoSetWithTags(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := &MockRistrettoClientInterface{}
	client.On("Set", cacheKey, cacheValue, int64(0)).Return(true)
	client.On("Get", "gocache_tag_tag1").Return(nil, true)
	client.On("Set", "gocache_tag_tag1", []byte("my-key"), int64(0)).Return(true)

	store := NewRistretto(client, nil)

	// When
	err := store.Set(cacheKey, cacheValue, &Options{Tags: []string{"tag1"}})

	// Then
	assert.Nil(t, err)
}

func TestRistrettoDelete(t *testing.T) {
	// Given
	cacheKey := "my-key"

	client := &MockRistrettoClientInterface{}
	client.On("Del", cacheKey).Return(nil)

	store := NewRistretto(client, nil)

	// When
	err := store.Delete(cacheKey)

	// Then
	assert.Nil(t, err)
}

func TestRistrettoInvalidate(t *testing.T) {
	// Given
	options := InvalidateOptions{
		Tags: []string{"tag1"},
	}

	cacheKeys := []byte("a23fdf987h2svc23,jHG2372x38hf74")

	client := &MockRistrettoClientInterface{}
	client.On("Get", "gocache_tag_tag1").Return(cacheKeys, true)
	client.On("Del", "a23fdf987h2svc23").Return(nil)
	client.On("Del", "jHG2372x38hf74").Return(nil)

	store := NewRistretto(client, nil)

	// When
	err := store.Invalidate(options)

	// Then
	assert.Nil(t, err)
}

func TestRistrettoInvalidateWhenError(t *testing.T) {
	// Given
	options := InvalidateOptions{
		Tags: []string{"tag1"},
	}

	cacheKeys := []byte("a23fdf987h2svc23,jHG2372x38hf74")

	client := &MockRistrettoClientInterface{}
	client.On("Get", "gocache_tag_tag1").Return(cacheKeys, false)
	client.On("Del", "a23fdf987h2svc23").Return(nil)
	client.On("Del", "jHG2372x38hf74").Return(nil)

	store := NewRistretto(client, nil)

	// When
	err := store.Invalidate(options)

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
