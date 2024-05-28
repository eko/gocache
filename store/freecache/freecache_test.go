package freecache

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewFreecache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockFreecacheClientInterface(ctrl)

	// When
	store := NewFreecache(client, lib_store.WithExpiration(6*time.Second))

	// Then
	assert.IsType(t, new(FreecacheStore), store)
	assert.Equal(t, client, store.client)
	assert.Equal(t, &lib_store.Options{
		Expiration: 6 * time.Second,
	}, store.options)
}

func TestNewFreecacheDefaultOptions(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockFreecacheClientInterface(ctrl)

	// When
	store := NewFreecache(client)

	// Then
	assert.IsType(t, new(FreecacheStore), store)
	assert.Equal(t, client, store.client)
	assert.Equal(t, new(lib_store.Options), store.options)
}

func TestFreecacheGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Get([]byte("key1")).Return([]byte("val1"), nil)
	client.EXPECT().Get([]byte("key2")).Return([]byte("val2"), nil)

	s := NewFreecache(client)

	value, err := s.Get(ctx, "key1")
	assert.Nil(t, err)
	assert.Equal(t, []byte("val1"), value)

	value, err = s.Get(ctx, "key2")
	assert.Nil(t, err)
	assert.Equal(t, []byte("val2"), value)
}

func TestFreecacheGetNotFound(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Get([]byte("key1")).Return(nil, errors.New("value not found in store"))

	s := NewFreecache(client)

	value, err := s.Get(ctx, "key1")
	assert.EqualError(t, err, "value not found in store")
	assert.Nil(t, value)
}

func TestFreecacheGetWithInvalidKey(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	client := NewMockFreecacheClientInterface(ctrl)

	s := NewFreecache(client)

	value, err := s.Get(ctx, []byte("key1"))
	assert.Error(t, err, "key type not supported by Freecache store")
	assert.Nil(t, value)
}

func TestFreecacheGetWithTTL(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Get([]byte(cacheKey)).Return(cacheValue, nil)
	client.EXPECT().TTL([]byte(cacheKey)).Return(uint32(5), nil)

	store := NewFreecache(client, lib_store.WithExpiration(3*time.Second))

	// When
	value, ttl, err := store.GetWithTTL(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
	assert.Equal(t, 5*time.Second, ttl)
}

func TestFreecacheGetWithTTLWhenMissingItem(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Get([]byte(cacheKey)).Return(nil, lib_store.NotFound{})

	store := NewFreecache(client, lib_store.WithExpiration(3*time.Second))

	// When
	value, ttl, err := store.GetWithTTL(ctx, cacheKey)

	// Then
	assert.ErrorIs(t, err, lib_store.NotFound{})
	assert.Nil(t, value)
	assert.Equal(t, 0*time.Second, ttl)
}

func TestFreecacheGetWithTTLWhenErrorAtTTL(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Get([]byte(cacheKey)).Return(cacheValue, nil)
	client.EXPECT().TTL([]byte(cacheKey)).Return(uint32(0), lib_store.NotFound{})

	store := NewFreecache(client, lib_store.WithExpiration(3*time.Second))

	// When
	value, ttl, err := store.GetWithTTL(ctx, cacheKey)

	// Then
	assert.ErrorIs(t, err, lib_store.NotFound{})
	assert.Nil(t, value)
	assert.Equal(t, 0*time.Second, ttl)
}

func TestFreecacheGetWithTTLWhenInvalidKey(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	client := NewMockFreecacheClientInterface(ctrl)

	s := NewFreecache(client)

	value, ttl, err := s.GetWithTTL(ctx, []byte("key1"))
	assert.Error(t, err, "key type not supported by Freecache store")
	assert.Nil(t, value)
	assert.Equal(t, 0*time.Second, ttl)
}

func TestFreecacheSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Set([]byte(cacheKey), cacheValue, 6).Return(nil)

	s := NewFreecache(client, lib_store.WithExpiration(6*time.Second))
	err := s.Set(ctx, cacheKey, cacheValue, lib_store.WithExpiration(6*time.Second))
	assert.Nil(t, err)
}

func TestFreecacheSetWithDefaultOptions(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Set([]byte(cacheKey), cacheValue, 0).Return(nil)

	s := NewFreecache(client)
	err := s.Set(ctx, cacheKey, cacheValue)
	assert.Nil(t, err)
}

func TestFreecacheSetInvalidValue(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"
	expectedErr := errors.New("value type not supported by Freecache store")

	client := NewMockFreecacheClientInterface(ctrl)

	s := NewFreecache(client, lib_store.WithExpiration(6*time.Second))
	err := s.Set(ctx, cacheKey, cacheValue, lib_store.WithExpiration(6*time.Second))
	assert.Equal(t, expectedErr, err)
}

func TestFreecacheSetInvalidSize(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")
	expectedErr := fmt.Errorf("size of key: %v, value: %v, err: %v", cacheKey, cacheValue, errors.New(""))
	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Set([]byte(cacheKey), cacheValue, 6).Return(expectedErr)

	s := NewFreecache(client, lib_store.WithExpiration(6*time.Second))
	err := s.Set(ctx, cacheKey, cacheValue, lib_store.WithExpiration(6*time.Second))
	assert.NotNil(t, err)
}

func TestFreecacheSetInvalidKey(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := 1
	cacheValue := []byte("my-cache-value")

	expectedErr := errors.New("key type not supported by Freecache store")

	client := NewMockFreecacheClientInterface(ctrl)

	s := NewFreecache(client, lib_store.WithExpiration(6*time.Second))
	err := s.Set(ctx, cacheKey, cacheValue, lib_store.WithExpiration(6*time.Second))
	assert.Equal(t, expectedErr, err)
}

func TestFreecacheDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "key"

	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Del(gomock.Any()).Return(true)

	s := NewFreecache(client)
	err := s.Delete(ctx, cacheKey)
	assert.Nil(t, err)
}

func TestFreecacheDeleteFailed(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "key"
	expectedErr := fmt.Errorf("failed to delete key %v", cacheKey)
	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Del(gomock.Any()).Return(false)

	s := NewFreecache(client)
	err := s.Delete(ctx, cacheKey)
	assert.Equal(t, expectedErr, err)
}

func TestFreecacheDeleteInvalidKey(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := 1
	expectedErr := errors.New("key type not supported by Freecache store")
	client := NewMockFreecacheClientInterface(ctrl)

	s := NewFreecache(client)
	err := s.Delete(ctx, cacheKey)
	assert.Equal(t, expectedErr, err)
}

func TestFreecacheSetWithTags(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Set([]byte(cacheKey), cacheValue, 6).Return(nil)
	client.EXPECT().Get([]byte("freecache_tag_tag1")).MaxTimes(1).Return(nil, errors.New("value not found in store"))
	client.EXPECT().Set([]byte("freecache_tag_tag1"), []byte("my-key"), 2592000).Return(nil)

	s := NewFreecache(client, lib_store.WithExpiration(6*time.Second))
	err := s.Set(ctx, cacheKey, cacheValue, lib_store.WithExpiration(6*time.Second), lib_store.WithTags([]string{"tag1"}))
	assert.Nil(t, err)
}

func TestFreecacheInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := []byte("my-key")

	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Get([]byte("freecache_tag_tag1")).Return(cacheKeys, nil)
	client.EXPECT().Del([]byte("my-key")).Return(true)
	client.EXPECT().Del([]byte("freecache_tag_tag1")).Return(true)

	s := NewFreecache(client, lib_store.WithExpiration(6*time.Second))

	// When
	err := s.Invalidate(ctx, lib_store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestFreecacheTagsAlreadyPresent(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	oldCacheKeys := []byte("key1,key2")

	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Set([]byte(cacheKey), cacheValue, 6).Return(nil)
	client.EXPECT().Get([]byte("freecache_tag_tag1")).MaxTimes(1).Return(oldCacheKeys, nil)
	client.EXPECT().Set([]byte("freecache_tag_tag1"), []byte("key1,key2,my-key"), 2592000).Return(nil)

	s := NewFreecache(client, lib_store.WithExpiration(6*time.Second))
	err := s.Set(ctx, cacheKey, cacheValue, lib_store.WithExpiration(6*time.Second), lib_store.WithTags([]string{"tag1"}))
	assert.Nil(t, err)
}

func TestFreecacheTagsRefreshTime(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	oldCacheKeys := []byte("my-key")

	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Set([]byte(cacheKey), cacheValue, 6).Return(nil)
	client.EXPECT().Get([]byte("freecache_tag_tag1")).MaxTimes(1).Return(oldCacheKeys, nil)
	client.EXPECT().Set([]byte("freecache_tag_tag1"), []byte("my-key"), 2592000).Return(nil)

	s := NewFreecache(client, lib_store.WithExpiration(6*time.Second))
	err := s.Set(ctx, cacheKey, cacheValue, lib_store.WithExpiration(6*time.Second), lib_store.WithTags([]string{"tag1"}))
	assert.Nil(t, err)
}

func TestFreecacheInvalidateMultipleKeys(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := []byte("my-key,key1,key2")

	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Get([]byte("freecache_tag_tag1")).Return(cacheKeys, nil)
	client.EXPECT().Del([]byte("my-key")).Return(true)
	client.EXPECT().Del([]byte("key1")).Return(true)
	client.EXPECT().Del([]byte("key2")).Return(true)
	client.EXPECT().Del([]byte("freecache_tag_tag1")).Return(true)

	s := NewFreecache(client, lib_store.WithExpiration(6*time.Second))

	// When
	err := s.Invalidate(ctx, lib_store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestFreecacheFailedInvalidateMultipleKeys(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := []byte("my-key,key1,key2")

	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Get([]byte("freecache_tag_tag1")).Return(cacheKeys, nil)
	client.EXPECT().Del([]byte("my-key")).Return(false)

	s := NewFreecache(client, lib_store.WithExpiration(6*time.Second))

	// When
	err := s.Invalidate(ctx, lib_store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.EqualError(t, err, "failed to delete key my-key")
}

func TestFreecacheFailedInvalidatePattern(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := []byte("my-key,key1,key2")

	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Get([]byte("freecache_tag_tag1")).Return(cacheKeys, nil)
	client.EXPECT().Del([]byte("my-key")).Return(true)
	client.EXPECT().Del([]byte("key1")).Return(true)
	client.EXPECT().Del([]byte("key2")).Return(true)
	client.EXPECT().Del([]byte("freecache_tag_tag1")).Return(false)

	s := NewFreecache(client, lib_store.WithExpiration(6*time.Second))

	// When
	err := s.Invalidate(ctx, lib_store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.EqualError(t, err, "failed to delete key freecache_tag_tag1")
}

func TestFreecacheClearAll(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	client := NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Clear()

	s := NewFreecache(client)

	// When
	err := s.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestFreecacheGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockFreecacheClientInterface(ctrl)

	s := NewFreecache(client)

	// When
	ty := s.GetType()

	// Then
	assert.Equal(t, FreecacheType, ty)
}
