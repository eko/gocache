package bigcache

import (
	"context"
	"errors"
	"testing"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewBigcache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockBigcacheClientInterface(ctrl)

	// When
	store := NewBigcache(client)

	// Then
	assert.IsType(t, new(BigcacheStore), store)
	assert.Equal(t, client, store.client)
	assert.Equal(t, new(lib_store.Options), store.options)
}

func TestBigcacheGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockBigcacheClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(cacheValue, nil)

	store := NewBigcache(client)

	// When
	value, err := store.Get(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestBigcacheGetWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	expectedErr := errors.New("an unexpected error occurred")

	client := NewMockBigcacheClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(nil, expectedErr)

	store := NewBigcache(client)

	// When
	value, err := store.Get(ctx, cacheKey)

	// Then
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, value)
}

func TestBigcacheGetWithTTL(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	client := NewMockBigcacheClientInterface(ctrl)
	store := NewBigcache(client)

	expectedErr := errors.New("method not implemented for codec, use Get() instead")

	// When
	value, _, err := store.GetWithTTL(ctx, cacheKey)

	// Then
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, value)
}

func TestBigcacheSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockBigcacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue).Return(nil)

	store := NewBigcache(client)

	// When
	err := store.Set(ctx, cacheKey, cacheValue)

	// Then
	assert.Nil(t, err)
}

func TestBigcacheSetString(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	// The value is string when failback from Redis
	cacheValue := "my-cache-value"

	client := NewMockBigcacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, []byte(cacheValue)).Return(nil)

	store := NewBigcache(client)

	// When
	err := store.Set(ctx, cacheKey, cacheValue)

	// Then
	assert.Nil(t, err)
}

func TestBigcacheSetWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	expectedErr := errors.New("an unexpected error occurred")

	client := NewMockBigcacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue).Return(expectedErr)

	store := NewBigcache(client)

	// When
	err := store.Set(ctx, cacheKey, cacheValue)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestBigcacheSetWithTags(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockBigcacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue).Return(nil)
	client.EXPECT().Get("gocache_tag_tag1").Return(nil, nil)
	client.EXPECT().Set("gocache_tag_tag1", []byte("my-key")).Return(nil)

	store := NewBigcache(client)

	// When
	err := store.Set(ctx, cacheKey, cacheValue, lib_store.WithTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestBigcacheSetWithTagsWhenAlreadyInserted(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockBigcacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue).Return(nil)
	client.EXPECT().Get("gocache_tag_tag1").Return([]byte("my-key,a-second-key"), nil)
	client.EXPECT().Set("gocache_tag_tag1", []byte("my-key,a-second-key")).Return(nil)

	store := NewBigcache(client)

	// When
	err := store.Set(ctx, cacheKey, cacheValue, lib_store.WithTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestBigcacheDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockBigcacheClientInterface(ctrl)
	client.EXPECT().Delete(cacheKey).Return(nil)

	store := NewBigcache(client)

	// When
	err := store.Delete(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
}

func TestBigcacheDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unable to delete key")

	cacheKey := "my-key"

	client := NewMockBigcacheClientInterface(ctrl)
	client.EXPECT().Delete(cacheKey).Return(expectedErr)

	store := NewBigcache(client)

	// When
	err := store.Delete(ctx, cacheKey)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestBigcacheInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := []byte("a23fdf987h2svc23,jHG2372x38hf74")

	client := NewMockBigcacheClientInterface(ctrl)
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, nil)
	client.EXPECT().Delete("a23fdf987h2svc23").Return(nil)
	client.EXPECT().Delete("jHG2372x38hf74").Return(nil)

	store := NewBigcache(client)

	// When
	err := store.Invalidate(ctx, lib_store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestBigcacheInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := []byte("a23fdf987h2svc23,jHG2372x38hf74")

	client := NewMockBigcacheClientInterface(ctrl)
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, nil)
	client.EXPECT().Delete("a23fdf987h2svc23").Return(errors.New("unexpected error"))
	client.EXPECT().Delete("jHG2372x38hf74").Return(nil)

	store := NewBigcache(client)

	// When
	err := store.Invalidate(ctx, lib_store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestBigcacheClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	client := NewMockBigcacheClientInterface(ctrl)
	client.EXPECT().Reset().Return(nil)

	store := NewBigcache(client)

	// When
	err := store.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestBigcacheClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("an unexpected error occurred")

	client := NewMockBigcacheClientInterface(ctrl)
	client.EXPECT().Reset().Return(expectedErr)

	store := NewBigcache(client)

	// When
	err := store.Clear(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestBigcacheGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockBigcacheClientInterface(ctrl)

	store := NewBigcache(client)

	// When - Then
	assert.Equal(t, BigcacheType, store.GetType())
}
