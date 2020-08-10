package wrapper

import (
	"fmt"
	"testing"
	"time"

	"github.com/yeqown/gocache"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	mocksCache "github.com/yeqown/gocache/test/mocks/cache"
	"github.com/yeqown/gocache/types"
)

func TestNewChain(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache1 := mocksCache.NewMockICache(ctrl)
	cache2 := mocksCache.NewMockICache(ctrl)

	// When
	chain := WrapAsChain(cache1, cache2)

	// Then
	assert.IsType(t, new(ChainCache), chain)

	assert.Equal(t, []gocache.ICache{cache1, cache2}, chain.(*ChainCache).caches)
}

//
//func TestChainGetCaches(t *testing.T) {
//	// Given
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	cache1 := mocksCache.NewMockICache(ctrl)
//	cache2 := mocksCache.NewMockICache(ctrl)
//
//	chain := WrapAsChain(cache1, cache2)
//
//	// When
//	caches := chain.GetCaches()
//
//	// Then
//	assert.Equal(t, []cache.ICache{cache1, cache2}, caches)
//
//	assert.Equal(t, cache1, caches[0])
//	assert.Equal(t, cache2, caches[1])
//}

func TestChainGetWhenAvailableInFirstCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheValue := []byte("this is value")

	// Cache 1
	cache1 := mocksCache.NewMockICache(ctrl)
	cache1.EXPECT().GetType().AnyTimes().Return("cache1")
	cache1.EXPECT().Get("my-key").Return(cacheValue, nil)

	// Cache 2
	cache2 := mocksCache.NewMockICache(ctrl)
	// cache2.EXPECT().GetType().Return("cache2")
	chain := WrapAsChain(cache1, cache2)

	// When
	value, err := chain.Get("my-key")

	// Wait for data to be processed
	for len(chain.(*ChainCache).setChannel) > 0 {
		time.Sleep(1 * time.Millisecond)
	}

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestChainGetWhenAvailableInSecondCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheValue := []byte("this is value")

	// Cache 1
	cache1 := mocksCache.NewMockICache(ctrl)
	cache1.EXPECT().GetType().AnyTimes().Return("cache1")
	cache1.EXPECT().Get("my-key").Return(nil, errors.New("Unable to find in cache 1"))
	cache1.EXPECT().Set("my-key", cacheValue, nil).AnyTimes().Return(nil)

	// Cache 2
	cache2 := mocksCache.NewMockICache(ctrl)
	cache2.EXPECT().GetType().AnyTimes().Return("cache2")
	cache2.EXPECT().Get("my-key").Return(cacheValue, nil)
	cache2.EXPECT().Set("my-key", cacheValue, nil).AnyTimes().Return(nil)

	chain := WrapAsChain(cache1, cache2)
	// When
	value, err := chain.Get("my-key")

	// Wait for data to be processed
	for len(chain.(*ChainCache).setChannel) > 0 {
		time.Sleep(10 * time.Millisecond)
	}

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestChainGetWhenNotAvailableInAnyCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Cache 1
	cache1 := mocksCache.NewMockICache(ctrl)
	cache1.EXPECT().GetType().Return("cache1")
	cache1.EXPECT().Get("my-key").Return(nil, errors.New("Unable to find in cache 1"))

	// Cache 2
	exceptedErr := errors.New("Unable to find in cache 2")
	cache2 := mocksCache.NewMockICache(ctrl)
	cache2.EXPECT().GetType().Return("cache2")
	cache2.EXPECT().Get("my-key").Return(nil, exceptedErr)

	chain := WrapAsChain(cache1, cache2)

	// When
	value, err := chain.Get("my-key")

	// Then
	assert.Equal(t, exceptedErr, err)
	t.Logf("value=%v", value)
	assert.Empty(t, value)
	//assert.Equal(t, nil, value)

	for len(chain.(*ChainCache).setChannel) > 0 {
		time.Sleep(time.Millisecond)
	}
}

func TestChainSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	options := &types.StoreOptions{}

	// Cache 1
	cache1 := mocksCache.NewMockICache(ctrl)
	cache1.EXPECT().Set("my-key", cacheValue, options).Return(nil)

	// Cache 2
	cache2 := mocksCache.NewMockICache(ctrl)
	cache2.EXPECT().Set("my-key", cacheValue, options).Return(nil)

	chain := WrapAsChain(cache1, cache2)

	// When
	err := chain.Set("my-key", cacheValue, options)

	// Then
	assert.Nil(t, err)
}

func TestChainSetWhenErrorOnSetting(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheValue := []byte("this is value")
	options := &types.StoreOptions{}
	expectedErr := errors.New("An unexpected error occurred while setting data")

	// Cache 1
	cache1 := mocksCache.NewMockICache(ctrl)
	cache1.EXPECT().GetType().Return("cache1")
	cache1.EXPECT().Set("my-key", cacheValue, options).DoAndReturn(
		func(key string, value interface{}, options *types.StoreOptions) error {
			t.Log("called cache1.Set")
			return expectedErr
		})

	// Cache 2
	cache2 := mocksCache.NewMockICache(ctrl)
	cache2.EXPECT().GetType().Return("cache2")
	cache2.EXPECT().Set("my-key", cacheValue, options).DoAndReturn(
		func(key string, value interface{}, options *types.StoreOptions) error {
			t.Log("called cache2.Set")
			return expectedErr
		})

	// When
	chain := WrapAsChain(cache1, cache2)
	err := chain.Set("my-key", cacheValue, options)

	// Then
	if assert.Error(t, err) {
		assert.Equal(t,
			fmt.Sprintf("Unable to set item into cache with store cache1: %s;"+
				"Unable to set item into cache with store cache2: %s;", expectedErr.Error(), expectedErr.Error()),
			err.Error(),
		)
	}
}

func TestChainDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Cache 1
	cache1 := mocksCache.NewMockICache(ctrl)
	cache1.EXPECT().Delete("my-key").Return(nil)

	// Cache 2
	cache2 := mocksCache.NewMockICache(ctrl)
	cache2.EXPECT().Delete("my-key").Return(nil)

	chain := WrapAsChain(cache1, cache2)

	// When
	err := chain.Delete("my-key")

	// Then
	assert.Nil(t, err)
}

func TestChainDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Cache 1
	expectedErr := errors.New("An error has occurred while deleting key")
	cache1 := mocksCache.NewMockICache(ctrl)
	cache1.EXPECT().Delete("my-key").Return(expectedErr)

	// Cache 2
	cache2 := mocksCache.NewMockICache(ctrl)
	cache2.EXPECT().Delete("my-key").Return(nil)

	cache := WrapAsChain(cache1, cache2)

	// When
	err := cache.Delete("my-key")

	// Then
	if assert.Error(t, err) {
		want := new(types.MultiError)
		want.Add(expectedErr)
		assert.Equal(t, want, err)
	}
}

func TestChainInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	options := types.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	// Cache 1
	cache1 := mocksCache.NewMockICache(ctrl)
	cache1.EXPECT().Invalidate(options).Return(nil)

	// Cache 2
	cache2 := mocksCache.NewMockICache(ctrl)
	cache2.EXPECT().Invalidate(options).Return(nil)

	chain := WrapAsChain(cache1, cache2)

	// When
	err := chain.Invalidate(options)

	// Then
	assert.Nil(t, err)
}

func TestChainInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	options := types.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	// Cache 1
	expectedErr := errors.New("An unexpected error has occurred while invalidation data")
	cache1 := mocksCache.NewMockICache(ctrl)
	cache1.EXPECT().Invalidate(options).Return(expectedErr)

	// Cache 2
	cache2 := mocksCache.NewMockICache(ctrl)
	cache2.EXPECT().Invalidate(options).Return(nil)

	chain := WrapAsChain(cache1, cache2)

	// When
	err := chain.Invalidate(options)

	// Then
	if assert.Error(t, err) {
		want := new(types.MultiError)
		want.Add(expectedErr)
		assert.Equal(t, want, err)
	}
}

//
//func TestChainClear(t *testing.T) {
//	// Given
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	// Cache 1
//	cache1 := mocksCache.NewMockICache(ctrl)
//	cache1.EXPECT().Clear().Return(nil)
//
//	// Cache 2
//	cache2 := mocksCache.NewMockICache(ctrl)
//	cache2.EXPECT().Clear().Return(nil)
//
//	chain := WrapAsChain(cache1, cache2)
//
//	// When
//	err := chain.Clear()
//
//	// Then
//	assert.Nil(t, err)
//}
//
//func TestChainClearWhenError(t *testing.T) {
//	// Given
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	// Cache 1
//	expectedErr := errors.New("An unexpected error has occurred while invalidation data")
//	cache1 := mocksCache.NewMockICache(ctrl)
//	cache1.EXPECT().Clear().Return(expectedErr)
//
//	// Cache 2
//	cache2 := mocksCache.NewMockICache(ctrl)
//	cache2.EXPECT().Clear().Return(nil)
//
//	chain := WrapAsChain(cache1, cache2)
//
//	// When
//	err := chain.Clear()
//
//	// Then
//	if assert.Error(t, err) {
//		want := new(types.MultiError)
//		want.Add(expectedErr)
//		assert.Equal(t, want, err)
//	}
//}

func TestChainGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache1 := mocksCache.NewMockICache(ctrl)

	chain := WrapAsChain(cache1)

	// When - Then
	assert.Equal(t, ChainType, chain.GetType())
}

//
//func TestCacheChecksum(t *testing.T) {
//	testCases := []struct {
//		value        interface{}
//		expectedHash string
//	}{
//		{value: 273273623, expectedHash: "a187c153af38575778244cb3796536da"},
//		{value: "hello-world", expectedHash: "f31215be6928a6f6e0c7c1cf2c68054e"},
//		{value: []byte(`hello-world`), expectedHash: "f097ebac995e666eb074e019cd39d99b"},
//		{value: struct{ Label string }{}, expectedHash: "2938da2beee350d6ea988e404109f428"},
//		{value: struct{ Label string }{Label: "hello-world"}, expectedHash: "4119a1c8530a0420859f1c6ecf2dc0b7"},
//		{value: struct{ Label string }{Label: "hello-everyone"}, expectedHash: "1d7e7ed4acd56d2635f7cb33aa702bdd"},
//	}
//
//	for _, tc := range testCases {
//		value := cache.Md5(tc.value)
//
//		assert.Equal(t, tc.expectedHash, value)
//	}
//}
