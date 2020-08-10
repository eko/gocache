package store

import (
	"strings"
	"time"

	"github.com/yeqown/gocache"

	"github.com/yeqown/gocache/types"

	redis "github.com/go-redis/redis/v7"
)

var (
	_defaultTagStoreOption = &types.StoreOptions{
		Expiration: 720 * time.Hour,
	}
)

// redisClientInterface represents a go-redis/redis client
// [refactor delay] TODO(@yeqown): implement `tag` function in redis with `Set` structure rather than `String`
type redisClientInterface interface {
	Get(key string) *redis.StringCmd
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(keys ...string) *redis.IntCmd
	FlushAll() *redis.StatusCmd
}

const (
	// _redisType represents the storage type as a string value
	_redisType = "redis"
)

// RedisStore is a store for Redis
type RedisStore struct {
	client redisClientInterface

	storeOpt *types.StoreOptions
}

// NewRedis creates a new store to Redis instance(s)
func NewRedis(client redisClientInterface, options *types.StoreOptions) gocache.IStore {
	if options == nil {
		options = &types.StoreOptions{}
	}

	return &RedisStore{
		client:   client,
		storeOpt: options,
	}
}

// Get returns data stored from a given key
func (s *RedisStore) Get(key string) ([]byte, error) {
	return s.client.Get(key).Bytes()
}

// Set defines data in Redis for given key identifier
func (s *RedisStore) Set(key string, value interface{}, options *types.StoreOptions) error {
	if options == nil {
		options = s.storeOpt
	}

	err := s.client.Set(key, value, options.ExpirationValue()).Err()
	if err != nil {
		return err
	}

	if tags := options.TagsValue(); len(tags) > 0 {
		return s.setTags(key, tags)
	}

	return nil
}

func (s *RedisStore) setTags(key string, tags []string) error {
	var multiErr = new(types.MultiError)

	for _, tag := range tags {
		tagKey := gocache.GenTagKey(tag)
		cacheKeys := s.getCacheKeysForTag(tagKey)
		exists := false

		for _, cacheKey := range cacheKeys {
			if cacheKey == key {
				exists = true
				break
			}
		}

		if !exists {
			// 不存在则更新
			cacheKeys = append(cacheKeys, key)
		}

		// FIXME: how to calculate tags expired timestamp
		tagValue := strings.Join(cacheKeys, ",")
		if err := s.Set(tagKey, tagValue, _defaultTagStoreOption); err != nil {
			multiErr.Add(err)
		}
	}

	if !multiErr.Nil() {
		return multiErr
	}

	return nil
}

func (s *RedisStore) getCacheKeysForTag(tagKey string) (cacheKeys []string) {
	if result, err := s.Get(tagKey); err == nil && result != nil {
		cacheKeys = strings.Split(string(result), ",")
	}

	return cacheKeys
}

// Invalidate invalidates some cache data in Redis for given options
func (s *RedisStore) Invalidate(options types.InvalidateOptions) error {
	var multiErr = new(types.MultiError)

	if tags := options.TagsValue(); len(tags) > 0 {
		for _, tag := range tags {
			tagKey := gocache.GenTagKey(tag)
			cacheKeys := s.getCacheKeysForTag(tagKey)

			for _, cacheKey := range cacheKeys {
				if err := s.Delete(cacheKey); err != nil {
					multiErr.Add(err)
				}
			}

			if err := s.Delete(tagKey); err != nil {
				multiErr.Add(err)
			}
		}
	}

	if !multiErr.Nil() {
		return multiErr
	}

	return nil
}

// FIXME: could not using flush, this would del all data in current DB
func (s *RedisStore) Clear() error    { return s.client.FlushAll().Err() }
func (s *RedisStore) GetType() string { return _redisType }
func (s *RedisStore) Delete(key string) error {
	_, err := s.client.Del(key).Result()
	return err
}
