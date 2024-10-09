package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	lib_store "github.com/eko/gocache/lib/v4/store"
)

func TestNewRedis(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockRedisClientInterface(ctrl)

	// When
	store := NewRedis(client, lib_store.WithExpiration(6*time.Second))

	// Then
	assert.IsType(t, new(RedisStore), store)
	assert.Equal(t, client, store.client)
	assert.Equal(t, &lib_store.Options{Expiration: 6 * time.Second}, store.options)
}

func TestRedisGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	client := NewMockRedisClientInterface(ctrl)
	store := NewRedis(client)

	tests := []struct {
		name      string
		key       string
		returnVal string
		returnErr error
		expectErr bool
		expectVal interface{}
	}{
		{
			name:      "Returns Value",
			key:       "my-key",
			returnVal: "value",
			returnErr: nil,
			expectErr: false,
			expectVal: "value",
		},
		{
			name:      "Key Not Found",
			key:       "non-existent-key",
			returnVal: "",
			returnErr: redis.Nil,
			expectErr: true,
			expectVal: nil,
		},
		{
			name:      "Return Error",
			key:       "my-key",
			returnVal: "",
			returnErr: fmt.Errorf("some error"),
			expectErr: true,
			expectVal: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: mock the Redis client's Get method
			client.EXPECT().Get(ctx, tt.key).Return(redis.NewStringResult(tt.returnVal, tt.returnErr))

			// When
			value, err := store.Get(ctx, tt.key)

			// Then
			if tt.expectErr {
				assert.Error(t, err)
				assert.Equal(t, tt.expectVal, value)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectVal, value)
			}
		})
	}
}

func TestRedisSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockRedisClientInterface(ctrl)
	client.EXPECT().Set(ctx, "my-key", cacheValue, 5*time.Second).Return(&redis.StatusCmd{})

	store := NewRedis(client, lib_store.WithExpiration(6*time.Second))

	// When
	err := store.Set(ctx, cacheKey, cacheValue, lib_store.WithExpiration(5*time.Second))

	// Then
	assert.Nil(t, err)
}

func TestRedisSetWhenNoOptionsGiven(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockRedisClientInterface(ctrl)
	client.EXPECT().Set(ctx, "my-key", cacheValue, 6*time.Second).Return(&redis.StatusCmd{})

	store := NewRedis(client, lib_store.WithExpiration(6*time.Second))

	// When
	err := store.Set(ctx, cacheKey, cacheValue)

	// Then
	assert.Nil(t, err)
}

func TestRedisSetWithTags(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockRedisClientInterface(ctrl)
	client.EXPECT().Set(ctx, cacheKey, cacheValue, time.Duration(0)).Return(&redis.StatusCmd{})
	client.EXPECT().SAdd(ctx, "gocache_tag_tag1", "my-key").Return(&redis.IntCmd{})
	client.EXPECT().Expire(ctx, "gocache_tag_tag1", 720*time.Hour).Return(&redis.BoolCmd{})

	store := NewRedis(client)

	// When
	err := store.Set(ctx, cacheKey, cacheValue, lib_store.WithTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestRedisDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockRedisClientInterface(ctrl)
	client.EXPECT().Del(ctx, "my-key").Return(&redis.IntCmd{})

	store := NewRedis(client)

	// When
	err := store.Delete(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
}

func TestRedisInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := &redis.StringSliceCmd{}

	client := NewMockRedisClientInterface(ctrl)
	client.EXPECT().SMembers(ctx, "gocache_tag_tag1").Return(cacheKeys)
	client.EXPECT().Del(ctx, "gocache_tag_tag1").Return(&redis.IntCmd{})

	store := NewRedis(client)

	// When
	err := store.Invalidate(ctx, lib_store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestRedisClear(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	client := NewMockRedisClientInterface(ctrl)
	store := NewRedis(client)

	tests := []struct {
		name          string
		returnValue   *redis.StatusCmd
		returnError   error
		expectError   bool
		expectedError string
	}{
		{
			name:        "Successfully clears data",
			returnValue: &redis.StatusCmd{},
			returnError: nil,
			expectError: false,
		},
		{
			name:          "Returns error on failure",
			returnValue:   redis.NewStatusResult("", fmt.Errorf("flush error")),
			returnError:   nil,
			expectError:   true,
			expectedError: "flush error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectError {
				client.EXPECT().FlushAll(ctx).Return(tt.returnValue).Times(1)
			} else {
				client.EXPECT().FlushAll(ctx).Return(tt.returnValue).Times(1)
			}

			err := store.Clear(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRedisGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockRedisClientInterface(ctrl)

	store := NewRedis(client)

	// When - Then
	assert.Equal(t, RedisType, store.GetType())
}

func TestRedisGetWithTTL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	client := NewMockRedisClientInterface(ctrl)
	store := NewRedis(client)

	t.Run("Returns Value and TTL", func(t *testing.T) {
		testReturnsValueAndTTL(t, ctx, client, store)
	})

	t.Run("Key Not Found", func(t *testing.T) {
		testKeyNotFound(t, ctx, client, store)
	})

	t.Run("Get Error", func(t *testing.T) {
		testGetError(t, ctx, client, store)
	})

	t.Run("TTL Fetch Error", func(t *testing.T) {
		testTTLFetchError(t, ctx, client, store)
	})
}

func testReturnsValueAndTTL(t *testing.T, ctx context.Context, client *MockRedisClientInterface, store *RedisStore) {
	// Given
	key := "my-key"
	returnValue := "value"
	returnTTL := 10 * time.Second
	client.EXPECT().
		Get(ctx, key).
		Return(redis.NewStringResult(returnValue, nil))
	client.EXPECT().
		TTL(ctx, key).
		Return(redis.NewDurationResult(returnTTL, nil))

	// When
	value, ttl, err := store.GetWithTTL(ctx, key)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, returnValue, value)
	assert.Equal(t, returnTTL, ttl)
}

func testKeyNotFound(t *testing.T, ctx context.Context, client *MockRedisClientInterface, store *RedisStore) {
	// Given
	key := "non-existent-key"
	client.EXPECT().
		Get(ctx, key).
		Return(redis.NewStringResult("", redis.Nil))

	// When
	value, ttl, err := store.GetWithTTL(ctx, key)

	// Then
	assert.Error(t, err)
	assert.Nil(t, value)
	assert.Equal(t, 0*time.Second, ttl)
}

func testGetError(t *testing.T, ctx context.Context, client *MockRedisClientInterface, store *RedisStore) {
	// Given
	key := "my-key"
	client.EXPECT().
		Get(ctx, key).
		Return(redis.NewStringResult("", fmt.Errorf("some error")))

	// When
	value, ttl, err := store.GetWithTTL(ctx, key)

	// Then
	assert.Error(t, err)
	assert.Equal(t, nil, value)
	assert.Equal(t, 0*time.Second, ttl)
}

func testTTLFetchError(t *testing.T, ctx context.Context, client *MockRedisClientInterface, store *RedisStore) {
	// Given
	key := "my-key"
	client.EXPECT().
		Get(ctx, key).
		Return(redis.NewStringResult("", nil))
	client.EXPECT().
		TTL(ctx, key).
		Return(redis.NewDurationResult(0, fmt.Errorf("ttl error")))

	// When
	value, ttl, err := store.GetWithTTL(ctx, key)

	// Then
	assert.Error(t, err)
	assert.Equal(t, nil, value)
	assert.Equal(t, 0*time.Second, ttl)
}
