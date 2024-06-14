package hazelcast

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	lib_store "github.com/eko/gocache/lib/v4/store"
)

func TestNewHazelcast(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	hzMap := NewMockHazelcastMapInterface(ctrl)

	// When
	store := NewHazelcast(hzMap, lib_store.WithExpiration(6*time.Second))

	// Then
	assert.IsType(t, new(HazelcastStore), store)
	assert.Equal(t, hzMap, store.hzMap)
	assert.Equal(t, &lib_store.Options{Expiration: 6 * time.Second}, store.options)
}

func TestHazelcastGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	hzMap := NewMockHazelcastMapInterface(ctrl)
	hzMap.EXPECT().Get(ctx, "my-key").Return("my-value", nil)

	store := NewHazelcast(hzMap)

	// When
	value, err := store.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.NotNil(t, value)
}

func TestHazelcastSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	hzMap := NewMockHazelcastMapInterface(ctrl)
	hzMap.EXPECT().SetWithTTL(ctx, cacheKey, cacheValue, 5*time.Second).Return(nil)

	store := NewHazelcast(hzMap, lib_store.WithExpiration(6*time.Second))

	// When
	err := store.Set(ctx, cacheKey, cacheValue, lib_store.WithExpiration(5*time.Second))

	// Then
	assert.Nil(t, err)
}

func TestHazelcastSetWhenNoOptionsGiven(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	hzMap := NewMockHazelcastMapInterface(ctrl)
	hzMap.EXPECT().SetWithTTL(ctx, cacheKey, cacheValue, 6*time.Second).Return(nil)

	store := NewHazelcast(hzMap, lib_store.WithExpiration(6*time.Second))

	// When
	err := store.Set(ctx, cacheKey, cacheValue)

	// Then
	assert.Nil(t, err)
}

func TestHazelcastSetWithTags(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	hzMap := NewMockHazelcastMapInterface(ctrl)
	hzMap.EXPECT().SetWithTTL(ctx, cacheKey, cacheValue, time.Duration(0)).Return(nil)
	hzMap.EXPECT().SetWithTTL(gomock.Any(), "gocache_tag_tag1", cacheKey, TagKeyExpiry).Return(nil)
	hzMap.EXPECT().Get(gomock.Any(), "gocache_tag_tag1").Return(nil, nil)

	store := NewHazelcast(hzMap)

	// When
	err := store.Set(ctx, cacheKey, cacheValue, lib_store.WithTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestHazelcastDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	hzMap := NewMockHazelcastMapInterface(ctrl)
	hzMap.EXPECT().Remove(ctx, "my-key").Return(0, nil)

	store := NewHazelcast(hzMap)

	// When
	err := store.Delete(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
}

func TestHazelcastInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	hzMap := NewMockHazelcastMapInterface(ctrl)
	hzMap.EXPECT().Get(ctx, "gocache_tag_tag1").Return(nil, nil)

	store := NewHazelcast(hzMap)

	// When
	err := store.Invalidate(ctx, lib_store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestHazelcastInvalidateWhenCacheKeysExist(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := "my-key0,my-key1,my-key2"

	hzMap := NewMockHazelcastMapInterface(ctrl)
	hzMap.EXPECT().Get(ctx, "gocache_tag_tag1").Return(cacheKeys, nil)
	hzMap.EXPECT().Remove(ctx, "my-key0").Return("my-value0", nil)
	hzMap.EXPECT().Remove(ctx, "my-key1").Return("my-value1", nil)
	hzMap.EXPECT().Remove(ctx, "my-key2").Return("my-value2", nil)
	hzMap.EXPECT().Remove(ctx, "gocache_tag_tag1").Return(cacheKeys, nil)

	store := NewHazelcast(hzMap)

	// When
	err := store.Invalidate(ctx, lib_store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestHazelcastClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	hzMap := NewMockHazelcastMapInterface(ctrl)
	hzMap.EXPECT().Clear(ctx).Return(nil)

	store := NewHazelcast(hzMap, lib_store.WithExpiration(6*time.Second))

	// When
	err := store.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestHazelcastType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	hzMap := NewMockHazelcastMapInterface(ctrl)

	store := NewHazelcast(hzMap)

	// When - Then
	assert.Equal(t, HazelcastType, store.GetType())
}
