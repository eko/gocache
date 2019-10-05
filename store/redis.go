package store

import (
	"time"

	"github.com/go-redis/redis/v7"
)

const (
	RedisType = "redis"
)

// RedisStore is a store for Redis
type RedisStore struct {
	client *redis.Client
}

// NewRedis creates a new store to Redis instance(s)
func NewRedis(client *redis.Client) *RedisStore {
	return &RedisStore{
		client: client,
	}
}

// Get returns data stored from a given key
func (s *RedisStore) Get(key interface{}) (interface{}, error) {
	return s.client.Get(key.(string)).Result()
}

// Set defines data in Redis for given key idntifier
func (s *RedisStore) Set(key interface{}, value interface{}, expiration time.Duration) error {
	return s.client.Set(key.(string), value, expiration).Err()
}

// GetType returns the store type
func (s *RedisStore) GetType() string {
	return RedisType
}
