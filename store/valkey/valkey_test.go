package valkey

import (
	"context"
	"testing"
	"time"

	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/mock"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewValkey(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	// valkey mock client
	client := mock.NewClient(ctrl)

	// When
	store := NewValkey(client, lib_store.WithExpiration(6*time.Second), lib_store.WithClientSideCaching(time.Second*8))

	// Then
	assert.IsType(t, new(ValkeyStore), store)
	assert.Equal(t, client, store.client)
	assert.Equal(t, &lib_store.Options{Expiration: 6 * time.Second, ClientSideCacheExpiration: time.Second * 8}, store.options)
}

func TestValkeyGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// valkey mock client
	client := mock.NewClient(ctrl)
	client.EXPECT().DoCache(ctx, mock.Match("GET", "my-key"), defaultClientSideCacheExpiration).Return(mock.Result(mock.ValkeyString("my-value")))

	store := NewValkey(client)

	// When
	value, err := store.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, value, "my-value")
}

func TestValkeyGetNotFound(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// valkey mock client
	client := mock.NewClient(ctrl)
	client.EXPECT().DoCache(ctx, mock.Match("GET", "my-key"), defaultClientSideCacheExpiration).Return(mock.Result(mock.ValkeyNil()))

	store := NewValkey(client)

	// When
	value, err := store.Get(ctx, "my-key")

	// Then
	assert.NotNil(t, err)
	assert.Equal(t, value, "")
}

func TestValkeySet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	// valkey mock client
	client := mock.NewClient(ctrl)
	client.EXPECT().Do(ctx, mock.Match("SET", cacheKey, cacheValue, "EX", "10")).Return(mock.Result(mock.ValkeyString("")))

	store := NewValkey(client, lib_store.WithExpiration(time.Second*10))

	// When
	err := store.Set(ctx, cacheKey, cacheValue)

	// Then
	assert.Nil(t, err)
}

func TestValkeySetWhenNoOptionsGiven(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := mock.NewClient(ctrl)
	client.EXPECT().Do(ctx, mock.Match("SET", cacheKey, cacheValue, "EX", "6")).Return(mock.Result(mock.ValkeyString("")))

	store := NewValkey(client, lib_store.WithExpiration(6*time.Second))

	// When
	err := store.Set(ctx, cacheKey, cacheValue)

	// Then
	assert.Nil(t, err)
}

func TestValkeySetWithTags(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := mock.NewClient(ctrl)
	client.EXPECT().Do(ctx, mock.Match("SET", cacheKey, cacheValue, "EX", "10")).Return(mock.Result(mock.ValkeyString("")))
	client.EXPECT().DoMulti(ctx,
		mock.Match("SADD", "gocache_tag_tag1", "my-key"),
		mock.Match("EXPIRE", "gocache_tag_tag1", "2592000"),
	).Return([]valkey.ValkeyResult{
		mock.Result(mock.ValkeyString("")),
		mock.Result(mock.ValkeyString("")),
	})

	store := NewValkey(client, lib_store.WithExpiration(time.Second*10))

	// When
	err := store.Set(ctx, cacheKey, cacheValue, lib_store.WithTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestValkeyDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	client := mock.NewClient(ctrl)
	client.EXPECT().Do(ctx, mock.Match("DEL", cacheKey)).Return(mock.Result(mock.ValkeyInt64(1)))

	store := NewValkey(client)

	// When
	err := store.Delete(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
}

func TestValkeyInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	client := mock.NewClient(ctrl)
	client.EXPECT().Do(ctx, mock.Match("SMEMBERS", "gocache_tag_tag1")).Return(mock.Result(mock.ValkeyArray()))
	client.EXPECT().Do(ctx, mock.Match("DEL", "gocache_tag_tag1")).Return(mock.Result(mock.ValkeyInt64(1)))

	store := NewValkey(client)

	// When
	err := store.Invalidate(ctx, lib_store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestValkeyClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	client := mock.NewClient(ctrl)
	client.EXPECT().Nodes().Return(map[string]valkey.Client{
		"client1": client,
	})
	client.EXPECT().Do(ctx, mock.Match("ROLE")).Return(mock.Result(mock.ValkeyArray(mock.ValkeyString("master"))))
	client.EXPECT().Do(ctx, mock.Match("FLUSHALL")).Return(mock.Result(mock.ValkeyString("")))

	store := NewValkey(client)

	// When
	err := store.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestValkeyGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := mock.NewClient(ctrl)

	store := NewValkey(client)

	// When - Then
	assert.Equal(t, ValkeyType, store.GetType())
}
