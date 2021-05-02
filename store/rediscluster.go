package store

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClusterClientInterface represents a go-redis/redis clusclient
type RedisClusterClientInterface interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	TTL(ctx context.Context, key string) *redis.DurationCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Set(ctx context.Context, key string, values interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	FlushAll(ctx context.Context) *redis.StatusCmd
	SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	SMembers(ctx context.Context, key string) *redis.StringSliceCmd
}

const (
	// RedisType represents the storage type as a string value
	RedisClusterType = "rediscluster"
	// RedisTagPattern represents the tag pattern to be used as a key in specified storage
	RedisClusterTagPattern = "gocache_tag_%s"
)

// RedisStore is a store for Redis
type RedisClusterStore struct {
	clusclient RedisClusterClientInterface
	options    *Options
}

// NewRedis creates a new store to Redis instance(s)
func NewRedisCluster(client RedisClusterClientInterface, options *Options) *RedisClusterStore {
	if options == nil {
		options = &Options{}
	}
	if options.Ctx == nil {
		options.Ctx = context.Background()
	}

	return &RedisClusterStore{
		clusclient: client,
		options:    options,
	}
}

// Get returns data stored from a given key
func (s *RedisClusterStore) Get(key interface{}) (interface{}, error) {
	return s.clusclient.Get(s.options.Ctx, key.(string)).Result()
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *RedisClusterStore) GetWithTTL(key interface{}) (interface{}, time.Duration, error) {
	object, err := s.clusclient.Get(s.options.Ctx, key.(string)).Result()
	if err != nil {
		return nil, 0, err
	}

	ttl, err := s.clusclient.TTL(s.options.Ctx, key.(string)).Result()
	if err != nil {
		return nil, 0, err
	}

	return object, ttl, err
}

// Set defines data in Redis for given key identifier
func (s *RedisClusterStore) Set(key interface{}, value interface{}, options *Options) error {
	if options == nil {
		options = s.options
	}
	if options.Ctx == nil {
		options.Ctx = context.Background()
	}

	err := s.clusclient.Set(options.Ctx, key.(string), value, options.ExpirationValue()).Err()
	if err != nil {
		return err
	}

	if tags := options.TagsValue(); len(tags) > 0 {
		s.setTags(key, tags)
	}

	return nil
}

func (s *RedisClusterStore) setTags(key interface{}, tags []string) {
	for _, tag := range tags {
		tagKey := fmt.Sprintf(RedisTagPattern, tag)
		s.clusclient.SAdd(s.options.Ctx, tagKey, key.(string))
		s.clusclient.Expire(s.options.Ctx, tagKey, 720*time.Hour)
	}
}

// Delete removes data from Redis for given key identifier
func (s *RedisClusterStore) Delete(key interface{}) error {
	_, err := s.clusclient.Del(s.options.Ctx, key.(string)).Result()
	return err
}

// Invalidate invalidates some cache data in Redis for given options
func (s *RedisClusterStore) Invalidate(options InvalidateOptions) error {
	if tags := options.TagsValue(); len(tags) > 0 {
		for _, tag := range tags {
			tagKey := fmt.Sprintf(RedisTagPattern, tag)
			cacheKeys, err := s.clusclient.SMembers(s.options.Ctx, tagKey).Result()
			if err != nil {
				continue
			}

			for _, cacheKey := range cacheKeys {
				s.Delete(cacheKey)
			}

			s.Delete(tagKey)
		}
	}

	return nil
}

// GetType returns the store type
func (s *RedisClusterStore) GetType() string {
	return RedisClusterType
}

// Clear resets all data in the store
func (s *RedisClusterStore) Clear() error {
	if err := s.clusclient.FlushAll(s.options.Ctx).Err(); err != nil {
		return err
	}

	return nil
}
