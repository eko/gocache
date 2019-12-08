package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/eko/gocache/store"
	mocksCache "github.com/eko/gocache/test/mocks/cache"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewLoadable(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)

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

func TestLoadableGetWhenAlreadyInCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Get("my-key").Return(cacheValue, nil)

	loadFunc := func(key interface{}) (interface{}, error) {
		return nil, errors.New("Should not be called")
	}

	cache := NewLoadable(loadFunc, cache1)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestLoadableGetWhenNotAvailableInLoadFunc(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Cache
	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Get("my-key").Return(nil, errors.New("Unable to find in cache 1"))

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	// Cache 1
	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Get("my-key").Return(nil, errors.New("Unable to find in cache 1"))
	cache1.EXPECT().Set("my-key", cacheValue, (*store.Options)(nil)).AnyTimes().Return(nil)

	loadFunc := func(key interface{}) (interface{}, error) {
		return cacheValue, nil
	}

	cache := NewLoadable(loadFunc, cache1)

	// When
	value, err := cache.Get("my-key")

	// Wait for data to be processed
	for len(cache.setChannel) > 0 {
		time.Sleep(1 * time.Millisecond)
	}

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestLoadableDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Delete("my-key").Return(nil)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedErr := errors.New("Unable to delete key")

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Delete("my-key").Return(expectedErr)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Invalidate(options).Return(nil)

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	expectedErr := errors.New("Unexpected error when invalidating data")

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Invalidate(options).Return(expectedErr)

	loadFunc := func(key interface{}) (interface{}, error) {
		return "a value", nil
	}

	cache := NewLoadable(loadFunc, cache1)

	// When
	err := cache.Invalidate(options)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestLoadableClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Clear().Return(nil)

	loadFunc := func(key interface{}) (interface{}, error) {
		return "a value", nil
	}

	cache := NewLoadable(loadFunc, cache1)

	// When
	err := cache.Clear()

	// Then
	assert.Nil(t, err)
}

func TestLoadableClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedErr := errors.New("Unexpected error when invalidating data")

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Clear().Return(expectedErr)

	loadFunc := func(key interface{}) (interface{}, error) {
		return "a value", nil
	}

	cache := NewLoadable(loadFunc, cache1)

	// When
	err := cache.Clear()

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestLoadableGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)

	loadFunc := func(key interface{}) (interface{}, error) {
		return "test data loaded", nil
	}

	cache := NewLoadable(loadFunc, cache1)

	// When - Then
	assert.Equal(t, LoadableType, cache.GetType())
}
