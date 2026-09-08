package marshaler

import (
	"context"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	// MarshalerType represents the marshaler cache type as a string value
	MarshalerType = "marshaler"
)

// Cache marshals and unmarshals the values of a cache of bytes, so that any type
// can be stored in a store that only handles []byte or string.
//
// Unlike Marshaler, it returns the value instead of filling a given object,
// which makes it a cache.CacheInterface[T]: it can be given to NewLoadable and
// wrapped like any other cache.
type Cache[T any] struct {
	cache cache.CacheInterface[[]byte]
}

// NewCache creates a marshaler cache on top of a cache of bytes
func NewCache[T any](cache cache.CacheInterface[[]byte]) *Cache[T] {
	return &Cache[T]{
		cache: cache,
	}
}

// Get obtains a value from cache and unmarshals it
func (c *Cache[T]) Get(ctx context.Context, key any) (T, error) {
	value, err := c.cache.Get(ctx, key)
	if err != nil {
		return *new(T), err
	}

	return unmarshal[T](value)
}

// GetWithTTL obtains a value from cache with its corresponding TTL and unmarshals it.
//
// It returns store.ErrNotSupported when the underlying cache does not expose the
// TTL of its values.
func (c *Cache[T]) GetWithTTL(ctx context.Context, key any) (T, time.Duration, error) {
	getter, ok := c.cache.(interface {
		GetWithTTL(ctx context.Context, key any) ([]byte, time.Duration, error)
	})
	if !ok {
		return *new(T), 0, store.ErrNotSupported
	}

	value, ttl, err := getter.GetWithTTL(ctx, key)
	if err != nil {
		return *new(T), ttl, err
	}

	object, err := unmarshal[T](value)

	return object, ttl, err
}

// Set marshals a value and sets it in cache
func (c *Cache[T]) Set(ctx context.Context, key any, object T, options ...store.Option) error {
	value, err := msgpack.Marshal(object)
	if err != nil {
		return err
	}

	return c.cache.Set(ctx, key, value, options...)
}

// Delete removes a value from the cache
func (c *Cache[T]) Delete(ctx context.Context, key any) error {
	return c.cache.Delete(ctx, key)
}

// Invalidate invalidates cache values using given options
func (c *Cache[T]) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	return c.cache.Invalidate(ctx, options...)
}

// Clear resets all cache data
func (c *Cache[T]) Clear(ctx context.Context) error {
	return c.cache.Clear(ctx)
}

// GetType returns the cache type
func (c *Cache[T]) GetType() string {
	return MarshalerType
}

// unmarshal returns the given cached bytes as a T
func unmarshal[T any](value []byte) (T, error) {
	object := new(T)
	if err := msgpack.Unmarshal(value, object); err != nil {
		return *new(T), err
	}

	return *object, nil
}
