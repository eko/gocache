package store

import (
	"context"
	"fmt"
	"time"

	"github.com/rueian/rueidis"
	"github.com/rueian/rueidis/rueidiscompat"
)

const (
	// RueidisType represents the storage type as a string value
	RueidisType = "rueidis"
	// RueidisTagPattern represents the tag pattern to be used as a key in specified storage
	RueidisTagPattern = "gocache_tag_%s"
)

// RueidisStore is a store for Redis
type RueidisStore struct {
	client           rueidis.Client
	options          *Options
	clientExpiration time.Duration
	cacheCompat      rueidiscompat.CacheCompat
	compat           rueidiscompat.Cmdable
}

// NewRueidis creates a new store to Redis instance(s)
func NewRueidis(client rueidis.Client, clientExpiration *time.Duration, options ...Option) *RueidisStore {
	// defaults client expiration to 10s
	expiration := time.Second * 10
	if clientExpiration != nil {
		expiration = *clientExpiration
	}
	return &RueidisStore{
		client:           client,
		cacheCompat:      rueidiscompat.NewAdapter(client).Cache(time.Second * 10),
		compat:           rueidiscompat.NewAdapter(client),
		options:          applyOptions(options...),
		clientExpiration: expiration,
	}
}

// Get returns data stored from a given key
func (s *RueidisStore) Get(ctx context.Context, key any) (any, error) {
	object := s.client.DoCache(ctx, s.client.B().Get().Key(key.(string)).Cache(), s.clientExpiration)
	if object.RedisError() != nil && object.RedisError().IsNil() {
		return nil, NotFoundWithCause(object.Error())
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
func (s *RueidisStore) Set(ctx context.Context, key any, value any, options ...Option) error {
	opts := applyOptionsWithDefault(s.options, options...)
	err := s.compat.Set(ctx, key.(string), value, opts.expiration).Err()
	if err != nil {
		return err
	}

	if tags := opts.tags; len(tags) > 0 {
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
func (s *RueidisStore) Invalidate(ctx context.Context, options ...InvalidateOption) error {
	opts := applyInvalidateOptions(options...)

	if tags := opts.tags; len(tags) > 0 {
		for _, tag := range tags {
			tagKey := fmt.Sprintf(RedisTagPattern, tag)
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
