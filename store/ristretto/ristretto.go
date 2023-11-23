package ristretto

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
)

const (
	// RistrettoType represents the storage type as a string value
	RistrettoType = "ristretto"
	// RistrettoTagPattern represents the tag pattern to be used as a key in specified storage
	RistrettoTagPattern = "gocache_tag_%s"
)

// RistrettoClientInterface represents a dgraph-io/ristretto client
type RistrettoClientInterface interface {
	Get(key any) (any, bool)
	SetWithTTL(key, value any, cost int64, ttl time.Duration) bool
	Del(key any)
	Clear()
	Wait()
}

// RistrettoStore is a store for Ristretto (memory) library
type RistrettoStore struct {
	client  RistrettoClientInterface
	options *lib_store.Options
}

// NewRistretto creates a new store to Ristretto (memory) library instance
func NewRistretto(client RistrettoClientInterface, options ...lib_store.Option) *RistrettoStore {
	return &RistrettoStore{
		client:  client,
		options: lib_store.ApplyOptions(options...),
	}
}

// Get returns data stored from a given key
func (s *RistrettoStore) Get(_ context.Context, key any) (any, error) {
	var err error

	value, exists := s.client.Get(key)
	if !exists {
		err = lib_store.NotFoundWithCause(errors.New("value not found in Ristretto store"))
	}

	return value, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *RistrettoStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	value, err := s.Get(ctx, key)
	return value, 0, err
}

// Set defines data in Ristretto memory cache for given key identifier
func (s *RistrettoStore) Set(ctx context.Context, key any, value any, options ...lib_store.Option) error {
	opts := lib_store.ApplyOptionsWithDefault(s.options, options...)

	var err error

	if set := s.client.SetWithTTL(key, value, opts.Cost, opts.Expiration); !set {
		err = fmt.Errorf("An error has occurred while setting value '%v' on key '%v'", value, key)
	}

	if err != nil {
		return err
	}

	if opts.SynchronousSet {
		s.client.Wait()
	}

	if tags := opts.Tags; len(tags) > 0 {
		s.setTags(ctx, key, tags)
	}

	return nil
}

func (s *RistrettoStore) setTags(ctx context.Context, key any, tags []string) {
	for _, tag := range tags {
		tagKey := fmt.Sprintf(RistrettoTagPattern, tag)
		cacheKeys := []string{}

		if result, err := s.Get(ctx, tagKey); err == nil {
			if bytes, ok := result.([]byte); ok {
				cacheKeys = strings.Split(string(bytes), ",")
			}
		}

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

		s.Set(ctx, tagKey, []byte(strings.Join(cacheKeys, ",")), lib_store.WithExpiration(720*time.Hour))
	}
}

// Delete removes data in Ristretto memory cache for given key identifier
func (s *RistrettoStore) Delete(_ context.Context, key any) error {
	s.client.Del(key)
	return nil
}

// Invalidate invalidates some cache data in Redis for given options
func (s *RistrettoStore) Invalidate(ctx context.Context, options ...lib_store.InvalidateOption) error {
	opts := lib_store.ApplyInvalidateOptions(options...)

	if tags := opts.Tags; len(tags) > 0 {
		for _, tag := range tags {
			tagKey := fmt.Sprintf(RistrettoTagPattern, tag)
			result, err := s.Get(ctx, tagKey)
			if err != nil {
				return nil
			}

			cacheKeys := []string{}
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
