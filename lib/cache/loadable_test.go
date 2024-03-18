package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/eko/gocache/lib/v4/store"
	freecache2 "github.com/eko/gocache/store/freecache/v4"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewLoadable(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := NewMockSetterCacheInterface[any](ctrl)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return "test data loaded", nil
	}

	// When
	cache := NewLoadable[any](loadFunc, cache1)

	// Then
	assert.IsType(t, new(LoadableCache[any]), cache)

	assert.IsType(t, new(LoadFunction[any]), &cache.loadFunc)
	assert.Equal(t, cache1, cache.cache)
}

func TestLoadableGetWhenAlreadyInCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return(cacheValue, nil)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return nil, errors.New("should not be called")
	}

	cache := NewLoadable[any](loadFunc, cache1)

	// When
	value, err := cache.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestLoadableGetWhenNotAvailableInLoadFunc(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// Cache
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return(nil, errors.New("unable to find in cache 1"))

	loadFunc := func(_ context.Context, key any) (any, error) {
		return nil, errors.New("an error has occurred while loading data from custom source")
	}

	cache := NewLoadable[any](loadFunc, cache1)

	// When
	value, err := cache.Get(ctx, "my-key")

	// Then
	assert.Nil(t, value)
	assert.Equal(t, errors.New("an error has occurred while loading data from custom source"), err)
}

func TestLoadableGetWhenAvailableInLoadFunc(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	// Cache 1
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return(nil, errors.New("unable to find in cache 1"))
	cache1.EXPECT().Set(ctx, "my-key", cacheValue).AnyTimes().Return(nil)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return cacheValue, nil
	}

	cache := NewLoadable[any](loadFunc, cache1)

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

func TestLoadableDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(nil)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return "a value", nil
	}

	cache := NewLoadable[any](loadFunc, cache1)

	// When
	err := cache.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)
}

func TestLoadableDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unable to delete key")

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(expectedErr)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return "a value", nil
	}

	cache := NewLoadable[any](loadFunc, cache1)

	// When
	err := cache.Delete(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestLoadableInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Invalidate(ctx).Return(nil)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return "a value", nil
	}

	cache := NewLoadable[any](loadFunc, cache1)

	// When
	err := cache.Invalidate(ctx)

	// Then
	assert.Nil(t, err)
}

func TestLoadableInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unexpected error when invalidating data")

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Invalidate(ctx).Return(expectedErr)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return "a value", nil
	}

	cache := NewLoadable[any](loadFunc, cache1)

	// When
	err := cache.Invalidate(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestLoadableClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Clear(ctx).Return(nil)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return "a value", nil
	}

	cache := NewLoadable[any](loadFunc, cache1)

	// When
	err := cache.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestLoadableClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unexpected error when invalidating data")

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Clear(ctx).Return(expectedErr)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return "a value", nil
	}

	cache := NewLoadable[any](loadFunc, cache1)

	// When
	err := cache.Clear(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestLoadableGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := NewMockSetterCacheInterface[any](ctrl)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return "test data loaded", nil
	}

	cache := NewLoadable[any](loadFunc, cache1)

	// When - Then
	assert.Equal(t, LoadableType, cache.GetType())
}

func TestLoadableSynchronousSetOptions(t *testing.T) {

	count := 0
	cacheValue := []byte("foobar")

	ctx := context.Background()

	freeCacheStore := freecache2.NewFreecache(freecache.NewCache(1000), store.WithExpiration(10*time.Second))
	loadFunc := func(_ context.Context, key any) (any, error) {
		count++
		return cacheValue, nil
	}

	cache := NewLoadable[any](loadFunc, freeCacheStore, store.WithSynchronousSet())

	for i := 0; i < 10; i++ {
		value, err := cache.Get(ctx, "my-key")
		assert.NoError(t, err)
		assert.Equal(t, value, cacheValue)
	}

	assert.Equal(t, 1, count)
}
