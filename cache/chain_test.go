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

func TestNewChain(t *testing.T) {
	// Given
	cache1 := &mocksCache.SetterCacheInterface{}
	cache2 := &mocksCache.SetterCacheInterface{}

	// When
	cache := NewChain(cache1, cache2)

	// Then
	assert.IsType(t, new(ChainCache), cache)

	assert.Equal(t, []SetterCacheInterface{cache1, cache2}, cache.caches)
}

func TestChainGetCaches(t *testing.T) {
	// Given
	cache1 := &mocksCache.SetterCacheInterface{}
	cache2 := &mocksCache.SetterCacheInterface{}

	cache := NewChain(cache1, cache2)

	// When
	caches := cache.GetCaches()

	// Then
	assert.Equal(t, []SetterCacheInterface{cache1, cache2}, caches)

	assert.Equal(t, cache1, caches[0])
	assert.Equal(t, cache2, caches[1])
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

	cache := NewChain(cache1, cache2)

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
	cache1.On("Set", "my-key", cacheValue, (*store.Options)(nil)).Return(nil)

	// Cache 2
	store2 := &mocksStore.StoreInterface{}
	store2.On("GetType").Return("store2")

	codec2 := &mocksCodec.CodecInterface{}
	codec2.On("GetStore").Return(store2)

	cache2 := &mocksCache.SetterCacheInterface{}
	cache2.On("GetCodec").Return(codec2)
	cache2.On("Get", "my-key").Return(cacheValue, nil)
	cache2.AssertNotCalled(t, "Set")

	cache := NewChain(cache1, cache2)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestChainDelete(t *testing.T) {
	// Given
	// Cache 1
	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Delete", "my-key").Return(nil)

	// Cache 2
	cache2 := &mocksCache.SetterCacheInterface{}
	cache2.On("Delete", "my-key").Return(nil)

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Delete("my-key")

	// Then
	assert.Nil(t, err)
}

func TestChainDeleteWhenError(t *testing.T) {
	// Given
	// Cache 1
	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Delete", "my-key").Return(errors.New("An error has occured while deleting key"))

	// Cache 2
	cache2 := &mocksCache.SetterCacheInterface{}
	cache2.On("Delete", "my-key").Return(nil)

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Delete("my-key")

	// Then
	assert.Nil(t, err)
}

func TestChainGetType(t *testing.T) {
	// Given
	cache1 := &mocksCache.SetterCacheInterface{}

	cache := NewChain(cache1)

	// When - Then
	assert.Equal(t, ChainType, cache.GetType())
}
