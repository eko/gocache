package store

import (
	"context"
	"github.com/rueian/rueidis/mock"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewRueidis(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	// rueidis mock client
	client := mock.NewClient(ctrl)

	// clientExpiration
	clientExpiration := time.Second * 10

	// When
	store := NewRueidis(client, &clientExpiration, WithExpiration(6*time.Second))

	// Then
	assert.IsType(t, new(RueidisStore), store)
	assert.Equal(t, client, store.client)
	assert.Equal(t, &Options{expiration: 6 * time.Second}, store.options)
	assert.Equal(t, 10*time.Second, store.clientExpiration)
}

func TestRueidisGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// clientExpiration
	clientExpiration := time.Second * 10

	// rueidis mock client
	client := mock.NewClient(ctrl)
	client.EXPECT().DoCache(ctx, mock.Match("GET", "my-key"), clientExpiration).Return(mock.Result(mock.RedisString("")))

	store := NewRueidis(client, &clientExpiration)

	// When
	value, err := store.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.NotNil(t, value)
}

func TestRueidisSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	// clientExpiration
	clientExpiration := time.Second * 10

	// rueidis mock client
	client := mock.NewClient(ctrl)
	client.EXPECT().Do(ctx, mock.Match("SET", cacheKey, cacheValue, "EX", "5")).Return(mock.Result(mock.RedisString("")))

	store := NewRueidis(client, &clientExpiration)

	// When
	err := store.Set(ctx, cacheKey, cacheValue, WithExpiration(5*time.Second))

	// Then
	assert.Nil(t, err)
}
