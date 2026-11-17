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

func TestValkeyGetWithTTL(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name       string
		get        valkey.ValkeyMessage
		valkeyTTL  int64
		wantTTL    time.Duration
		wantMiss   bool
		wantGetErr bool
	}{
		{name: "returns value and TTL", get: mock.ValkeyString("my-value"), valkeyTTL: 10, wantTTL: 10 * time.Second},
		{name: "key without expiry is not a miss", get: mock.ValkeyString("my-value"), valkeyTTL: -1},
		{name: "key about to expire is a miss", get: mock.ValkeyString("my-value"), valkeyTTL: 0, wantMiss: true},
		{name: "key expired after get is a miss", get: mock.ValkeyString("my-value"), valkeyTTL: -2, wantMiss: true},
		{name: "get error is not turned into a miss", get: mock.ValkeyError("WRONGTYPE"), valkeyTTL: 0, wantGetErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			ctrl := gomock.NewController(t)

			// valkey mock client
			client := mock.NewClient(ctrl)
			client.EXPECT().DoCache(ctx, mock.Match("GET", "my-key"), defaultClientSideCacheExpiration).Return(mock.Result(tc.get))
			client.EXPECT().DoCache(ctx, mock.Match("TTL", "my-key"), defaultClientSideCacheExpiration).Return(mock.Result(mock.ValkeyInt64(tc.valkeyTTL))).MaxTimes(1)

			store := NewValkey(client)

			// When
			value, ttl, err := store.GetWithTTL(ctx, "my-key")

			// Then
			switch {
			case tc.wantMiss:
				assert.ErrorIs(t, err, lib_store.NotFound{})
				assert.Nil(t, value)
				assert.Equal(t, time.Duration(0), ttl)
			case tc.wantGetErr:
				var valkeyErr *valkey.ValkeyError
				assert.ErrorAs(t, err, &valkeyErr)
				assert.NotErrorIs(t, err, lib_store.NotFound{})
			default:
				assert.NoError(t, err)
				assert.Equal(t, "my-value", value)
				if tc.wantTTL != 0 {
					assert.Equal(t, tc.wantTTL, ttl)
				}
			}
		})
	}
}

func TestValkeyGetWithTTLNotFound(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// valkey mock client
	client := mock.NewClient(ctrl)
	client.EXPECT().DoCache(ctx, mock.Match("GET", "my-key"), defaultClientSideCacheExpiration).Return(mock.Result(mock.ValkeyNil()))

	store := NewValkey(client)

	// When
	_, ttl, err := store.GetWithTTL(ctx, "my-key")

	// Then
	assert.ErrorIs(t, err, lib_store.NotFound{})
	assert.Equal(t, time.Duration(0), ttl)
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

func TestValkeySetWhenValueTypeIsNotSupported(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	store := NewValkey(mock.NewClient(ctrl))

	// When
	err := store.Set(ctx, "my-key", 42)

	// Then
	assert.EqualError(t, err, "value type not supported by Valkey store: int")
}
