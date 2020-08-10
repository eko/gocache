package gocache

import (
	"testing"
	"time"

	mocksStore "github.com/yeqown/gocache/test/mocks/store"
	"github.com/yeqown/gocache/types"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func _testMockCache(store IStore) ICache {
	return &cache{store: store}
}

func TestNew(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// when
	store := mocksStore.NewMockIStore(ctrl)
	mockCache := _testMockCache(store)

	// Then
	assert.IsType(t, new(cache), mockCache)
}

func TestCacheSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	options := &types.StoreOptions{
		Expiration: 5 * time.Second,
	}

	value := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store := mocksStore.NewMockIStore(ctrl)
	store.EXPECT().Set("my-key", value, options).Return(nil)

	cache := _testMockCache(store)

	// When
	err := cache.Set("my-key", value, options)
	assert.Nil(t, err)
}

func TestCacheSetWhenErrorOccurs(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	options := &types.StoreOptions{
		Expiration: 5 * time.Second,
	}

	value := &struct {
		Hello string
	}{
		Hello: "world",
	}

	storeErr := errors.New("An error has occurred while inserting data into store")

	store := mocksStore.NewMockIStore(ctrl)
	store.EXPECT().Set("my-key", value, options).Return(storeErr)

	cache := _testMockCache(store)

	// When
	err := cache.Set("my-key", value, options)
	assert.Equal(t, storeErr, err)
}

func TestCacheGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocksStore.NewMockIStore(ctrl)

	cacheValue := []byte("this is value")
	store.EXPECT().Get("my-key").Return(cacheValue, nil)

	// When
	cache := _testMockCache(store)
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestCacheGetWhenNotFound(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	returnedErr := errors.New("Unable to find item in store")

	store := mocksStore.NewMockIStore(ctrl)
	store.EXPECT().Get("my-key").Return(nil, returnedErr)

	cache := _testMockCache(store)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, value)
	assert.Equal(t, returnedErr, err)
}

//
//func TestCacheGetCodec(t *testing.T) {
//	// Given
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	store := mocksStore.NewMockIStore(ctrl)
//
//	cache := _testMockCache(store)
//
//	// When
//	value := cache.GetCodec()
//
//	// Then
//	assert.IsType(t, new(codec.Codec), value)
//	assert.Equal(t, store, value.GetStore())
//}

func TestCacheGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocksStore.NewMockIStore(ctrl)

	cache := _testMockCache(store)

	// When - Then
	assert.Equal(t, PureCacheType, cache.GetType())
}

func TestCacheDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocksStore.NewMockIStore(ctrl)
	store.EXPECT().Delete("my-key").Return(nil)

	cache := _testMockCache(store)

	// When
	err := cache.Delete("my-key")

	// Then
	assert.Nil(t, err)
}

//
//func TestCacheInvalidate(t *testing.T) {
//	// Given
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	options := types.InvalidateOptions{
//		Tags: []string{"tag1"},
//	}
//
//	store := mocksStore.NewMockIStore(ctrl)
//	store.EXPECT().Invalidate(options).Return(nil)
//
//	cache := _testMockCache(store)
//
//	// When
//	err := cache.Invalidate(options)
//
//	// Then
//	assert.Nil(t, err)
//}
//
//func TestCacheInvalidateWhenError(t *testing.T) {
//	// Given
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	options := types.InvalidateOptions{
//		Tags: []string{"tag1"},
//	}
//
//	expectedErr := errors.New("Unexpected error during invalidation")
//
//	store := mocksStore.NewMockIStore(ctrl)
//	store.EXPECT().Invalidate(options).Return(expectedErr)
//
//	cache := _testMockCache(store)
//
//	// When
//	err := cache.Invalidate(options)
//
//	// Then
//	assert.Equal(t, expectedErr, err)
//}
//
//func TestCacheClear(t *testing.T) {
//	// Given
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	store := mocksStore.NewMockIStore(ctrl)
//	store.EXPECT().Clear().Return(nil)
//
//	cache := _testMockCache(store)
//
//	// When
//	err := cache.Clear()
//
//	// Then
//	assert.Nil(t, err)
//}
//
//func TestCacheClearWhenError(t *testing.T) {
//	// Given
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	expectedErr := errors.New("Unexpected error during invalidation")
//
//	store := mocksStore.NewMockIStore(ctrl)
//	store.EXPECT().Clear().Return(expectedErr)
//
//	cache := _testMockCache(store)
//
//	// When
//	err := cache.Clear()
//
//	// Then
//	assert.Equal(t, expectedErr, err)
//}

func TestCacheDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedErr := errors.New("Unable to delete key")

	store := mocksStore.NewMockIStore(ctrl)
	store.EXPECT().Delete("my-key").Return(expectedErr)

	mockCache := _testMockCache(store)

	// When
	err := mockCache.Delete("my-key")

	// Then
	assert.Equal(t, expectedErr, err)
}
