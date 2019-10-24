package cache

import (
	"errors"
	"fmt"
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

func TestChainGetWhenNotAvailableInAnyCache(t *testing.T) {
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
	cache2.AssertNotCalled(t, "Set")

	cache := NewChain(cache1, cache2)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Equal(t, errors.New("Unable to find in cache 2"), err)
	assert.Equal(t, nil, value)
}

func TestChainSet(t *testing.T) {
	// Given
	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	options := &store.Options{}

	// Cache 1
	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Set", "my-key", cacheValue, options).Return(nil)

	// Cache 2
	cache2 := &mocksCache.SetterCacheInterface{}
	cache2.On("Set", "my-key", cacheValue, options).Return(nil)

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Set("my-key", cacheValue, options)

	// Then
	assert.Nil(t, err)
}

func TestChainSetWhenErrorOnSetting(t *testing.T) {
	// Given
	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	options := &store.Options{}

	expectedErr := errors.New("An unexpected error occurred while setting data")

	// Cache 1
	store1 := &mocksStore.StoreInterface{}
	store1.On("GetType").Return("store1")

	codec1 := &mocksCodec.CodecInterface{}
	codec1.On("GetStore").Return(store1)

	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("GetCodec").Return(codec1)
	cache1.On("Set", "my-key", cacheValue, options).Return(expectedErr)

	// Cache 2
	cache2 := &mocksCache.SetterCacheInterface{}
	cache2.AssertNotCalled(t, "Set")

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Set("my-key", cacheValue, options)

	// Then
	assert.Error(t, err)
	assert.Equal(t, err.Error(), fmt.Sprintf("Unable to set item into cache with store 'store1': %s", expectedErr.Error()))
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
	cache1.On("Delete", "my-key").Return(errors.New("An error has occurred while deleting key"))

	// Cache 2
	cache2 := &mocksCache.SetterCacheInterface{}
	cache2.On("Delete", "my-key").Return(nil)

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Delete("my-key")

	// Then
	assert.Nil(t, err)
}

func TestChainInvalidate(t *testing.T) {
	// Given
	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	// Cache 1
	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Invalidate", options).Return(nil)

	// Cache 2
	cache2 := &mocksCache.SetterCacheInterface{}
	cache2.On("Invalidate", options).Return(nil)

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Invalidate(options)

	// Then
	assert.Nil(t, err)
}

func TestChainInvalidateWhenError(t *testing.T) {
	// Given
	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	// Cache 1
	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Invalidate", options).Return(errors.New("An unexpected error has occurred while invalidation data"))

	// Cache 2
	cache2 := &mocksCache.SetterCacheInterface{}
	cache2.On("Invalidate", options).Return(nil)

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Invalidate(options)

	// Then
	assert.Nil(t, err)
}

func TestChainClear(t *testing.T) {
	// Given
	// Cache 1
	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Clear").Return(nil)

	// Cache 2
	cache2 := &mocksCache.SetterCacheInterface{}
	cache2.On("Clear").Return(nil)

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Clear()

	// Then
	assert.Nil(t, err)
}

func TestChainClearWhenError(t *testing.T) {
	// Given
	// Cache 1
	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Clear").Return(errors.New("An unexpected error has occurred while invalidation data"))

	// Cache 2
	cache2 := &mocksCache.SetterCacheInterface{}
	cache2.On("Clear").Return(nil)

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Clear()

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

func TestCacheChecksum(t *testing.T) {
	testCases := []struct {
		value        interface{}
		expectedHash string
	}{
		{value: 273273623, expectedHash: "a187c153af38575778244cb3796536da"},
		{value: "hello-world", expectedHash: "f31215be6928a6f6e0c7c1cf2c68054e"},
		{value: []byte(`hello-world`), expectedHash: "f097ebac995e666eb074e019cd39d99b"},
		{value: struct{ Label string }{}, expectedHash: "2938da2beee350d6ea988e404109f428"},
		{value: struct{ Label string }{Label: "hello-world"}, expectedHash: "4119a1c8530a0420859f1c6ecf2dc0b7"},
		{value: struct{ Label string }{Label: "hello-everyone"}, expectedHash: "1d7e7ed4acd56d2635f7cb33aa702bdd"},
	}

	for _, tc := range testCases {
		value := checksum(tc.value)

		assert.Equal(t, tc.expectedHash, value)
	}
}
