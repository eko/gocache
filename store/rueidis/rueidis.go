package rueidis

import (
	"context"
	"fmt"
	"time"

	lib_store "github.com/eko/gocache/v4/lib/store"
	"github.com/rueian/rueidis"
	"github.com/rueian/rueidis/rueidiscompat"
)

const (
	// RueidisType represents the storage type as a string value
	RueidisType = "rueidis"
	// RueidisTagPattern represents the tag pattern to be used as a key in specified storage
	RueidisTagPattern = "gocache_tag_%s"

	defaultClientSideCacheExpiration = 10 * time.Second
)

// RueidisStore is a store for Redis
type RueidisStore struct {
	client      rueidis.Client
	options     *lib_store.Options
	cacheCompat rueidiscompat.CacheCompat
	compat      rueidiscompat.Cmdable
}

// NewRueidis creates a new store to Redis instance(s)
func NewRueidis(client rueidis.Client, options ...lib_store.Option) *RueidisStore {
	// defaults client side cache expiration to 10s
	appliedOptions := lib_store.ApplyOptions(options...)

	if appliedOptions.ClientSideCacheExpiration == 0 {
		appliedOptions.ClientSideCacheExpiration = defaultClientSideCacheExpiration
	}

	return &RueidisStore{
		client:      client,
		cacheCompat: rueidiscompat.NewAdapter(client).Cache(appliedOptions.ClientSideCacheExpiration),
		compat:      rueidiscompat.NewAdapter(client),
		options:     appliedOptions,
	}
}

// Get returns data stored from a given key
func (s *RueidisStore) Get(ctx context.Context, key any) (any, error) {
	object := s.client.DoCache(ctx, s.client.B().Get().Key(key.(string)).Cache(), s.options.ClientSideCacheExpiration)
	if object.RedisError() != nil && object.RedisError().IsNil() {
		return nil, lib_store.NotFoundWithCause(object.Error())
	}
	return object, object.Error()
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *RueidisStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	// get object first
	object, err := s.Get(ctx, key)
	if err != nil {
		return nil, 0, err
	}

	// get TTL and return
	ttl, err := s.cacheCompat.TTL(ctx, key.(string)).Result()
	if err != nil {
		return nil, 0, err
	}

	return object, ttl, err
}

// Set defines data in Redis for given key identifier
func (s *RueidisStore) Set(ctx context.Context, key any, value any, options ...lib_store.Option) error {
	opts := lib_store.ApplyOptionsWithDefault(s.options, options...)
	err := s.compat.Set(ctx, key.(string), value, opts.Expiration).Err()
	if err != nil {
		return err
	}

	if tags := opts.Tags; len(tags) > 0 {
		s.setTags(ctx, key, tags)
	}

	return nil
}

func (s *RueidisStore) setTags(ctx context.Context, key any, tags []string) {
	for _, tag := range tags {
		tagKey := fmt.Sprintf(RueidisTagPattern, tag)
		s.compat.SAdd(ctx, tagKey, key.(string))
		s.compat.Expire(ctx, tagKey, 720*time.Hour)
	}
}

// Delete removes data from Redis for given key identifier
func (s *RueidisStore) Delete(ctx context.Context, key any) error {
	_, err := s.compat.Del(ctx, key.(string)).Result()
	return err
}

// Invalidate invalidates some cache data in Redis for given options
func (s *RueidisStore) Invalidate(ctx context.Context, options ...lib_store.InvalidateOption) error {
	opts := lib_store.ApplyInvalidateOptions(options...)

	if tags := opts.Tags; len(tags) > 0 {
		for _, tag := range tags {
			tagKey := fmt.Sprintf(RueidisTagPattern, tag)

			cacheKeys, err := s.cacheCompat.SMembers(ctx, tagKey).Result()
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

// GetType returns the store type
func (s *RueidisStore) GetType() string {
	return RueidisType
}

// Clear resets all data in the store
func (s *RueidisStore) Clear(ctx context.Context) error {
	if err := s.compat.FlushAll(ctx).Err(); err != nil {
		return err
	}

	return nil
}
