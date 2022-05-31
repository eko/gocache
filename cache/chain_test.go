package cache

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/eko/gocache/v2/store"
	mocksCache "github.com/eko/gocache/v2/test/mocks/cache"
	mocksCodec "github.com/eko/gocache/v2/test/mocks/codec"
	mocksStore "github.com/eko/gocache/v2/test/mocks/store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewChain(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache2 := mocksCache.NewMockSetterCacheInterface(ctrl)

	// When
	cache := NewChain(cache1, cache2)

	// Then
	assert.IsType(t, new(ChainCache), cache)

	assert.Equal(t, []SetterCacheInterface{cache1, cache2}, cache.caches)
}

func TestChainGetCaches(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache2 := mocksCache.NewMockSetterCacheInterface(ctrl)

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
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	// Cache 1
	store1 := mocksStore.NewMockStoreInterface(ctrl)
	store1.EXPECT().GetType().AnyTimes().Return("store1")

	codec1 := mocksCodec.NewMockCodecInterface(ctrl)
	codec1.EXPECT().GetStore().AnyTimes().Return(store1)

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().GetCodec().AnyTimes().Return(codec1)
	cache1.EXPECT().GetWithTTL(ctx, "my-key").Return(cacheValue,
		0*time.Second, nil)

	// Cache 2
	cache2 := mocksCache.NewMockSetterCacheInterface(ctrl)

	cache := NewChain(cache1, cache2)

	// When
	value, err := cache.Get(ctx, "my-key")

	// Wait for data to be processed
	for len(cache.setChannel) > 0 {
		time.Sleep(1 * time.Millisecond)
	}

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestChainGetWhenAvailableInSecondCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	options := &store.Options{Expiration: 0 * time.Second}

	// Cache 1
	store1 := mocksStore.NewMockStoreInterface(ctrl)
	store1.EXPECT().GetType().AnyTimes().Return("store1")

	codec1 := mocksCodec.NewMockCodecInterface(ctrl)
	codec1.EXPECT().GetStore().AnyTimes().Return(store1)

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().GetCodec().AnyTimes().Return(codec1)
	cache1.EXPECT().GetWithTTL(ctx, "my-key").Return(nil, 0*time.Second,
		errors.New("Unable to find in cache 1"))
	cache1.EXPECT().Set(ctx, "my-key", cacheValue, options).AnyTimes().Return(nil)

	// Cache 2
	store2 := mocksStore.NewMockStoreInterface(ctrl)
	store2.EXPECT().GetType().AnyTimes().Return("store2")

	codec2 := mocksCodec.NewMockCodecInterface(ctrl)
	codec2.EXPECT().GetStore().AnyTimes().Return(store2)

	cache2 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache2.EXPECT().GetCodec().AnyTimes().Return(codec2)
	cache2.EXPECT().GetWithTTL(ctx, "my-key").Return(cacheValue,
		0*time.Second, nil)

	cache := NewChain(cache1, cache2)

	// When
	value, err := cache.Get(ctx, "my-key")

	// Wait for data to be processed
	for len(cache.setChannel) > 0 {
		time.Sleep(1 * time.Millisecond)
	}

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestChainGetWhenNotAvailableInAnyCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// Cache 1
	store1 := mocksStore.NewMockStoreInterface(ctrl)
	store1.EXPECT().GetType().Return("store1")

	codec1 := mocksCodec.NewMockCodecInterface(ctrl)
	codec1.EXPECT().GetStore().Return(store1)

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().GetCodec().Return(codec1)
	cache1.EXPECT().GetWithTTL(ctx, "my-key").Return(nil, 0*time.Second,
		errors.New("Unable to find in cache 1"))

	// Cache 2
	store2 := mocksStore.NewMockStoreInterface(ctrl)
	store2.EXPECT().GetType().Return("store2")

	codec2 := mocksCodec.NewMockCodecInterface(ctrl)
	codec2.EXPECT().GetStore().Return(store2)

	cache2 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache2.EXPECT().GetCodec().Return(codec2)
	cache2.EXPECT().GetWithTTL(ctx, "my-key").Return(nil, 0*time.Second,
		errors.New("Unable to find in cache 2"))

	cache := NewChain(cache1, cache2)

	// When
	value, err := cache.Get(ctx, "my-key")

	// Then
	assert.Equal(t, errors.New("Unable to find in cache 2"), err)
	assert.Equal(t, nil, value)
}

func TestChainSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	options := &store.Options{}

	// Cache 1
	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Set(ctx, "my-key", cacheValue, options).Return(nil)

	// Cache 2
	cache2 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache2.EXPECT().Set(ctx, "my-key", cacheValue, options).Return(nil)

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Set(ctx, "my-key", cacheValue, options)

	// Then
	assert.Nil(t, err)
}

func TestChainSetWhenErrorOnSetting(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	options := &store.Options{}

	expectedErr := errors.New("An unexpected error occurred while setting data")

	// Cache 1
	store1 := mocksStore.NewMockStoreInterface(ctrl)
	store1.EXPECT().GetType().Return("store1")

	codec1 := mocksCodec.NewMockCodecInterface(ctrl)
	codec1.EXPECT().GetStore().Return(store1)

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().GetCodec().Return(codec1)
	cache1.EXPECT().Set(ctx, "my-key", cacheValue, options).Return(expectedErr)

	// Cache 2
	cache2 := mocksCache.NewMockSetterCacheInterface(ctrl)

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Set(ctx, "my-key", cacheValue, options)

	// Then
	assert.Error(t, err)
	assert.Equal(t, err.Error(), fmt.Sprintf("Unable to set item into cache with store 'store1': %s", expectedErr.Error()))
}

func TestChainDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// Cache 1
	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(nil)

	// Cache 2
	cache2 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache2.EXPECT().Delete(ctx, "my-key").Return(nil)

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)
}

func TestChainDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// Cache 1
	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(errors.New("An error has occurred while deleting key"))

	// Cache 2
	cache2 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache2.EXPECT().Delete(ctx, "my-key").Return(nil)

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)
}

func TestChainInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	// Cache 1
	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Invalidate(ctx, options).Return(nil)

	// Cache 2
	cache2 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache2.EXPECT().Invalidate(ctx, options).Return(nil)

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Invalidate(ctx, options)

	// Then
	assert.Nil(t, err)
}

func TestChainInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	// Cache 1
	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Invalidate(ctx, options).Return(errors.New("An unexpected error has occurred while invalidation data"))

	// Cache 2
	cache2 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache2.EXPECT().Invalidate(ctx, options).Return(nil)

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Invalidate(ctx, options)

	// Then
	assert.Nil(t, err)
}

func TestChainClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// Cache 1
	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Clear(ctx).Return(nil)

	// Cache 2
	cache2 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache2.EXPECT().Clear(ctx).Return(nil)

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestChainClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// Cache 1
	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Clear(ctx).Return(errors.New("An unexpected error has occurred while invalidation data"))

	// Cache 2
	cache2 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache2.EXPECT().Clear(ctx).Return(nil)

	cache := NewChain(cache1, cache2)

	// When
	err := cache.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestChainGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)

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

func TestChainSetWhenErrorInChain(t *testing.T) {

	// Given
	ctrl := gomock.NewController(t)
	store1 := mocksCache.NewMockSetterCacheInterface(ctrl)

	store1.EXPECT().GetType().AnyTimes().Return("store1")
	codec1 := mocksCodec.NewMockCodecInterface(ctrl)
	codec1.EXPECT().GetStore().AnyTimes().Return(store1)
	store1.EXPECT().GetCodec().AnyTimes().Return(codec1)

	ctx := context.Background()
	key := "test-key"
	value := "test-value"
	interError := errors.New("An issue occurred with the cache")
	store1.EXPECT().Set(ctx, key, value, nil).DoAndReturn(func(_, _, _, _ interface{}) error {
		return interError
	})

	store2 := mocksCache.NewMockSetterCacheInterface(ctrl)

	cache := NewChain(store1, store2)

	// assert store2 set is called
	store2.EXPECT().Set(ctx, key, value, nil).Return(nil)

	// When - Then
	err := cache.Set(ctx, key, value, nil)

	expErr := errors.New("error 1 of 1: Unable to set item into cache with store 'store1': An issue occurred with the cache")
	// Then
	assert.Equal(t, expErr, err)
}
