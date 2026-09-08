package marshaler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	mockcache "github.com/eko/gocache/lib/v4/mocks/cache"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/mock/gomock"
)

// Cache has to stay usable everywhere a CacheInterface is expected
var _ cache.CacheInterface[*testCacheValue] = new(Cache[*testCacheValue])

func TestNewCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := mockcache.NewMockCacheInterface[[]byte](ctrl)

	// When
	marshaler := NewCache[*testCacheValue](cache1)

	// Then
	assert.IsType(t, new(Cache[*testCacheValue]), marshaler)
	assert.Equal(t, cache1, marshaler.cache)
	assert.Equal(t, MarshalerType, marshaler.GetType())
}

func TestCacheGetWhenPointerType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &testCacheValue{Hello: "world"}
	cacheValueBytes, err := msgpack.Marshal(cacheValue)
	assert.Nil(t, err)

	cache1 := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return(cacheValueBytes, nil)

	marshaler := NewCache[*testCacheValue](cache1)

	// When
	value, err := marshaler.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestCacheGetWhenStructType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := testCacheValue{Hello: "world"}
	cacheValueBytes, err := msgpack.Marshal(cacheValue)
	assert.Nil(t, err)

	cache1 := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return(cacheValueBytes, nil)

	marshaler := NewCache[testCacheValue](cache1)

	// When
	value, err := marshaler.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestCacheGetWhenNotFound(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := store.NotFoundWithCause(errors.New("value not found in store"))

	cache1 := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return(nil, expectedErr)

	marshaler := NewCache[*testCacheValue](cache1)

	// When
	value, err := marshaler.Get(ctx, "my-key")

	// Then
	assert.Nil(t, value)
	assert.ErrorIs(t, err, store.NotFound{})
}

func TestCacheGetWhenUnmarshalingFails(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return([]byte("not-msgpack"), nil)

	marshaler := NewCache[*testCacheValue](cache1)

	// When
	value, err := marshaler.Get(ctx, "my-key")

	// Then
	assert.Nil(t, value)
	assert.NotNil(t, err)
}

func TestCacheGetWithTTL(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &testCacheValue{Hello: "world"}
	cacheValueBytes, err := msgpack.Marshal(cacheValue)
	assert.Nil(t, err)

	cache1 := mockcache.NewMockSetterCacheInterface[[]byte](ctrl)
	cache1.EXPECT().GetWithTTL(ctx, "my-key").Return(cacheValueBytes, 5*time.Second, nil)

	marshaler := NewCache[*testCacheValue](cache1)

	// When
	value, ttl, err := marshaler.GetWithTTL(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
	assert.Equal(t, 5*time.Second, ttl)
}

func TestCacheGetWithTTLWhenNotSupported(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := mockcache.NewMockCacheInterface[[]byte](ctrl)

	marshaler := NewCache[*testCacheValue](cache1)

	// When
	value, ttl, err := marshaler.GetWithTTL(ctx, "my-key")

	// Then
	assert.Nil(t, value)
	assert.Equal(t, time.Duration(0), ttl)
	assert.ErrorIs(t, err, store.ErrNotSupported)
}

func TestCacheSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &testCacheValue{Hello: "world"}
	cacheValueBytes, err := msgpack.Marshal(cacheValue)
	assert.Nil(t, err)

	cache1 := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache1.EXPECT().Set(ctx, "my-key", cacheValueBytes, store.OptionsMatcher{
		Expiration: 5 * time.Second,
	}).Return(nil)

	marshaler := NewCache[*testCacheValue](cache1)

	// When
	err = marshaler.Set(ctx, "my-key", cacheValue, store.WithExpiration(5*time.Second))

	// Then
	assert.Nil(t, err)
}

func TestCacheDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(nil)

	marshaler := NewCache[*testCacheValue](cache1)

	// When - Then
	assert.Nil(t, marshaler.Delete(ctx, "my-key"))
}

func TestCacheInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache1.EXPECT().Invalidate(ctx, gomock.Any()).Return(nil)

	marshaler := NewCache[*testCacheValue](cache1)

	// When - Then
	assert.Nil(t, marshaler.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"})))
}

func TestCacheClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache1.EXPECT().Clear(ctx).Return(nil)

	marshaler := NewCache[*testCacheValue](cache1)

	// When - Then
	assert.Nil(t, marshaler.Clear(ctx))
}

func TestCacheAsLoadableCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &testCacheValue{Hello: "world"}

	cache1 := mockcache.NewMockCacheInterface[[]byte](ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return(nil, errors.New("unable to find in cache"))
	cache1.EXPECT().Set(ctx, "my-key", gomock.Any(), gomock.Any()).Return(nil)

	loadFunc := func(_ context.Context, _ any) (*testCacheValue, []store.Option, error) {
		return cacheValue, []store.Option{store.WithExpiration(1 * time.Hour)}, nil
	}

	// When
	loadable := cache.NewLoadable[*testCacheValue](loadFunc, NewCache[*testCacheValue](cache1))
	defer loadable.Close()

	value, err := loadable.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}
