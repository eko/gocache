package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

// MemcacheClientInterface represents a bradfitz/gomemcache client
type MemcacheClientInterface interface {
	Get(key string) (item *memcache.Item, err error)
	Set(item *memcache.Item) error
	Delete(item string) error
	FlushAll() error
}

const (
	// MemcacheType represents the storage type as a string value
	MemcacheType = "memcache"
	// MemcacheTagPattern represents the tag pattern to be used as a key in specified storage
	MemcacheTagPattern = "gocache_tag_%s"
)

// MemcacheStore is a store for Memcache
type MemcacheStore struct {
	client  MemcacheClientInterface
	options *Options
}

// NewMemcache creates a new store to Memcache instance(s)
func NewMemcache(client MemcacheClientInterface, options *Options) *MemcacheStore {
	if options == nil {
		options = &Options{}
	}

	return &MemcacheStore{
		client:  client,
		options: options,
	}
}

// Get returns data stored from a given key
func (s *MemcacheStore) Get(_ context.Context, key interface{}) (interface{}, error) {
	item, err := s.client.Get(key.(string))
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("Unable to retrieve data from memcache")
	}

	return item.Value, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *MemcacheStore) GetWithTTL(_ context.Context, key interface{}) (interface{}, time.Duration, error) {
	item, err := s.client.Get(key.(string))
	if err != nil {
		return nil, 0, err
	}
	if item == nil {
		return nil, 0, errors.New("Unable to retrieve data from memcache")
	}

	return item.Value, time.Duration(item.Expiration) * time.Second, err
}

// Set defines data in Memcache for given key identifier
func (s *MemcacheStore) Set(ctx context.Context, key interface{}, value interface{}, options *Options) error {
	if options == nil {
		options = s.options
	}

	item := &memcache.Item{
		Key:        key.(string),
		Value:      value.([]byte),
		Expiration: int32(options.ExpirationValue().Seconds()),
	}

	err := s.client.Set(item)
	if err != nil {
		return err
	}

	if tags := options.TagsValue(); len(tags) > 0 {
		s.setTags(ctx, key, tags)
	}

	return nil
}

func (s *MemcacheStore) setTags(ctx context.Context, key interface{}, tags []string) {
	for _, tag := range tags {
		var tagKey = fmt.Sprintf(MemcacheTagPattern, tag)
		var cacheKeys = []string{}

		if result, err := s.Get(ctx, tagKey); err == nil {
			if bytes, ok := result.([]byte); ok {
				cacheKeys = strings.Split(string(bytes), ",")
			}
		}

		var alreadyInserted = false
		for _, cacheKey := range cacheKeys {
			if cacheKey == key.(string) {
				alreadyInserted = true
				break
			}
		}

		if !alreadyInserted {
			cacheKeys = append(cacheKeys, key.(string))
		}

		s.Set(ctx, tagKey, []byte(strings.Join(cacheKeys, ",")), &Options{
			Expiration: 720 * time.Hour,
		})
	}
}

// Delete removes data from Memcache for given key identifier
func (s *MemcacheStore) Delete(_ context.Context, key interface{}) error {
	return s.client.Delete(key.(string))
}

// Invalidate invalidates some cache data in Redis for given options
func (s *MemcacheStore) Invalidate(ctx context.Context, options InvalidateOptions) error {
	if tags := options.TagsValue(); len(tags) > 0 {
		for _, tag := range tags {
			var tagKey = fmt.Sprintf(MemcacheTagPattern, tag)
			result, err := s.Get(ctx, tagKey)
			if err != nil {
				return nil
			}

			var cacheKeys = []string{}
			if bytes, ok := result.([]byte); ok {
				cacheKeys = strings.Split(string(bytes), ",")
			}

			for _, cacheKey := range cacheKeys {
				s.Delete(ctx, cacheKey)
			}
		}
	}

	return nil
}

// Clear resets all data in the store
func (s *MemcacheStore) Clear(_ context.Context) error {
	return s.client.FlushAll()
}

// GetType returns the store type
func (s *MemcacheStore) GetType() string {
	return MemcacheType
}
