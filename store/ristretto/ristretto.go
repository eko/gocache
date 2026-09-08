package ristretto

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	lib_store "github.com/eko/gocache/lib/v4/store"
)

const (
	// RistrettoType represents the storage type as a string value
	RistrettoType = "ristretto"
	// RistrettoTagPattern represents the tag pattern to be used as a key in specified storage
	RistrettoTagPattern = "gocache_tag_%s"
)

// RistrettoClientInterface represents a dgraph-io/ristretto client
type RistrettoClientInterface[K ristretto.Key, V any] interface {
	Get(key K) (V, bool)
	GetTTL(key K) (time.Duration, bool)
	SetWithTTL(key K, value V, cost int64, ttl time.Duration) bool
	Del(key K)
	Clear()
	Wait()
}

// RistrettoStore is a store for Ristretto (memory) library
type RistrettoStore[K ristretto.Key, V any] struct {
	mu      sync.Mutex
	client  RistrettoClientInterface[K, V]
	options *lib_store.Options
}

// NewRistretto creates a new store to Ristretto (memory) library instance
func NewRistretto[K ristretto.Key, V any](
	client RistrettoClientInterface[K, V],
	options ...lib_store.Option,
) *RistrettoStore[K, V] {
	return &RistrettoStore[K, V]{
		client:  client,
		options: lib_store.ApplyOptions(options...),
	}
}

// Get returns data stored from a given key
func (s *RistrettoStore[K, V]) Get(_ context.Context, key any) (any, error) {
	k, ok := key.(K)
	if !ok {
		return nil, unsupportedTypeError[K]("key", key)
	}

	var err error

	value, exists := s.client.Get(k)
	if !exists {
		err = lib_store.NotFoundWithCause(errors.New("value not found in Ristretto store"))
	}

	return value, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *RistrettoStore[K, V]) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return value, 0, err
	}

	ttl, _ := s.client.GetTTL(key.(K))

	return value, ttl, nil
}

// Set defines data in Ristretto memory cache for given key identifier
func (s *RistrettoStore[K, V]) Set(ctx context.Context, key any, value any, options ...lib_store.Option) error {
	opts := lib_store.ApplyOptionsWithDefault(s.options, options...)

	k, ok := key.(K)
	if !ok {
		return unsupportedTypeError[K]("key", key)
	}

	v, ok := value.(V)
	if !ok && value != nil {
		return unsupportedTypeError[V]("value", value)
	}

	if set := s.client.SetWithTTL(k, v, opts.Cost, opts.Expiration); !set {
		return fmt.Errorf("An error has occurred while setting value '%v' on key '%v'", value, key)
	}

	if opts.SynchronousSet {
		s.client.Wait()
	}

	if tags := opts.Tags; len(tags) > 0 {
		s.setTags(ctx, key, tags)
	}

	return nil
}

func (s *RistrettoStore[K, V]) setTags(ctx context.Context, key any, tags []string) {
	cacheKey, ok := key.(string)
	if !ok {
		// Tags are stored as a comma-separated list of cache keys, which requires
		// the keys to be strings.
		return
	}

	for _, tag := range tags {
		tagKey := fmt.Sprintf(RistrettoTagPattern, tag)

		// The whole read-modify-write sequence has to happen under the same lock:
		// otherwise two concurrent Set calls on the same tag both read the same
		// list of keys and the last write wins, losing the other one key.
		s.mu.Lock()

		cacheKeys := s.getCacheKeysForTag(ctx, tagKey)

		alreadyInserted := false
		for _, currentKey := range cacheKeys {
			if currentKey == cacheKey {
				alreadyInserted = true
				break
			}
		}

		if !alreadyInserted {
			cacheKeys = append(cacheKeys, cacheKey)
		}

		// The tag list can only be persisted when the store value type is able to
		// hold it, which is the case for the usual string, []byte and any types.
		if value, ok := tagValue[V](cacheKeys); ok {
			_ = s.Set(ctx, tagKey, value, lib_store.WithExpiration(720*time.Hour))
		}

		s.mu.Unlock()
	}
}

// getCacheKeysForTag returns the list of cache keys associated to a given tag key
func (s *RistrettoStore[K, V]) getCacheKeysForTag(ctx context.Context, tagKey string) []string {
	result, err := s.Get(ctx, tagKey)
	if err != nil {
		return []string{}
	}

	switch v := any(result).(type) {
	case string:
		if v == "" {
			return []string{}
		}
		return strings.Split(v, ",")
	case []byte:
		if len(v) == 0 {
			return []string{}
		}
		return strings.Split(string(v), ",")
	}

	return []string{}
}

// tagValue returns the list of cache keys of a tag as a value the store is able
// to hold, when its value type allows it
func tagValue[V any](cacheKeys []string) (V, bool) {
	joined := strings.Join(cacheKeys, ",")

	if value, ok := any(joined).(V); ok {
		return value, true
	}

	value, ok := any([]byte(joined)).(V)

	return value, ok
}

// unsupportedTypeError returns the error to be returned when a key or a value
// does not match the type the store has been instantiated with
func unsupportedTypeError[T any](name string, value any) error {
	return fmt.Errorf(
		"%s type not supported by Ristretto store: got %T, expected %s",
		name,
		value,
		reflect.TypeOf(new(T)).Elem(),
	)
}

// Delete removes data in Ristretto memory cache for given key identifier
func (s *RistrettoStore[K, V]) Delete(_ context.Context, key any) error {
	k, ok := key.(K)
	if !ok {
		return unsupportedTypeError[K]("key", key)
	}

	s.client.Del(k)

	return nil
}

// Invalidate invalidates some cache data in Ristretto for given options
func (s *RistrettoStore[K, V]) Invalidate(ctx context.Context, options ...lib_store.InvalidateOption) error {
	opts := lib_store.ApplyInvalidateOptions(options...)

	for _, tag := range opts.Tags {
		tagKey := fmt.Sprintf(RistrettoTagPattern, tag)

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
func (s *RistrettoStore[K, V]) Clear(_ context.Context) error {
	s.client.Clear()
	return nil
}

// GetType returns the store type
func (s *RistrettoStore[K, V]) GetType() string {
	return RistrettoType
}

// Close stops the goroutines owned by the underlying Ristretto client.
//
// Ristretto spawns goroutines that live until Close is called, so this has to be
// done when the store is not used anymore: cache.Cache also exposes a Close
// method that calls this one.
func (s *RistrettoStore[K, V]) Close() error {
	if closer, ok := s.client.(interface{ Close() }); ok {
		closer.Close()
	}

	return nil
}
