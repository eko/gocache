package go_cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
)

const (
	// GoCacheType represents the storage type as a string value
	GoCacheType = "go-cache"
	// GoCacheTagPattern represents the tag pattern to be used as a key in specified storage
	GoCacheTagPattern = "gocache_tag_%s"
)

// GoCacheClientInterface represents a github.com/patrickmn/go-cache client
type GoCacheClientInterface interface {
	Get(k string) (any, bool)
	GetWithExpiration(k string) (any, time.Time, bool)
	Set(k string, x any, d time.Duration)
	Delete(k string)
	Flush()
}

// GoCacheStore is a store for GoCache (memory) library
type GoCacheStore struct {
	mu      sync.RWMutex
	client  GoCacheClientInterface
	options *lib_store.Options
}

// NewGoCache creates a new store to GoCache (memory) library instance
func NewGoCache(client GoCacheClientInterface, options ...lib_store.Option) *GoCacheStore {
	return &GoCacheStore{
		client:  client,
		options: lib_store.ApplyOptions(options...),
	}
}

// Get returns data stored from a given key
func (s *GoCacheStore) Get(_ context.Context, key any) (any, error) {
	var err error
	keyStr := key.(string)
	value, exists := s.client.Get(keyStr)
	if !exists {
		err = lib_store.NotFoundWithCause(errors.New("value not found in GoCache store"))
	}

	return value, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *GoCacheStore) GetWithTTL(_ context.Context, key any) (any, time.Duration, error) {
	data, t, exists := s.client.GetWithExpiration(key.(string))
	if !exists {
		return data, 0, lib_store.NotFoundWithCause(errors.New("value not found in GoCache store"))
	}
	duration := time.Until(t)
	return data, duration, nil
}

// Set defines data in GoCache memoey cache for given key identifier
func (s *GoCacheStore) Set(ctx context.Context, key any, value any, options ...lib_store.Option) error {
	opts := lib_store.ApplyOptions(options...)
	if opts == nil {
		opts = s.options
	}

	s.client.Set(key.(string), value, opts.Expiration)

	if tags := opts.Tags; len(tags) > 0 {
		s.setTags(ctx, key, tags)
	}

	return nil
}

func (s *GoCacheStore) setTags(ctx context.Context, key any, tags []string) {
	for _, tag := range tags {
		tagKey := fmt.Sprintf(GoCacheTagPattern, tag)
		var cacheKeys map[string]struct{}

		if result, err := s.Get(ctx, tagKey); err == nil {
			if bytes, ok := result.(map[string]struct{}); ok {
				cacheKeys = bytes
			}
		}

		s.mu.RLock()
		if _, exists := cacheKeys[key.(string)]; exists {
			s.mu.RUnlock()
			continue
		}
		s.mu.RUnlock()

		if cacheKeys == nil {
			cacheKeys = make(map[string]struct{})
		}

		s.mu.Lock()
		cacheKeys[key.(string)] = struct{}{}
		s.mu.Unlock()

		s.client.Set(tagKey, cacheKeys, 720*time.Hour)
	}
}

// Delete removes data in GoCache memoey cache for given key identifier
func (s *GoCacheStore) Delete(_ context.Context, key any) error {
	s.client.Delete(key.(string))
	return nil
}

// Invalidate invalidates some cache data in GoCache memoey cache for given options
func (s *GoCacheStore) Invalidate(ctx context.Context, options ...lib_store.InvalidateOption) error {
	opts := lib_store.ApplyInvalidateOptions(options...)

	if tags := opts.Tags; len(tags) > 0 {
		for _, tag := range tags {
			tagKey := fmt.Sprintf(GoCacheTagPattern, tag)
			result, err := s.Get(ctx, tagKey)
			if err != nil {
				return nil
			}

			var cacheKeys map[string]struct{}
			if bytes, ok := result.(map[string]struct{}); ok {
				cacheKeys = bytes
			}

			s.mu.RLock()
			for cacheKey := range cacheKeys {
				_ = s.Delete(ctx, cacheKey)
			}
			s.mu.RUnlock()
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
