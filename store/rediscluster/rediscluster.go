package rediscluster

import (
	"context"
	"fmt"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	redis "github.com/redis/go-redis/v9"
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
	// RedisClusterType represents the storage type as a string value
	RedisClusterType = "rediscluster"
	// RedisClusterTagPattern represents the tag pattern to be used as a key in specified storage
	RedisClusterTagPattern = "gocache_tag_%s"
)

// RedisClusterStore is a store for Redis
type RedisClusterStore struct {
	clusclient RedisClusterClientInterface
	options    *lib_store.Options
}

// NewRedisCluster creates a new store to Redis cluster
func NewRedisCluster(client RedisClusterClientInterface, options ...lib_store.Option) *RedisClusterStore {
	return &RedisClusterStore{
		clusclient: client,
		options:    lib_store.ApplyOptions(options...),
	}
}

// Get returns data stored from a given key
func (s *RedisClusterStore) Get(ctx context.Context, key any) (any, error) {
	object, err := s.clusclient.Get(ctx, key.(string)).Result()
	if err == redis.Nil {
		return nil, lib_store.NotFoundWithCause(err)
	}
	return object, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *RedisClusterStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	object, err := s.clusclient.Get(ctx, key.(string)).Result()
	if err == redis.Nil {
		return nil, 0, lib_store.NotFoundWithCause(err)
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
func (s *RedisClusterStore) Set(ctx context.Context, key any, value any, options ...lib_store.Option) error {
	opts := lib_store.ApplyOptionsWithDefault(s.options, options...)

	err := s.clusclient.Set(ctx, key.(string), value, opts.Expiration).Err()
	if err != nil {
		return err
	}

	if tags := opts.Tags; len(tags) > 0 {
		s.setTags(ctx, key, tags)
	}

	return nil
}

func (s *RedisClusterStore) setTags(ctx context.Context, key any, tags []string) {
	for _, tag := range tags {
		tagKey := fmt.Sprintf(RedisClusterTagPattern, tag)
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
func (s *RedisClusterStore) Invalidate(ctx context.Context, options ...lib_store.InvalidateOption) error {
	opts := lib_store.ApplyInvalidateOptions(options...)

	if tags := opts.Tags; len(tags) > 0 {
		for _, tag := range tags {
			tagKey := fmt.Sprintf(RedisClusterTagPattern, tag)
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
