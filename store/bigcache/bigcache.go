package bigcache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/eko/gocache/lib/v4/store"
)

// BigcacheClientInterface represents a allegro/bigcache client
type BigcacheClientInterface interface {
	Get(key string) ([]byte, error)
	Set(key string, entry []byte) error
	Delete(key string) error
	Reset() error
}

const (
	// BigcacheType represents the storage type as a string value
	BigcacheType = "bigcache"
	// BigcacheTagPattern represents the tag pattern to be used as a key in specified storage
	BigcacheTagPattern = "gocache_tag_%s"
)

// BigcacheStore is a store for Bigcache
type BigcacheStore struct {
	mu      sync.Mutex
	client  BigcacheClientInterface
	options *store.Options
}

// NewBigcache creates a new store to Bigcache instance(s)
func NewBigcache(client BigcacheClientInterface, options ...store.Option) *BigcacheStore {
	return &BigcacheStore{
		client:  client,
		options: store.ApplyOptions(options...),
	}
}

// Get returns data stored from a given key
func (s *BigcacheStore) Get(_ context.Context, key any) (any, error) {
	item, err := s.client.Get(key.(string))
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			return nil, store.NotFoundWithCause(err)
		}

		return nil, err
	}
	if item == nil {
		return nil, store.NotFoundWithCause(errors.New("unable to retrieve data from bigcache"))
	}

	return item, nil
}

// Even though Bigcache does not support a TTL, try our best to implement this
// because it's needed by ChainCache -- if we just return an error, ChainCache
// will never see any cache hits. Arbitrarily pick a 5 minute timeout since it's
// not "too big" or "too small".
func (s *BigcacheStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	result, err := s.Get(ctx, key)
	return result, time.Minute * 5, err
}

// Set defines data in Bigcache for given key identifier
func (s *BigcacheStore) Set(ctx context.Context, key any, value any, options ...store.Option) error {
	opts := store.ApplyOptionsWithDefault(s.options, options...)

	var val []byte
	switch v := value.(type) {
	case string:
		val = []byte(v)
	case []byte:
		val = v
	default:
		return errors.New("value type not supported by Bigcache store")
	}

	err := s.client.Set(key.(string), val)
	if err != nil {
		return err
	}

	if tags := opts.Tags; len(tags) > 0 {
		s.setTags(ctx, key, tags)
	}

	return nil
}

func (s *BigcacheStore) setTags(ctx context.Context, key any, tags []string) {
	for _, tag := range tags {
		tagKey := fmt.Sprintf(BigcacheTagPattern, tag)

		// The whole read-modify-write sequence has to happen under the same lock:
		// otherwise two concurrent Set calls on the same tag both read the same
		// list of keys and the last write wins, losing the other one key.
		s.mu.Lock()

		cacheKeys := s.getCacheKeysForTag(ctx, tagKey)

		alreadyInserted := false
		for _, cacheKey := range cacheKeys {
			if cacheKey == key.(string) {
				alreadyInserted = true
				break
			}
		}

		if !alreadyInserted {
			cacheKeys = append(cacheKeys, key.(string))
		}

		s.client.Set(tagKey, []byte(strings.Join(cacheKeys, ",")))

		s.mu.Unlock()
	}
}

// getCacheKeysForTag returns the list of cache keys associated to a given tag key
func (s *BigcacheStore) getCacheKeysForTag(ctx context.Context, tagKey string) []string {
	result, err := s.Get(ctx, tagKey)
	if err != nil {
		return []string{}
	}

	bytes, ok := result.([]byte)
	if !ok || len(bytes) == 0 {
		return []string{}
	}

	return strings.Split(string(bytes), ",")
}

// Delete removes data from Bigcache for given key identifier
func (s *BigcacheStore) Delete(_ context.Context, key any) error {
	return s.client.Delete(key.(string))
}

// Invalidate invalidates some cache data in Bigcache for given options
func (s *BigcacheStore) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	opts := store.ApplyInvalidateOptions(options...)

	for _, tag := range opts.Tags {
		tagKey := fmt.Sprintf(BigcacheTagPattern, tag)

		s.mu.Lock()
		cacheKeys := s.getCacheKeysForTag(ctx, tagKey)
		s.mu.Unlock()

		for _, cacheKey := range cacheKeys {
			_ = s.Delete(ctx, cacheKey)
		}
	}

	return nil
}

// Clear resets all data in the store
func (s *BigcacheStore) Clear(_ context.Context) error {
	return s.client.Reset()
}

// GetType returns the store type
func (s *BigcacheStore) GetType() string {
	return BigcacheType
}
