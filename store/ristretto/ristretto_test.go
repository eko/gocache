package ristretto

import (
	"context"
	"fmt"
	"testing"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/stretchr/testify/assert"
)

func TestNewRistretto(t *testing.T) {
	// Given
	client := NewMockRistrettoClientInterface[string, []byte](t)

	// When
	store := NewRistretto(client, lib_store.WithCost(8))

	// Then
	assert.IsType(t, new(RistrettoStore[string, []byte]), store)
	assert.Equal(t, client, store.client)
	assert.Equal(t, &lib_store.Options{Cost: 8}, store.options)
}

func TestRistrettoGet(t *testing.T) {
	// Given
	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockRistrettoClientInterface[string, string](t)
	client.EXPECT().Get(cacheKey).Return(cacheValue, true)

	store := NewRistretto(client)

	// When
	value, err := store.Get(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestRistrettoGetWhenError(t *testing.T) {
	// Given
	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockRistrettoClientInterface[string, []byte](t)
	client.EXPECT().Get(cacheKey).Return(nil, false)

	store := NewRistretto(client)

	// When
	value, err := store.Get(ctx, cacheKey)

	// Then
	assert.Nil(t, value)
	assert.IsType(t, &lib_store.NotFound{}, err)
}

func TestRistrettoGetWithTTL(t *testing.T) {
	// Given
	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockRistrettoClientInterface[string, string](t)
	client.EXPECT().Get(cacheKey).Return(cacheValue, true)
	client.EXPECT().GetTTL(cacheKey).Return(time.Minute, true)

	store := NewRistretto(client)

	// When
	value, ttl, err := store.GetWithTTL(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
	assert.Equal(t, time.Minute, ttl)
}

func TestRistrettoGetWithTTLWhenError(t *testing.T) {
	// Given
	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockRistrettoClientInterface[string, []byte](t)
	client.EXPECT().Get(cacheKey).Return(nil, false)

	store := NewRistretto(client)

	// When
	value, ttl, err := store.GetWithTTL(ctx, cacheKey)

	// Then
	assert.Nil(t, value)
	assert.IsType(t, &lib_store.NotFound{}, err)
	assert.Equal(t, 0*time.Second, ttl)
}

func TestRistrettoSet(t *testing.T) {
	// Given
	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockRistrettoClientInterface[string, string](t)
	client.EXPECT().SetWithTTL(cacheKey, cacheValue, int64(4), 0*time.Second).Return(true)

	store := NewRistretto(client, lib_store.WithCost(7))

	// When
	err := store.Set(ctx, cacheKey, cacheValue, lib_store.WithCost(4))

	// Then
	assert.Nil(t, err)
}

func TestRistrettoSetWhenNoOptionsGiven(t *testing.T) {
	// Given
	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockRistrettoClientInterface[string, string](t)
	client.EXPECT().SetWithTTL(cacheKey, cacheValue, int64(7), 0*time.Second).Return(true)

	store := NewRistretto(client, lib_store.WithCost(7))

	// When
	err := store.Set(ctx, cacheKey, cacheValue)

	// Then
	assert.Nil(t, err)
}

func TestRistrettoSetWhenError(t *testing.T) {
	// Given
	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockRistrettoClientInterface[string, string](t)
	client.EXPECT().SetWithTTL(cacheKey, cacheValue, int64(7), 0*time.Second).Return(false)

	store := NewRistretto(client, lib_store.WithCost(7))

	// When
	err := store.Set(ctx, cacheKey, cacheValue)

	// Then
	assert.Equal(t, fmt.Errorf("An error has occurred while setting value '%v' on key '%v'", cacheValue, cacheKey), err)
}

func TestRistrettoSetWithSynchronousSet(t *testing.T) {
	// Given
	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockRistrettoClientInterface[string, []byte](t)
	client.EXPECT().SetWithTTL(cacheKey, cacheValue, int64(7), 0*time.Second).Return(true)
	client.EXPECT().Wait()

	store := NewRistretto(client, lib_store.WithCost(7), lib_store.WithSynchronousSet())

	// When
	err := store.Set(ctx, cacheKey, cacheValue, lib_store.WithSynchronousSet())

	// Then
	assert.Nil(t, err)
}

func TestRistrettoSetWithTags(t *testing.T) {
	// Given
	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockRistrettoClientInterface[string, []byte](t)
	client.EXPECT().SetWithTTL(cacheKey, cacheValue, int64(0), 0*time.Second).Return(true)
	client.EXPECT().Get("gocache_tag_tag1").Return(nil, true)
	client.EXPECT().SetWithTTL("gocache_tag_tag1", []byte(",my-key"), int64(0), 720*time.Hour).Return(true)

	store := NewRistretto(client)

	// When
	err := store.Set(ctx, cacheKey, cacheValue, lib_store.WithTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestRistrettoSetWithTagsWhenAlreadyInserted(t *testing.T) {
	// Given
	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockRistrettoClientInterface[string, []byte](t)
	client.EXPECT().SetWithTTL(cacheKey, cacheValue, int64(0), 0*time.Second).Return(true)
	client.EXPECT().Get("gocache_tag_tag1").Return([]byte("my-key,a-second-key"), true)
	client.EXPECT().SetWithTTL("gocache_tag_tag1", []byte("my-key,a-second-key"), int64(0), 720*time.Hour).Return(true)

	store := NewRistretto(client)

	// When
	err := store.Set(ctx, cacheKey, cacheValue, lib_store.WithTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestRistrettoDelete(t *testing.T) {
	// Given
	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockRistrettoClientInterface[string, []byte](t)
	client.EXPECT().Del(cacheKey)

	store := NewRistretto(client)

	// When
	err := store.Delete(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
}

func TestRistrettoInvalidate(t *testing.T) {
	// Given
	ctx := context.Background()

	cacheKeys := []byte("a23fdf987h2svc23,jHG2372x38hf74")

	client := NewMockRistrettoClientInterface[string, []byte](t)
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, true)
	client.EXPECT().Del("a23fdf987h2svc23")
	client.EXPECT().Del("jHG2372x38hf74")

	store := NewRistretto(client)

	// When
	err := store.Invalidate(ctx, lib_store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestRistrettoInvalidateWhenError(t *testing.T) {
	// Given
	ctx := context.Background()

	cacheKeys := []byte("a23fdf987h2svc23,jHG2372x38hf74")

	client := NewMockRistrettoClientInterface[string, []byte](t)
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, false)

	store := NewRistretto(client)

	// When
	err := store.Invalidate(ctx, lib_store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestRistrettoClear(t *testing.T) {
	// Given
	ctx := context.Background()

	client := NewMockRistrettoClientInterface[string, []byte](t)
	client.EXPECT().Clear()

	store := NewRistretto(client)

	// When
	err := store.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestRistrettoGetType(t *testing.T) {
	// Given
	client := NewMockRistrettoClientInterface[string, []byte](t)

	store := NewRistretto(client)

	// When - Then
	assert.Equal(t, RistrettoType, store.GetType())
}
