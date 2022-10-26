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
	Set(ctx context.Context, key string, values any, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	FlushAll(ctx context.Context) *redis.StatusCmd
	SAdd(ctx context.Context, key string, members ...any) *redis.IntCmd
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
func NewRedisCluster(client RedisClusterClientInterface, options ...Option) *RedisClusterStore {
	return &RedisClusterStore{
		clusclient: client,
		options:    applyOptions(options...),
	}
}

// Get returns data stored from a given key
func (s *RedisClusterStore) Get(ctx context.Context, key any) (any, error) {
	object, err := s.clusclient.Get(ctx, key.(string)).Result()
	if err == redis.Nil {
		return nil, NotFoundWithCause(err)
	}
	return object, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *RedisClusterStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	object, err := s.clusclient.Get(ctx, key.(string)).Result()
	if err == redis.Nil {
		return nil, 0, NotFoundWithCause(err)
	}
	if err != nil {
		return nil, 0, err
	}

	ttl, err := s.clusclient.TTL(ctx, key.(string)).Result()
	if err != nil {
		return nil, 0, err
	}

	return object, ttl, err
}

// Set defines data in Redis for given key identifier
func (s *RedisClusterStore) Set(ctx context.Context, key any, value any, options ...Option) error {
	opts := applyOptionsWithDefault(s.options, options...)

	err := s.clusclient.Set(ctx, key.(string), value, opts.expiration).Err()
	if err != nil {
		return err
	}

	if tags := opts.tags; len(tags) > 0 {
		s.setTags(ctx, key, tags)
	}

	return nil
}

func (s *RedisClusterStore) setTags(ctx context.Context, key any, tags []string) {
	for _, tag := range tags {
		tagKey := fmt.Sprintf(RedisTagPattern, tag)
		s.clusclient.SAdd(ctx, tagKey, key.(string))
		s.clusclient.Expire(ctx, tagKey, 720*time.Hour)
	}
}

// Delete removes data from Redis for given key identifier
func (s *RedisClusterStore) Delete(ctx context.Context, key any) error {
	_, err := s.clusclient.Del(ctx, key.(string)).Result()
	return err
}

// Invalidate invalidates some cache data in Redis for given options
func (s *RedisClusterStore) Invalidate(ctx context.Context, options ...InvalidateOption) error {
	opts := ApplyInvalidateOptions(options...)

	if tags := opts.Tags; len(tags) > 0 {
		for _, tag := range tags {
			tagKey := fmt.Sprintf(RedisTagPattern, tag)
			cacheKeys, err := s.clusclient.SMembers(ctx, tagKey).Result()
			if err != nil {
				continue
			}

			for _, cacheKey := range cacheKeys {
				s.Delete(ctx, cacheKey)
			}

			s.Delete(ctx, tagKey)
		}
	}

	return nil
}

// Clear resets all data in the store
func (s *RedisClusterStore) Clear(ctx context.Context) error {
	if err := s.clusclient.FlushAll(ctx).Err(); err != nil {
		return err
	}

	return nil
}

// GetType returns the store type
func (s *RedisClusterStore) GetType() string {
	return RedisClusterType
}
