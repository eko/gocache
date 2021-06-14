package store

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
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
	CompareAndSwap(item *memcache.Item) error
	Add(item *memcache.Item) error
}

const (
	// MemcacheType represents the storage type as a string value
	MemcacheType = "memcache"
	// MemcacheTagPattern represents the tag pattern to be used as a key in specified storage
	MemcacheTagPattern = "gocache_tag_%s"

	TagKeyExpiry = 720 * time.Hour
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
	group, ctx := errgroup.WithContext(ctx)
	for _, tag := range tags {
		currentTag := tag
		group.Go(func() error {
			var tagKey = fmt.Sprintf(MemcacheTagPattern, currentTag)

			var err error
			for i := 0; i < 3; i++ {
				if err = s.addKeyToTagValue(tagKey, key); err == nil {
					return nil
				}
				// loop to retry any failure (including race conditions)
			}

			return err
		})
	}

	group.Wait()
}

func (s *MemcacheStore) addKeyToTagValue(tagKey string, key interface{}) error {
	var (
		cacheKeys = []string{}
		result    *memcache.Item
		err       error
	)

	result, err = s.client.Get(tagKey)
	if err == nil {
		cacheKeys = strings.Split(string(result.Value), ",")
	} else if !errors.Is(err, memcache.ErrCacheMiss) {
		return err
	}

	for _, cacheKey := range cacheKeys {
		// if key already exists, nothing to do
		if cacheKey == key.(string) {
			return nil
		}
	}

	cacheKeys = append(cacheKeys, key.(string))

	newVal := []byte(strings.Join(cacheKeys, ","))

	if result == nil {
		// if key didnt exist, use Add to create only if still not there
		return s.client.Add(&memcache.Item{
			Key:        tagKey,
			Value:      newVal,
			Expiration: int32(TagKeyExpiry.Seconds()),
		})
	}

	// update existing value
	// using CompareAndSwap to ensure not to run over writes between Get and here
	result.Value = newVal
	result.Expiration = int32(TagKeyExpiry.Seconds())
	return s.client.CompareAndSwap(result)
}

// Delete removes data from Memcache for given key identifier
func (s *MemcacheStore) Delete(_ context.Context, key interface{}) error {
	return s.client.Delete(key.(string))
}

// Invalidate invalidates some cache data in Memcache for given options
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
