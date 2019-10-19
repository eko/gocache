package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/eko/gocache/codec"
	"github.com/eko/gocache/store"
	mocksStore "github.com/eko/gocache/test/mocks/store"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Given
	store := &mocksStore.StoreInterface{}

	// When
	cache := New(store)

	// Then
	assert.IsType(t, new(Cache), cache)
	assert.IsType(t, new(codec.Codec), cache.codec)

	assert.Equal(t, store, cache.codec.GetStore())
}

func TestCacheSet(t *testing.T) {
	// Given
	options := &store.Options{
		Expiration: 5 * time.Second,
	}

	value := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store := &mocksStore.StoreInterface{}
	store.On("Set", "9b1ac8a6e8ca8ca9477c0a252eb37756", value, options).
		Return(nil)

	cache := New(store)

	// When
	err := cache.Set("my-key", value, options)
	assert.Nil(t, err)
}

func TestCacheSetWhenErrorOccurs(t *testing.T) {
	// Given
	options := &store.Options{
		Expiration: 5 * time.Second,
	}

	value := &struct {
		Hello string
	}{
		Hello: "world",
	}

	storeErr := errors.New("An error has occurred while inserting data into store")

	store := &mocksStore.StoreInterface{}
	store.On("Set", "9b1ac8a6e8ca8ca9477c0a252eb37756", value, options).
		Return(storeErr)

	cache := New(store)

	// When
	err := cache.Set("my-key", value, options)
	assert.Equal(t, storeErr, err)
}

func TestCacheGet(t *testing.T) {
	// Given
	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store := &mocksStore.StoreInterface{}
	store.On("Get", "9b1ac8a6e8ca8ca9477c0a252eb37756").Return(cacheValue, nil)

	cache := New(store)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestCacheGetWhenNotFound(t *testing.T) {
	// Given
	returnedErr := errors.New("Unable to find item in store")

	store := &mocksStore.StoreInterface{}
	store.On("Get", "9b1ac8a6e8ca8ca9477c0a252eb37756").Return(nil, returnedErr)

	cache := New(store)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, value)
	assert.Equal(t, returnedErr, err)
}

func TestCacheGetCodec(t *testing.T) {
	// Given
	store := &mocksStore.StoreInterface{}

	cache := New(store)

	// When
	value := cache.GetCodec()

	// Then
	assert.IsType(t, new(codec.Codec), value)
	assert.Equal(t, store, value.GetStore())
}

func TestCacheGetType(t *testing.T) {
	// Given
	store := &mocksStore.StoreInterface{}

	cache := New(store)

	// When - Then
	assert.Equal(t, CacheType, cache.GetType())
}

func TestCacheDelete(t *testing.T) {
	// Given
	store := &mocksStore.StoreInterface{}
	store.On("Delete", "9b1ac8a6e8ca8ca9477c0a252eb37756").Return(nil)

	cache := New(store)

	// When
	err := cache.Delete("my-key")

	// Then
	assert.Nil(t, err)
}

func TestCacheInvalidate(t *testing.T) {
	// Given
	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	store := &mocksStore.StoreInterface{}
	store.On("Invalidate", options).Return(nil)

	cache := New(store)

	// When
	err := cache.Invalidate(options)

	// Then
	assert.Nil(t, err)
}

func TestCacheInvalidateWhenError(t *testing.T) {
	// Given
	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	expectedErr := errors.New("Unexpected error during invalidation")

	store := &mocksStore.StoreInterface{}
	store.On("Invalidate", options).Return(expectedErr)

	cache := New(store)

	// When
	err := cache.Invalidate(options)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestCacheDeleteWhenError(t *testing.T) {
	// Given
	expectedErr := errors.New("Unable to delete key")

	store := &mocksStore.StoreInterface{}
	store.On("Delete", "9b1ac8a6e8ca8ca9477c0a252eb37756").Return(expectedErr)

	cache := New(store)

	// When
	err := cache.Delete("my-key")

	// Then
	assert.Equal(t, expectedErr, err)
}
