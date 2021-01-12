package store

import (
	"fmt"
	"time"

	redis "github.com/go-redis/redis/v7"
)

// RedisClientInterface represents a go-redis/redis client
type RedisClientInterface interface {
	Get(key string) *redis.StringCmd
	HGetAll(key string) *redis.StringStringMapCmd
	TTL(key string) *redis.DurationCmd
	Expire(key string, expiration time.Duration) *redis.BoolCmd
	Set(key string, values interface{}, expiration time.Duration) *redis.StatusCmd
	HSet(key string, values ...interface{}) *redis.IntCmd
	Del(keys ...string) *redis.IntCmd
	FlushAll() *redis.StatusCmd
}

const (
	// RedisType represents the storage type as a string value
	RedisType = "redis"
	// RedisTagPattern represents the tag pattern to be used as a key in specified storage
	RedisTagPattern = "gocache_tag_%s"
	// RedisEmptyValue represents an empty value to be used in hsets when the content is not used
	RedisEmptyValue = 0
)

// RedisStore is a store for Redis
type RedisStore struct {
	client  RedisClientInterface
	options *Options
}

// NewRedis creates a new store to Redis instance(s)
func NewRedis(client RedisClientInterface, options *Options) *RedisStore {
	if options == nil {
		options = &Options{}
	}

	return &RedisStore{
		client:  client,
		options: options,
	}
}

// Get returns data stored from a given key
func (s *RedisStore) Get(key interface{}) (interface{}, error) {
	return s.client.Get(key.(string)).Result()
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *RedisStore) GetWithTTL(key interface{}) (interface{}, time.Duration, error) {
	object, err := s.client.Get(key.(string)).Result()
	if err != nil {
		return nil, 0, err
	}

	ttl, err := s.client.TTL(key.(string)).Result()
	if err != nil {
		return nil, 0, err
	}

	return object, ttl, err
}

// Set defines data in Redis for given key identifier
func (s *RedisStore) Set(key interface{}, value interface{}, options *Options) error {
	if options == nil {
		options = s.options
	}

	err := s.client.Set(key.(string), value, options.ExpirationValue()).Err()
	if err != nil {
		return err
	}

	if tags := options.TagsValue(); len(tags) > 0 {
		s.setTags(key, tags)
	}

	return nil
}

func (s *RedisStore) setTags(key interface{}, tags []string) {
	for _, tag := range tags {
		tagKey := fmt.Sprintf(RedisTagPattern, tag)
		s.client.HSet(tagKey, key.(string), RedisEmptyValue)
		s.client.Expire(tagKey, 720*time.Hour)
	}
}

// Delete removes data from Redis for given key identifier
func (s *RedisStore) Delete(key interface{}) error {
	_, err := s.client.Del(key.(string)).Result()
	return err
}

// Invalidate invalidates some cache data in Redis for given options
func (s *RedisStore) Invalidate(options InvalidateOptions) error {
	if tags := options.TagsValue(); len(tags) > 0 {
		for _, tag := range tags {
			tagKey := fmt.Sprintf(RedisTagPattern, tag)
			cacheKeys, err := s.client.HGetAll(tagKey).Result()
			if err != nil {
				continue
			}

			for cacheKey := range cacheKeys {
				s.Delete(cacheKey)
			}

			s.Delete(tagKey)
		}
	}

	return nil
}

// GetType returns the store type
func (s *RedisStore) GetType() string {
	return RedisType
}

// Clear resets all data in the store
func (s *RedisStore) Clear() error {
	if err := s.client.FlushAll().Err(); err != nil {
		return err
	}

	return nil
}
