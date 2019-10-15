package store

import (
	"time"

	"github.com/go-redis/redis/v7"
)

// RedisClientInterface represents a go-redis/redis client
type RedisClientInterface interface {
	Get(key string) *redis.StringCmd
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}

const (
	RedisType = "redis"
)

// RedisStore is a store for Redis
type RedisStore struct {
	client RedisClientInterface
	options *Options
}

// NewRedis creates a new store to Redis instance(s)
func NewRedis(client RedisClientInterface, options *Options) *RedisStore {
	if options == nil {
		options = &Options{}
	}

	return &RedisStore{
		client: client,
		options: options,
	}
}

// Get returns data stored from a given key
func (s *RedisStore) Get(key interface{}) (interface{}, error) {
	return s.client.Get(key.(string)).Result()
}

// Set defines data in Redis for given key idntifier
func (s *RedisStore) Set(key interface{}, value interface{}, options *Options) error {
	if options == nil {
		options = s.options
	}

	return s.client.Set(key.(string), value, options.ExpirationValue()).Err()
}

// GetType returns the store type
func (s *RedisStore) GetType() string {
	return RedisType
}
