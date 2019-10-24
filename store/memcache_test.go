package store

import (
	"errors"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	mocksStore "github.com/eko/gocache/test/mocks/store/clients"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestNewMemcache(t *testing.T) {
	// Given
	client := &mocksStore.MemcacheClientInterface{}
	options := &Options{Expiration: 3 * time.Second}

	// When
	store := NewMemcache(client, options)

	// Then
	assert.IsType(t, new(MemcacheStore), store)
	assert.Equal(t, client, store.client)
	assert.Equal(t, options, store.options)
}

func TestMemcacheGet(t *testing.T) {
	// Given
	options := &Options{Expiration: 3 * time.Second}

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := &mocksStore.MemcacheClientInterface{}
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

func TestMemcacheGetWhenError(t *testing.T) {
	// Given
	options := &Options{Expiration: 3 * time.Second}

	cacheKey := "my-key"

	expectedErr := errors.New("An unexpected error occurred")

	client := &mocksStore.MemcacheClientInterface{}
	client.On("Get", cacheKey).Return(nil, expectedErr)

	store := NewMemcache(client, options)

	// When
	value, err := store.Get(cacheKey)

	// Then
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, value)
}

func TestMemcacheSet(t *testing.T) {
	// Given
	options := &Options{Expiration: 3 * time.Second}

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := &mocksStore.MemcacheClientInterface{}
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

func TestMemcacheSetWhenNoOptionsGiven(t *testing.T) {
	// Given
	options := &Options{Expiration: 3 * time.Second}

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := &mocksStore.MemcacheClientInterface{}
	client.On("Set", &memcache.Item{
		Key:        cacheKey,
		Value:      cacheValue,
		Expiration: int32(3),
	}).Return(nil)

	store := NewMemcache(client, options)

	// When
	err := store.Set(cacheKey, cacheValue, nil)

	// Then
	assert.Nil(t, err)
}

func TestMemcacheSetWhenError(t *testing.T) {
	// Given
	options := &Options{Expiration: 3 * time.Second}

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	expectedErr := errors.New("An unexpected error occurred")

	client := &mocksStore.MemcacheClientInterface{}
	client.On("Set", &memcache.Item{
		Key:        cacheKey,
		Value:      cacheValue,
		Expiration: int32(3),
	}).Return(expectedErr)

	store := NewMemcache(client, options)

	// When
	err := store.Set(cacheKey, cacheValue, nil)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMemcacheSetWithTags(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := &mocksStore.MemcacheClientInterface{}
	client.On("Set", mock.Anything).Return(nil)
	client.On("Get", "gocache_tag_tag1").Return(nil, nil)

	store := NewMemcache(client, nil)

	// When
	err := store.Set(cacheKey, cacheValue, &Options{Tags: []string{"tag1"}})

	// Then
	assert.Nil(t, err)
	client.AssertNumberOfCalls(t, "Set", 2)
}

func TestMemcacheSetWithTagsWhenAlreadyInserted(t *testing.T) {
	// Given
	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := &mocksStore.MemcacheClientInterface{}
	client.On("Set", mock.Anything).Return(nil)
	client.On("Get", "gocache_tag_tag1").Return(&memcache.Item{
		Value: []byte("my-key,a-second-key"),
	}, nil)

	store := NewMemcache(client, nil)

	// When
	err := store.Set(cacheKey, cacheValue, &Options{Tags: []string{"tag1"}})

	// Then
	assert.Nil(t, err)
	client.AssertNumberOfCalls(t, "Set", 2)
}

func TestMemcacheDelete(t *testing.T) {
	// Given
	cacheKey := "my-key"

	client := &mocksStore.MemcacheClientInterface{}
	client.On("Delete", cacheKey).Return(nil)

	store := NewMemcache(client, nil)

	// When
	err := store.Delete(cacheKey)

	// Then
	assert.Nil(t, err)
}

func TestMemcacheDeleteWhenError(t *testing.T) {
	// Given
	expectedErr := errors.New("Unable to delete key")

	cacheKey := "my-key"

	client := &mocksStore.MemcacheClientInterface{}
	client.On("Delete", cacheKey).Return(expectedErr)

	store := NewMemcache(client, nil)

	// When
	err := store.Delete(cacheKey)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMemcacheInvalidate(t *testing.T) {
	// Given
	options := InvalidateOptions{
		Tags: []string{"tag1"},
	}

	cacheKeys := &memcache.Item{
		Value: []byte("a23fdf987h2svc23,jHG2372x38hf74"),
	}

	client := &mocksStore.MemcacheClientInterface{}
	client.On("Get", "gocache_tag_tag1").Return(cacheKeys, nil)
	client.On("Delete", "a23fdf987h2svc23").Return(nil)
	client.On("Delete", "jHG2372x38hf74").Return(nil)

	store := NewMemcache(client, nil)

	// When
	err := store.Invalidate(options)

	// Then
	assert.Nil(t, err)
}

func TestMemcacheInvalidateWhenError(t *testing.T) {
	// Given
	options := InvalidateOptions{
		Tags: []string{"tag1"},
	}

	cacheKeys := &memcache.Item{
		Value: []byte("a23fdf987h2svc23,jHG2372x38hf74"),
	}

	client := &mocksStore.MemcacheClientInterface{}
	client.On("Get", "gocache_tag_tag1").Return(cacheKeys, nil)
	client.On("Delete", "a23fdf987h2svc23").Return(errors.New("Unexpected error"))
	client.On("Delete", "jHG2372x38hf74").Return(nil)

	store := NewMemcache(client, nil)

	// When
	err := store.Invalidate(options)

	// Then
	assert.Nil(t, err)
}

func TestMemcacheClear(t *testing.T) {
	// Given
	client := &mocksStore.MemcacheClientInterface{}
	client.On("FlushAll").Return(nil)

	store := NewMemcache(client, nil)

	// When
	err := store.Clear()

	// Then
	assert.Nil(t, err)
}

func TestMemcacheClearWhenError(t *testing.T) {
	// Given
	expectedErr := errors.New("An unexpected error occurred")

	client := &mocksStore.MemcacheClientInterface{}
	client.On("FlushAll").Return(expectedErr)

	store := NewMemcache(client, nil)

	// When
	err := store.Clear()

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMemcacheGetType(t *testing.T) {
	// Given
	client := &mocksStore.MemcacheClientInterface{}

	store := NewMemcache(client, nil)

	// When - Then
	assert.Equal(t, MemcacheType, store.GetType())
}
