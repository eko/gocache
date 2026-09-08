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

	// TagKeyExpiry is the default expiration applied to tag keys
	TagKeyExpiry = 720 * time.Hour
)

// GoCacheClientInterface represents a github.com/patrickmn/go-cache client
type GoCacheClientInterface interface {
	Get(k string) (any, bool)
	GetWithExpiration(k string) (any, time.Time, bool)
	Set(k string, x any, d time.Duration)
	Add(k string, x any, d time.Duration) error
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
	opts := lib_store.ApplyOptionsWithDefault(s.options, options...)

	s.client.Set(key.(string), value, opts.Expiration)

	if tags := opts.Tags; len(tags) > 0 {
		ttl := opts.TagsTTL
		if ttl == 0 {
			ttl = TagKeyExpiry
		}
		s.setTags(key, tags, ttl)
	}

	return nil
}

// SetIfNotExists defines data in GoCache memory cache for given key identifier
// only when it does not already exist, and reports whether it has been written
func (s *GoCacheStore) SetIfNotExists(_ context.Context, key any, value any, options ...lib_store.Option) (bool, error) {
	opts := lib_store.ApplyOptionsWithDefault(s.options, options...)

	// go-cache only fails Add when the key is already there and not expired
	if err := s.client.Add(key.(string), value, opts.Expiration); err != nil {
		return false, nil
	}

	if tags := opts.Tags; len(tags) > 0 {
		ttl := opts.TagsTTL
		if ttl == 0 {
			ttl = TagKeyExpiry
		}
		s.setTags(key, tags, ttl)
	}

	return true, nil
}

func (s *GoCacheStore) setTags(key any, tags []string, ttl time.Duration) {
	for _, tag := range tags {
		tagKey := fmt.Sprintf(GoCacheTagPattern, tag)

		// The whole read-modify-write sequence has to happen under the same lock:
		// otherwise two concurrent Set calls on the same tag both read the same
		// list of keys and the last write wins, losing the other one key.
		s.mu.Lock()

		var currentKeys map[string]struct{}
		if result, exists := s.client.Get(tagKey); exists {
			currentKeys, _ = result.(map[string]struct{})
		}

		if _, alreadyInserted := currentKeys[key.(string)]; alreadyInserted {
			s.mu.Unlock()
			continue
		}

		// Store a copy: the previous map may still be read by another goroutine.
		cacheKeys := make(map[string]struct{}, len(currentKeys)+1)
		for cacheKey := range currentKeys {
			cacheKeys[cacheKey] = struct{}{}
		}
		cacheKeys[key.(string)] = struct{}{}

		s.client.Set(tagKey, cacheKeys, ttl)

		s.mu.Unlock()
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

	for _, tag := range opts.Tags {
		tagKey := fmt.Sprintf(GoCacheTagPattern, tag)

		s.mu.RLock()
		result, exists := s.client.Get(tagKey)
		s.mu.RUnlock()

		if !exists {
			continue
		}

		cacheKeys, ok := result.(map[string]struct{})
		if !ok {
			continue
		}

		for cacheKey := range cacheKeys {
			_ = s.Delete(ctx, cacheKey)
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
