package store

import (
	"errors"
	"testing"

	mocksStore "github.com/eko/gocache/test/mocks/store/clients"
	"github.com/stretchr/testify/assert"
)

func TestNewBigcache(t *testing.T) {
	// Given
	client := &mocksStore.BigcacheClientInterface{}

	// When
	store := NewBigcache(client, nil)

	// Then
	assert.IsType(t, new(BigcacheStore), store)
	assert.Equal(t, client, store.client)
	assert.IsType(t, new(Options), store.options)
}

func TestBigcacheGet(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := &mocksStore.BigcacheClientInterface{}
	client.On("Get", cacheKey).Return(cacheValue, nil)

	store := NewBigcache(client, nil)

	// When
	value, err := store.Get(cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestBigcacheGetWhenError(t *testing.T) {
	// Given
	cacheKey := "my-key"

	expectedErr := errors.New("An unexpected error occurred")

	client := &mocksStore.BigcacheClientInterface{}
	client.On("Get", cacheKey).Return(nil, expectedErr)

	store := NewBigcache(client, nil)

	// When
	value, err := store.Get(cacheKey)

	// Then
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, value)
}

func TestBigcacheSet(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	options := &Options{}

	client := &mocksStore.BigcacheClientInterface{}
	client.On("Set", cacheKey, cacheValue).Return(nil)

	store := NewBigcache(client, nil)

	// When
	err := store.Set(cacheKey, cacheValue, options)

	// Then
	assert.Nil(t, err)
}

func TestBigcacheSetWhenError(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	options := &Options{}

	expectedErr := errors.New("An unexpected error occurred")

	client := &mocksStore.BigcacheClientInterface{}
	client.On("Set", cacheKey, cacheValue).Return(expectedErr)

	store := NewBigcache(client, options)

	// When
	err := store.Set(cacheKey, cacheValue, nil)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestBigcacheSetWithTags(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := &mocksStore.BigcacheClientInterface{}
	client.On("Set", cacheKey, cacheValue).Return(nil)
	client.On("Get", "gocache_tag_tag1").Return(nil, nil)
	client.On("Set", "gocache_tag_tag1", []byte("my-key")).Return(nil)

	store := NewBigcache(client, nil)

	// When
	err := store.Set(cacheKey, cacheValue, &Options{Tags: []string{"tag1"}})

	// Then
	assert.Nil(t, err)
}

func TestBigcacheSetWithTagsWhenAlreadyInserted(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := &mocksStore.BigcacheClientInterface{}
	client.On("Set", cacheKey, cacheValue).Return(nil)
	client.On("Get", "gocache_tag_tag1").Return([]byte("my-key,a-second-key"), nil)
	client.On("Set", "gocache_tag_tag1", []byte("my-key,a-second-key")).Return(nil)

	store := NewBigcache(client, nil)

	// When
	err := store.Set(cacheKey, cacheValue, &Options{Tags: []string{"tag1"}})

	// Then
	assert.Nil(t, err)
}

func TestBigcacheDelete(t *testing.T) {
	// Given
	cacheKey := "my-key"

	client := &mocksStore.BigcacheClientInterface{}
	client.On("Delete", cacheKey).Return(nil)

	store := NewBigcache(client, nil)

	// When
	err := store.Delete(cacheKey)

	// Then
	assert.Nil(t, err)
}

func TestBigcacheDeleteWhenError(t *testing.T) {
	// Given
	expectedErr := errors.New("Unable to delete key")

	cacheKey := "my-key"

	client := &mocksStore.BigcacheClientInterface{}
	client.On("Delete", cacheKey).Return(expectedErr)

	store := NewBigcache(client, nil)

	// When
	err := store.Delete(cacheKey)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestBigcacheInvalidate(t *testing.T) {
	// Given
	options := InvalidateOptions{
		Tags: []string{"tag1"},
	}

	cacheKeys := []byte("a23fdf987h2svc23,jHG2372x38hf74")

	client := &mocksStore.BigcacheClientInterface{}
	client.On("Get", "gocache_tag_tag1").Return(cacheKeys, nil)
	client.On("Delete", "a23fdf987h2svc23").Return(nil)
	client.On("Delete", "jHG2372x38hf74").Return(nil)

	store := NewBigcache(client, nil)

	// When
	err := store.Invalidate(options)

	// Then
	assert.Nil(t, err)
}

func TestBigcacheInvalidateWhenError(t *testing.T) {
	// Given
	options := InvalidateOptions{
		Tags: []string{"tag1"},
	}

	cacheKeys := []byte("a23fdf987h2svc23,jHG2372x38hf74")

	client := &mocksStore.BigcacheClientInterface{}
	client.On("Get", "gocache_tag_tag1").Return(cacheKeys, nil)
	client.On("Delete", "a23fdf987h2svc23").Return(errors.New("Unexpected error"))
	client.On("Delete", "jHG2372x38hf74").Return(nil)

	store := NewBigcache(client, nil)

	// When
	err := store.Invalidate(options)

	// Then
	assert.Nil(t, err)
}

func TestBigcacheClear(t *testing.T) {
	// Given
	client := &mocksStore.BigcacheClientInterface{}
	client.On("Reset").Return(nil)

	store := NewBigcache(client, nil)

	// When
	err := store.Clear()

	// Then
	assert.Nil(t, err)
}

func TestBigcacheClearWhenError(t *testing.T) {
	// Given
	expectedErr := errors.New("An unexpected error occurred")

	client := &mocksStore.BigcacheClientInterface{}
	client.On("Reset").Return(expectedErr)

	store := NewBigcache(client, nil)

	// When
	err := store.Clear()

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestBigcacheGetType(t *testing.T) {
	// Given
	client := &mocksStore.BigcacheClientInterface{}

	store := NewBigcache(client, nil)

	// When - Then
	assert.Equal(t, BigcacheType, store.GetType())
}
