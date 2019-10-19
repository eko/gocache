package cache

import (
	"errors"
	"testing"

	"github.com/eko/gocache/store"
	mocksCache "github.com/eko/gocache/test/mocks/cache"
	mocksCodec "github.com/eko/gocache/test/mocks/codec"
	mocksStore "github.com/eko/gocache/test/mocks/store"
	"github.com/stretchr/testify/assert"
)

func TestNewLoadable(t *testing.T) {
	// Given
	cache1 := &mocksCache.SetterCacheInterface{}

	loadFunc := func(key interface{}) (interface{}, error) {
		return "test data loaded", nil
	}

	// When
	cache := NewLoadable(loadFunc, cache1)

	// Then
	assert.IsType(t, new(LoadableCache), cache)

	assert.IsType(t, new(loadFunction), &cache.loadFunc)
	assert.Equal(t, cache1, cache.cache)
}

func TestLoadableGetWhenNotAvailableInLoadFunc(t *testing.T) {
	// Given
	// Cache
	store1 := &mocksStore.StoreInterface{}
	store1.On("GetType").Return("store1")

	codec1 := &mocksCodec.CodecInterface{}
	codec1.On("GetStore").Return(store1)

	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("GetCodec").Return(codec1)
	cache1.On("Get", "my-key").Return(nil, errors.New("Unable to find in cache 1"))

	loadFunc := func(key interface{}) (interface{}, error) {
		return nil, errors.New("An error has occurred while loading data from custom source")
	}

	cache := NewLoadable(loadFunc, cache1)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, value)
	assert.Equal(t, errors.New("An error has occurred while loading data from custom source"), err)
}

func TestLoadableGetWhenAvailableInLoadFunc(t *testing.T) {
	// Given
	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	// Cache 1
	store1 := &mocksStore.StoreInterface{}
	store1.On("GetType").Return("store1")

	codec1 := &mocksCodec.CodecInterface{}
	codec1.On("GetStore").Return(store1)

	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("GetCodec").Return(codec1)
	cache1.On("Get", "my-key").Return(nil, errors.New("Unable to find in cache 1"))
	cache1.On("Set", "my-key", cacheValue, (*store.Options)(nil)).Return(nil)

	loadFunc := func(key interface{}) (interface{}, error) {
		return cacheValue, nil
	}

	cache := NewLoadable(loadFunc, cache1)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestLoadableDelete(t *testing.T) {
	// Given
	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Delete", "my-key").Return(nil)

	loadFunc := func(key interface{}) (interface{}, error) {
		return "a value", nil
	}

	cache := NewLoadable(loadFunc, cache1)

	// When
	err := cache.Delete("my-key")

	// Then
	assert.Nil(t, err)
}

func TestLoadableDeleteWhenError(t *testing.T) {
	// Given
	expectedErr := errors.New("Unable to delete key")

	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Delete", "my-key").Return(expectedErr)

	loadFunc := func(key interface{}) (interface{}, error) {
		return "a value", nil
	}

	cache := NewLoadable(loadFunc, cache1)

	// When
	err := cache.Delete("my-key")

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestLoadableInvalidate(t *testing.T) {
	// Given
	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Invalidate", options).Return(nil)

	loadFunc := func(key interface{}) (interface{}, error) {
		return "a value", nil
	}

	cache := NewLoadable(loadFunc, cache1)

	// When
	err := cache.Invalidate(options)

	// Then
	assert.Nil(t, err)
}

func TestLoadableInvalidateWhenError(t *testing.T) {
	// Given
	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	expectedErr := errors.New("Unexpected error when invalidating data")

	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Invalidate", options).Return(expectedErr)

	loadFunc := func(key interface{}) (interface{}, error) {
		return "a value", nil
	}

	cache := NewLoadable(loadFunc, cache1)

	// When
	err := cache.Invalidate(options)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestLoadableGetType(t *testing.T) {
	// Given
	cache1 := &mocksCache.SetterCacheInterface{}

	loadFunc := func(key interface{}) (interface{}, error) {
		return "test data loaded", nil
	}

	cache := NewLoadable(loadFunc, cache1)

	// When - Then
	assert.Equal(t, LoadableType, cache.GetType())
}
