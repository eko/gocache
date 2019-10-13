package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/eko/gache/codec"
	mocksStore "github.com/eko/gache/test/mocks/store"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Given
	store := &mocksStore.StoreInterface{}
	options := &Options{
		Expiration: 5 * time.Second,
	}

	// When
	cache := New(store, options)

	// Then
	assert.IsType(t, new(Cache), cache)
	assert.IsType(t, new(codec.Codec), cache.codec)

	assert.Equal(t, store, cache.codec.GetStore())
	assert.Equal(t, options, cache.options)
}

func TestCacheSet(t *testing.T) {
	// Given
	options := &Options{
		Expiration: 5 * time.Second,
	}

	value := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store := &mocksStore.StoreInterface{}
	store.On("Set", "9b1ac8a6e8ca8ca9477c0a252eb37756", value, options.ExpirationValue()).
		Return(nil)

	cache := New(store, options)

	// When
	err := cache.Set("my-key", value)
	assert.Nil(t, err)
}

func TestCacheSetWhenErrorOccurs(t *testing.T) {
	// Given
	options := &Options{
		Expiration: 5 * time.Second,
	}

	value := &struct {
		Hello string
	}{
		Hello: "world",
	}

	storeErr := errors.New("An error has occured while inserting data into store")

	store := &mocksStore.StoreInterface{}
	store.On("Set", "9b1ac8a6e8ca8ca9477c0a252eb37756", value, options.ExpirationValue()).
		Return(storeErr)

	cache := New(store, options)

	// When
	err := cache.Set("my-key", value)
	assert.Equal(t, storeErr, err)
}

func TestCacheGet(t *testing.T) {
	// Given
	options := &Options{
		Expiration: 5 * time.Second,
	}

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store := &mocksStore.StoreInterface{}
	store.On("Get", "9b1ac8a6e8ca8ca9477c0a252eb37756").Return(cacheValue, nil)

	cache := New(store, options)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestCacheGetWhenNotFound(t *testing.T) {
	// Given
	options := &Options{
		Expiration: 5 * time.Second,
	}

	returnedErr := errors.New("Unable to find item in store")

	store := &mocksStore.StoreInterface{}
	store.On("Get", "9b1ac8a6e8ca8ca9477c0a252eb37756").Return(nil, returnedErr)

	cache := New(store, options)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, value)
	assert.Equal(t, returnedErr, err)
}

func TestCacheGetCodec(t *testing.T) {
	// Given
	store := &mocksStore.StoreInterface{}
	options := &Options{
		Expiration: 5 * time.Second,
	}

	cache := New(store, options)

	// When
	value := cache.GetCodec()

	// Then
	assert.IsType(t, new(codec.Codec), value)
	assert.Equal(t, store, value.GetStore())
}

func TestCacheGetType(t *testing.T) {
	// Given
	store := &mocksStore.StoreInterface{}
	options := &Options{
		Expiration: 5 * time.Second,
	}

	cache := New(store, options)

	// When - Then
	assert.Equal(t, CacheType, cache.GetType())
}
