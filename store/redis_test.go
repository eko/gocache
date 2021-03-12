package store

import (
	"context"
	"testing"
	"time"

	mocksStore "github.com/eko/gocache/test/mocks/store/clients"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewRedis(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocksStore.NewMockRedisClientInterface(ctrl)
	options := &Options{
		Expiration: 6 * time.Second,
	}

	// When
	store := NewRedis(client, options)

	// Then
	assert.IsType(t, new(RedisStore), store)
	assert.Equal(t, client, store.client)
	assert.Equal(t, options, store.options)
}

func TestRedisGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocksStore.NewMockRedisClientInterface(ctrl)
	client.EXPECT().Get(context.TODO(), "my-key").Return(&redis.StringCmd{})

	store := NewRedis(client, nil)

	// When
	value, err := store.Get(context.TODO(), "my-key")

	// Then
	assert.Nil(t, err)
	assert.NotNil(t, value)
}

func TestRedisSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"
	options := &Options{
		Expiration: 6 * time.Second,
	}

	client := mocksStore.NewMockRedisClientInterface(ctrl)
	client.EXPECT().Set(context.TODO(), "my-key", cacheValue, 5*time.Second).Return(&redis.StatusCmd{})

	store := NewRedis(client, options)

	// When
	err := store.Set(context.TODO(), cacheKey, cacheValue, &Options{
		Expiration: 5 * time.Second,
	})

	// Then
	assert.Nil(t, err)
}

func TestRedisSetWhenNoOptionsGiven(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"
	options := &Options{
		Expiration: 6 * time.Second,
	}

	client := mocksStore.NewMockRedisClientInterface(ctrl)
	client.EXPECT().Set(context.TODO(), "my-key", cacheValue, 6*time.Second).Return(&redis.StatusCmd{})

	store := NewRedis(client, options)

	// When
	err := store.Set(context.TODO(), cacheKey, cacheValue, nil)

	// Then
	assert.Nil(t, err)
}

func TestRedisSetWithTags(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := mocksStore.NewMockRedisClientInterface(ctrl)
	client.EXPECT().Set(context.TODO(), cacheKey, cacheValue, time.Duration(0)).Return(&redis.StatusCmd{})
	client.EXPECT().SAdd(context.TODO(), "gocache_tag_tag1", "my-key").Return(&redis.IntCmd{})
	client.EXPECT().Expire(context.TODO(), "gocache_tag_tag1", 720*time.Hour).Return(&redis.BoolCmd{})

	store := NewRedis(client, nil)

	// When
	err := store.Set(context.TODO(), cacheKey, cacheValue, &Options{Tags: []string{"tag1"}})

	// Then
	assert.Nil(t, err)
}

func TestRedisDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "my-key"

	client := mocksStore.NewMockRedisClientInterface(ctrl)
	client.EXPECT().Del(context.TODO(), "my-key").Return(&redis.IntCmd{})

	store := NewRedis(client, nil)

	// When
	err := store.Delete(context.TODO(), cacheKey)

	// Then
	assert.Nil(t, err)
}

func TestRedisInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	options := InvalidateOptions{
		Tags: []string{"tag1"},
	}

	cacheKeys := &redis.StringSliceCmd{}

	client := mocksStore.NewMockRedisClientInterface(ctrl)
	client.EXPECT().SMembers(context.TODO(), "gocache_tag_tag1").Return(cacheKeys)
	client.EXPECT().Del(context.TODO(), "gocache_tag_tag1").Return(&redis.IntCmd{})

	store := NewRedis(client, nil)

	// When
	err := store.Invalidate(context.TODO(), options)

	// Then
	assert.Nil(t, err)
}

func TestRedisClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocksStore.NewMockRedisClientInterface(ctrl)
	client.EXPECT().FlushAll(context.TODO()).Return(&redis.StatusCmd{})

	store := NewRedis(client, nil)

	// When
	err := store.Clear(context.TODO())

	// Then
	assert.Nil(t, err)
}

func TestRedisGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocksStore.NewMockRedisClientInterface(ctrl)

	store := NewRedis(client, nil)

	// When - Then
	assert.Equal(t, RedisType, store.GetType())
}
