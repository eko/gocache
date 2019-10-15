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

	storeErr := errors.New("An error has occured while inserting data into store")

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
