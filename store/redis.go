package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
)

// RedisClientInterface represents a go-redis/redis client
type RedisClientInterface interface {
	Get(key string) *redis.StringCmd
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(keys ...string) *redis.IntCmd
}

const (
	// RedisType represents the storage type as a string value
	RedisType = "redis"
	// RedisTagPattern represents the tag pattern to be used as a key in specified storage
	RedisTagPattern = "gocache_tag_%s"
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
		var tagKey = fmt.Sprintf(RedisTagPattern, tag)
		var cacheKeys = []string{}

		if result, err := s.Get(tagKey); err == nil {
			if bytes, ok := result.([]byte); ok {
				cacheKeys = strings.Split(string(bytes), ",")
			}
		}

		var alreadyInserted = false
		for _, cacheKey := range cacheKeys {
			if cacheKey == key.(string) {
				alreadyInserted = true
				break
			}
		}

		if !alreadyInserted {
			cacheKeys = append(cacheKeys, key.(string))
		}

		s.Set(tagKey, []byte(strings.Join(cacheKeys, ",")), &Options{
			Expiration: 720 * time.Hour,
		})
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
			var tagKey = fmt.Sprintf(RedisTagPattern, tag)
			result, err := s.Get(tagKey)
			if err != nil {
				return nil
			}

			var cacheKeys = []string{}
			if bytes, ok := result.([]byte); ok {
				cacheKeys = strings.Split(string(bytes), ",")
			}

			for _, cacheKey := range cacheKeys {
				s.Delete(cacheKey)
			}
		}
	}

	return nil
}

// GetType returns the store type
func (s *RedisStore) GetType() string {
	return RedisType
}
