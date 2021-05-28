package store

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	// GoCacheType represents the storage type as a string value
	GoCacheType = "go-cache"
	// GoCacheTagPattern represents the tag pattern to be used as a key in specified storage
	GoCacheTagPattern = "gocache_tag_%s"
)

// GoCacheClientInterface represents a github.com/patrickmn/go-cache client
type GoCacheClientInterface interface {
	Get(k string) (interface{}, bool)
	GetWithExpiration(k string) (interface{}, time.Time, bool)
	Set(k string, x interface{}, d time.Duration)
	Delete(k string)
	Flush()
}

// GoCacheStore is a store for GoCache (memory) library
type GoCacheStore struct {
	client  GoCacheClientInterface
	options *Options
}

// NewGoCache creates a new store to GoCache (memory) library instance
func NewGoCache(client GoCacheClientInterface, options *Options) *GoCacheStore {
	if options == nil {
		options = &Options{}
	}

	return &GoCacheStore{
		client:  client,
		options: options,
	}
}

// Get returns data stored from a given key
func (s *GoCacheStore) Get(_ context.Context, key interface{}) (interface{}, error) {
	var err error
	keyStr := key.(string)
	value, exists := s.client.Get(keyStr)
	if !exists {
		err = errors.New("Value not found in GoCache store")
	}

	return value, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *GoCacheStore) GetWithTTL(_ context.Context, key interface{}) (interface{}, time.Duration, error) {
	data, t, exists := s.client.GetWithExpiration(key.(string))
	if !exists {
		return data, 0, errors.New("Value not found in GoCache store")
	}
	duration := t.Sub(time.Now())
	return data, duration, nil
}

// Set defines data in GoCache memoey cache for given key identifier
func (s *GoCacheStore) Set(ctx context.Context, key interface{}, value interface{}, options *Options) error {

	if options == nil {
		options = s.options
	}

	s.client.Set(key.(string), value, options.ExpirationValue())

	if tags := options.TagsValue(); len(tags) > 0 {
		s.setTags(ctx, key, tags)
	}

	return nil
}

func (s *GoCacheStore) setTags(ctx context.Context, key interface{}, tags []string) {
	for _, tag := range tags {
		var tagKey = fmt.Sprintf(GoCacheTagPattern, tag)
		var cacheKeys map[string]struct{}

		if result, err := s.Get(ctx, tagKey); err == nil {
			if bytes, ok := result.(map[string]struct{}); ok {
				cacheKeys = bytes
			}
		}
		if _, exists := cacheKeys[key.(string)]; exists {
			continue
		}

		if cacheKeys == nil {
			cacheKeys = make(map[string]struct{})
		}

		cacheKeys[key.(string)] = struct{}{}

		s.client.Set(tagKey, cacheKeys, 720*time.Hour)
	}
}

// Delete removes data in GoCache memoey cache for given key identifier
func (s *GoCacheStore) Delete(_ context.Context, key interface{}) error {
	s.client.Delete(key.(string))
	return nil
}

// Invalidate invalidates some cache data in GoCache memoey cache for given options
func (s *GoCacheStore) Invalidate(ctx context.Context, options InvalidateOptions) error {
	if tags := options.TagsValue(); len(tags) > 0 {
		for _, tag := range tags {
			var tagKey = fmt.Sprintf(GoCacheTagPattern, tag)
			result, err := s.Get(ctx, tagKey)
			if err != nil {
				return nil
			}

			var cacheKeys map[string]struct{}
			if bytes, ok := result.(map[string]struct{}); ok {
				cacheKeys = bytes
			}

			for cacheKey := range cacheKeys {
				_ = s.Delete(ctx, cacheKey)
			}
		}
	}

	return nil
}

// GetType returns the store type
func (s *GoCacheStore) GetType() string {
	return GoCacheType
}

// Clear resets all data in the store
func (s *GoCacheStore) Clear(_ context.Context) error {
	s.client.Flush()
	return nil
}
