package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	// RistrettoType represents the storage type as a string value
	RistrettoType = "ristretto"
	// RistrettoTagPattern represents the tag pattern to be used as a key in specified storage
	RistrettoTagPattern = "gocache_tag_%s"
)

// RistrettoClientInterface represents a dgraph-io/ristretto client
type RistrettoClientInterface interface {
	Get(key interface{}) (interface{}, bool)
	SetWithTTL(key, value interface{}, cost int64, ttl time.Duration) bool
	Del(key interface{})
	Clear()
}

// RistrettoStore is a store for Ristretto (memory) library
type RistrettoStore struct {
	client  RistrettoClientInterface
	options *Options
}

// NewRistretto creates a new store to Ristretto (memory) library instance
func NewRistretto(client RistrettoClientInterface, options *Options) *RistrettoStore {
	if options == nil {
		options = &Options{}
	}

	return &RistrettoStore{
		client:  client,
		options: options,
	}
}

// Get returns data stored from a given key
func (s *RistrettoStore) Get(_ context.Context, key interface{}) (interface{}, error) {
	var err error

	value, exists := s.client.Get(key)
	if !exists {
		err = errors.New("Value not found in Ristretto store")
	}

	return value, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *RistrettoStore) GetWithTTL(ctx context.Context, key interface{}) (interface{}, time.Duration, error) {
	value, err := s.Get(ctx, key)
	return value, 0, err
}

// Set defines data in Ristretto memoey cache for given key identifier
func (s *RistrettoStore) Set(ctx context.Context, key interface{}, value interface{}, options *Options) error {
	var err error

	if options == nil {
		options = s.options
	}

	if set := s.client.SetWithTTL(key, value, options.CostValue(), options.ExpirationValue()); !set {
		err = fmt.Errorf("An error has occurred while setting value '%v' on key '%v'", value, key)
	}

	if err != nil {
		return err
	}

	if tags := options.TagsValue(); len(tags) > 0 {
		s.setTags(ctx, key, tags)
	}

	return nil
}

func (s *RistrettoStore) setTags(ctx context.Context, key interface{}, tags []string) {
	for _, tag := range tags {
		var tagKey = fmt.Sprintf(RistrettoTagPattern, tag)
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

// Delete removes data in Ristretto memoey cache for given key identifier
func (s *RistrettoStore) Delete(_ context.Context, key interface{}) error {
	s.client.Del(key)
	return nil
}

// Invalidate invalidates some cache data in Redis for given options
func (s *RistrettoStore) Invalidate(ctx context.Context, options InvalidateOptions) error {
	if tags := options.TagsValue(); len(tags) > 0 {
		for _, tag := range tags {
			var tagKey = fmt.Sprintf(RistrettoTagPattern, tag)
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
func (s *RistrettoStore) Clear(_ context.Context) error {
	s.client.Clear()
	return nil
}

// GetType returns the store type
func (s *RistrettoStore) GetType() string {
	return RistrettoType
}
