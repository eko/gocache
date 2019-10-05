package cache

import (
	"errors"
	"testing"

	mocksCache "github.com/eko/gache/test/mocks/cache"
	mocksCodec "github.com/eko/gache/test/mocks/codec"
	mocksStore "github.com/eko/gache/test/mocks/store"
	"github.com/stretchr/testify/assert"
)

func TestNewChain(t *testing.T) {
	// Given
	cache1 := &mocksCache.SetterCacheInterface{}
	cache2 := &mocksCache.SetterCacheInterface{}

	loadFunc := func(key interface{}) (interface{}, error) {
		return "test data loaded", nil
	}

	// When
	cache := NewChain(loadFunc, cache1, cache2)

	// Then
	assert.IsType(t, new(ChainCache), cache)

	assert.IsType(t, new(loadFunction), &cache.loadFunc)
	assert.Equal(t, []SetterCacheInterface{cache1, cache2}, cache.caches)
}

func TestChainGetCaches(t *testing.T) {
	// Given
	cache1 := &mocksCache.SetterCacheInterface{}
	cache2 := &mocksCache.SetterCacheInterface{}

	loadFunc := func(key interface{}) (interface{}, error) {
		return "test data loaded", nil
	}

	cache := NewChain(loadFunc, cache1, cache2)

	// When
	caches := cache.GetCaches()

	// Then
	assert.Equal(t, []SetterCacheInterface{cache1, cache2}, caches)

	assert.Equal(t, cache1, caches[0])
	assert.Equal(t, cache2, caches[1])
}

func TestChainGetWhenNotAvailableInLoadFunc(t *testing.T) {
	// Given
	// Cache 1
	store1 := &mocksStore.StoreInterface{}
	store1.On("GetType").Return("store1")

	codec1 := &mocksCodec.CodecInterface{}
	codec1.On("GetStore").Return(store1)

	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("GetCodec").Return(codec1)
	cache1.On("Get", "my-key").Return(nil, errors.New("Unable to find in cache 1"))

	// Cache 2
	store2 := &mocksStore.StoreInterface{}
	store2.On("GetType").Return("store2")

	codec2 := &mocksCodec.CodecInterface{}
	codec2.On("GetStore").Return(store2)

	cache2 := &mocksCache.SetterCacheInterface{}
	cache2.On("GetCodec").Return(codec2)
	cache2.On("Get", "my-key").Return(nil, errors.New("Unable to find in cache 2"))

	loadFunc := func(key interface{}) (interface{}, error) {
		return nil, errors.New("An error has occured while loading data from custom source")
	}

	cache := NewChain(loadFunc, cache1, cache2)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, value)
	assert.Equal(t, errors.New("An error has occured while loading data from custom source"), err)
}

func TestChainGetWhenAvailableInLoadFunc(t *testing.T) {
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
	cache1.On("Set", "my-key", cacheValue).Return(nil)

	// Cache 2
	store2 := &mocksStore.StoreInterface{}
	store2.On("GetType").Return("store2")

	codec2 := &mocksCodec.CodecInterface{}
	codec2.On("GetStore").Return(store2)

	cache2 := &mocksCache.SetterCacheInterface{}
	cache2.On("GetCodec").Return(codec2)
	cache2.On("Get", "my-key").Return(nil, errors.New("Unable to find in cache 2"))
	cache2.On("Set", "my-key", cacheValue).Return(nil)

	loadFunc := func(key interface{}) (interface{}, error) {
		return cacheValue, nil
	}

	cache := NewChain(loadFunc, cache1, cache2)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestChainGetWhenAvailableInFirstCache(t *testing.T) {
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
	cache1.On("Get", "my-key").Return(cacheValue, nil)
	cache1.AssertNotCalled(t, "Set")

	// Cache 2
	cache2 := &mocksCache.SetterCacheInterface{}
	cache2.AssertNotCalled(t, "Get")

	loadFunc := func(key interface{}) (interface{}, error) {
		return cacheValue, nil
	}

	cache := NewChain(loadFunc, cache1, cache2)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestChainGetWhenAvailableInSecondCache(t *testing.T) {
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
	cache1.On("Set", "my-key", cacheValue).Return(nil)

	// Cache 2
	store2 := &mocksStore.StoreInterface{}
	store2.On("GetType").Return("store2")

	codec2 := &mocksCodec.CodecInterface{}
	codec2.On("GetStore").Return(store2)

	cache2 := &mocksCache.SetterCacheInterface{}
	cache2.On("GetCodec").Return(codec2)
	cache2.On("Get", "my-key").Return(cacheValue, nil)
	cache2.AssertNotCalled(t, "Set")

	loadFunc := func(key interface{}) (interface{}, error) {
		return cacheValue, nil
	}

	cache := NewChain(loadFunc, cache1, cache2)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestChainGetType(t *testing.T) {
	// Given
	cache1 := &mocksCache.SetterCacheInterface{}

	loadFunc := func(key interface{}) (interface{}, error) {
		return "test data loaded", nil
	}

	cache := NewChain(loadFunc, cache1)

	// When - Then
	assert.Equal(t, ChainType, cache.GetType())
}
