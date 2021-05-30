package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eko/gocache/v2/codec"
	"github.com/eko/gocache/v2/store"
	mocksStore "github.com/eko/gocache/v2/test/mocks/store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	store := mocksStore.NewMockStoreInterface(ctrl)

	// When
	cache := New(store)

	// Then
	assert.IsType(t, new(Cache), cache)
	assert.IsType(t, new(codec.Codec), cache.codec)

	assert.Equal(t, store, cache.codec.GetStore())
}

func TestCacheSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	options := &store.Options{
		Expiration: 5 * time.Second,
	}

	value := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Set(ctx, "my-key", value, options).Return(nil)

	cache := New(store)

	// When
	err := cache.Set(ctx, "my-key", value, options)
	assert.Nil(t, err)
}

func TestCacheSetWhenErrorOccurs(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	options := &store.Options{
		Expiration: 5 * time.Second,
	}

	value := &struct {
		Hello string
	}{
		Hello: "world",
	}

	storeErr := errors.New("An error has occurred while inserting data into store")

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Set(ctx, "my-key", value, options).Return(storeErr)

	cache := New(store)

	// When
	err := cache.Set(ctx, "my-key", value, options)
	assert.Equal(t, storeErr, err)
}

func TestCacheGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Get(ctx, "my-key").Return(cacheValue, nil)

	cache := New(store)

	// When
	value, err := cache.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestCacheGetWhenNotFound(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	returnedErr := errors.New("Unable to find item in store")

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Get(ctx, "my-key").Return(nil, returnedErr)

	cache := New(store)

	// When
	value, err := cache.Get(ctx, "my-key")

	// Then
	assert.Nil(t, value)
	assert.Equal(t, returnedErr, err)
}

func TestCacheGetWithTTL(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}
	expiration := 1 * time.Second

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().GetWithTTL(ctx, "my-key").
		Return(cacheValue, expiration, nil)

	cache := New(store)

	// When
	value, ttl, err := cache.GetWithTTL(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
	assert.Equal(t, expiration, ttl)
}

func TestCacheGetWithTTLWhenNotFound(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	returnedErr := errors.New("Unable to find item in store")
	expiration := 0 * time.Second

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().GetWithTTL(ctx, "my-key").
		Return(nil, expiration, returnedErr)

	cache := New(store)

	// When
	value, ttl, err := cache.GetWithTTL(ctx, "my-key")

	// Then
	assert.Nil(t, value)
	assert.Equal(t, returnedErr, err)
	assert.Equal(t, expiration, ttl)
}

func TestCacheGetCodec(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	store := mocksStore.NewMockStoreInterface(ctrl)

	cache := New(store)

	// When
	value := cache.GetCodec()

	// Then
	assert.IsType(t, new(codec.Codec), value)
	assert.Equal(t, store, value.GetStore())
}

func TestCacheGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	store := mocksStore.NewMockStoreInterface(ctrl)

	cache := New(store)

	// When - Then
	assert.Equal(t, CacheType, cache.GetType())
}

func TestCacheGetCacheKeyWhenKeyIsString(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	store := mocksStore.NewMockStoreInterface(ctrl)

	cache := New(store)

	// When
	computedKey := cache.getCacheKey("my-Key")

	// Then
	assert.Equal(t, "my-Key", computedKey)
}

func TestCacheGetCacheKeyWhenKeyIsStruct(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	store := mocksStore.NewMockStoreInterface(ctrl)

	cache := New(store)

	// When
	key := &struct {
		Hello string
	}{
		Hello: "world",
	}

	computedKey := cache.getCacheKey(key)

	// Then
	assert.Equal(t, "8144fe5310cf0e62ac83fd79c113aad2", computedKey)
}

func TestCacheDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Delete(ctx, "my-key").Return(nil)

	cache := New(store)

	// When
	err := cache.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)
}

func TestCacheInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Invalidate(ctx, options).Return(nil)

	cache := New(store)

	// When
	err := cache.Invalidate(ctx, options)

	// Then
	assert.Nil(t, err)
}

func TestCacheInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	expectedErr := errors.New("Unexpected error during invalidation")

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Invalidate(ctx, options).Return(expectedErr)

	cache := New(store)

	// When
	err := cache.Invalidate(ctx, options)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestCacheClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Clear(ctx).Return(nil)

	cache := New(store)

	// When
	err := cache.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestCacheClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("Unexpected error during invalidation")

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Clear(ctx).Return(expectedErr)

	cache := New(store)

	// When
	err := cache.Clear(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestCacheDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("Unable to delete key")

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Delete(ctx, "my-key").Return(expectedErr)

	cache := New(store)

	// When
	err := cache.Delete(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)
}
