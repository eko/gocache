package store

import (
	"errors"
	"testing"
	"time"

	mocksStore "github.com/eko/gocache/test/mocks/store/clients"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewGoCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	options := &Options{
		Cost: 8,
	}

	// When
	store := NewGoCache(client, options)

	// Then
	assert.IsType(t, new(GoCacheStore), store)
	assert.Equal(t, client, store.client)
	assert.Equal(t, options, store.options)
}

func TestGoCacheGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(cacheValue, true)

	store := NewGoCache(client, nil)

	// When
	value, err := store.Get(cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestGoCacheGetWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "my-key"

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(nil, false)

	store := NewGoCache(client, nil)

	// When
	value, err := store.Get(cacheKey)

	// Then
	assert.Nil(t, value)
	assert.Equal(t, errors.New("Value not found in GoCache store"), err)
}

func TestGoCacheGetWithTTL(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().GetWithExpiration(cacheKey).Return(cacheValue, time.Now(), true)

	store := NewGoCache(client, nil)

	// When
	value, ttl, err := store.GetWithTTL(cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
	assert.Equal(t, int64(0), ttl.Milliseconds())
}

func TestGoCacheGetWithTTLWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "my-key"

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().GetWithExpiration(cacheKey).Return(nil, time.Now(), false)

	store := NewGoCache(client, nil)

	// When
	value, ttl, err := store.GetWithTTL(cacheKey)

	// Then
	assert.Nil(t, value)
	assert.Equal(t, errors.New("Value not found in GoCache store"), err)
	assert.Equal(t, 0*time.Second, ttl)
}

func TestGoCacheSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"
	options := &Options{}

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue, 0*time.Second)

	store := NewGoCache(client, options)

	// When
	err := store.Set(cacheKey, cacheValue, &Options{
		Cost: 4,
	})

	// Then
	assert.Nil(t, err)
}

func TestGoCacheSetWhenNoOptionsGiven(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"
	options := &Options{}

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue, 0*time.Second)

	store := NewGoCache(client, options)

	// When
	err := store.Set(cacheKey, cacheValue, nil)

	// Then
	assert.Nil(t, err)
}

func TestGoCacheSetWithTags(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue, 0*time.Second)
	client.EXPECT().Get("gocache_tag_tag1").Return(nil, true)
	var cacheKeys = map[string]struct{}{"my-key": {}}
	client.EXPECT().Set("gocache_tag_tag1", cacheKeys, 720*time.Hour)

	store := NewGoCache(client, nil)

	// When
	err := store.Set(cacheKey, cacheValue, &Options{Tags: []string{"tag1"}})

	// Then
	assert.Nil(t, err)
}

func TestGoCacheSetWithTagsWhenAlreadyInserted(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue, 0*time.Second)

	var cacheKeys = map[string]struct{}{"my-key": {}, "a-second-key": {}}
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, true)

	store := NewGoCache(client, nil)

	// When
	err := store.Set(cacheKey, cacheValue, &Options{Tags: []string{"tag1"}})

	// Then
	assert.Nil(t, err)
}

func TestGoCacheDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "my-key"

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Delete(cacheKey)

	store := NewGoCache(client, nil)

	// When
	err := store.Delete(cacheKey)

	// Then
	assert.Nil(t, err)
}

func TestGoCacheInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	options := InvalidateOptions{
		Tags: []string{"tag1"},
	}

	var cacheKeys = map[string]struct{}{"a23fdf987h2svc23": {}, "jHG2372x38hf74": {}}

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, true)
	client.EXPECT().Delete("a23fdf987h2svc23")
	client.EXPECT().Delete("jHG2372x38hf74")

	store := NewGoCache(client, nil)

	// When
	err := store.Invalidate(options)

	// Then
	assert.Nil(t, err)
}

func TestGoCacheInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	options := InvalidateOptions{
		Tags: []string{"tag1"},
	}

	cacheKeys := []byte("a23fdf987h2svc23,jHG2372x38hf74")

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, false)

	store := NewGoCache(client, nil)

	// When
	err := store.Invalidate(options)

	// Then
	assert.Nil(t, err)
}

func TestGoCacheClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Flush()

	store := NewGoCache(client, nil)

	// When
	err := store.Clear()

	// Then
	assert.Nil(t, err)
}

func TestGoCacheGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocksStore.NewMockGoCacheClientInterface(ctrl)

	store := NewGoCache(client, nil)

	// When - Then
	assert.Equal(t, GoCacheType, store.GetType())
}
