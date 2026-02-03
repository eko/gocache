package serializer

import (
	"context"
	"errors"
	"testing"
	"time"

	mockcache "github.com/eko/gocache/lib/v4/internal/mocks/cache"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/mock/gomock"
)

type testCacheValue struct {
	Hello string
}

func TestNew(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache := mockcache.NewMockCacheInterface[[]byte](ctrl)

	// When
	serializer := New[testCacheValue](DefaultSerializer{}, cache)

	// Then
	assert.IsType(t, new(SerializerCache[testCacheValue]), serializer)
	assert.Equal(t, cache, serializer.cache)
}

func TestGetWhenStoreReturnsSliceOfBytes(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &testCacheValue{
		Hello: "world",
	}

	cacheValueBytes, err := msgpack.Marshal(cacheValue)
	if err != nil {
		assert.Error(t, err)
	}

	cache := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache.EXPECT().Get(ctx, "my-key").Return(cacheValueBytes, nil)

	serializer := New[*testCacheValue](DefaultSerializer{}, cache)

	// When
	value, err := serializer.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestGetWhenUnmarshalingError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache.EXPECT().Get(ctx, "my-key").Return([]byte("unknown-string"), nil)

	serializer := New[*testCacheValue](DefaultSerializer{}, cache)

	// When
	value, err := serializer.Get(ctx, "my-key")

	// Then
	assert.NotNil(t, err)
	assert.Nil(t, value)
}

func TestGetWhenNotFoundInStore(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unable to find item in store")

	cache := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache.EXPECT().Get(ctx, "my-key").Return(nil, expectedErr)

	serializer := New[*testCacheValue](DefaultSerializer{}, cache)

	// When
	value, err := serializer.Get(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, value)
}

func TestSetWhenStruct(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &testCacheValue{
		Hello: "world",
	}

	cache := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache.EXPECT().Set(
		ctx,
		"my-key",
		[]byte{0x81, 0xa5, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0xa5, 0x77, 0x6f, 0x72, 0x6c, 0x64},
		store.OptionsMatcher{
			Expiration: 5 * time.Second,
		},
	).Return(nil)

	serializer := New[*testCacheValue](DefaultSerializer{}, cache)

	// When
	err := serializer.Set(ctx, "my-key", cacheValue, store.WithExpiration(5*time.Second))

	// Then
	assert.Nil(t, err)
}

func TestSetWhenString(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := "test"

	cache := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache.EXPECT().Set(
		ctx,
		"my-key",
		[]byte{0xa4, 0x74, 0x65, 0x73, 0x74},
		store.OptionsMatcher{
			Expiration: 5 * time.Second,
		},
	).Return(nil)

	serializer := New[string](DefaultSerializer{}, cache)

	// When
	err := serializer.Set(ctx, "my-key", cacheValue, store.WithExpiration(5*time.Second))

	// Then
	assert.Nil(t, err)
}

func TestSetWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := "test"

	expectedErr := errors.New("an unexpected error occurred")

	cache := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache.EXPECT().Set(
		ctx,
		"my-key",
		[]byte{0xa4, 0x74, 0x65, 0x73, 0x74},
		store.OptionsMatcher{Expiration: 5 * time.Second},
	).Return(expectedErr)

	serializer := New[string](DefaultSerializer{}, cache)

	// When
	err := serializer.Set(ctx, "my-key", cacheValue, store.WithExpiration(5*time.Second))

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache.EXPECT().Delete(ctx, "my-key").Return(nil)

	serializer := New[string](DefaultSerializer{}, cache)

	// When
	err := serializer.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)
}

func TestDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unable to delete key")

	cache := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache.EXPECT().Delete(ctx, "my-key").Return(expectedErr)

	serializer := New[string](DefaultSerializer{}, cache)

	// When
	err := serializer.Delete(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache.EXPECT().Invalidate(ctx, store.InvalidateOptionsMatcher{
		Tags: []string{"tag1"},
	}).Return(nil)

	serializer := New[string](DefaultSerializer{}, cache)

	// When
	err := serializer.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestInvalidatingWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unexpected error when invalidating data")

	cache := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache.EXPECT().Invalidate(ctx, store.InvalidateOptionsMatcher{Tags: []string{"tag1"}}).Return(expectedErr)

	serializer := New[string](DefaultSerializer{}, cache)

	// When
	err := serializer.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache.EXPECT().Clear(ctx).Return(nil)

	serializer := New[string](DefaultSerializer{}, cache)

	// When
	err := serializer.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("an unexpected error occurred")

	cache := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache.EXPECT().Clear(ctx).Return(expectedErr)

	serializer := New[string](DefaultSerializer{}, cache)

	// When
	err := serializer.Clear(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}
