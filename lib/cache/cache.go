package cache

import (
	"context"
	"crypto"
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/eko/gocache/lib/v4/codec"
	"github.com/eko/gocache/lib/v4/store"
)

const (
	// CacheType represents the cache type as a string value
	CacheType = "cache"
)

// ErrValueTypeMismatch is returned when the value read from the store cannot be
// represented by the type the cache has been instantiated with.
var ErrValueTypeMismatch = errors.New("value type mismatch")

// codecSetIfNotExists is implemented by codecs forwarding SetIfNotExists to
// their store. It is not part of codec.CodecInterface to keep it compatible.
type codecSetIfNotExists interface {
	SetIfNotExists(ctx context.Context, key any, value any, options ...store.Option) (bool, error)
}

// Cache represents the configuration needed by a cache
type Cache[T any] struct {
	codec codec.CodecInterface
}

// New instantiates a new cache entry
func New[T any](store store.StoreInterface) *Cache[T] {
	return &Cache[T]{
		codec: codec.New(store),
	}
}

// Get returns the object stored in cache if it exists
func (c *Cache[T]) Get(ctx context.Context, key any) (T, error) {
	cacheKey := c.getCacheKey(key)

	value, err := c.codec.Get(ctx, cacheKey)
	if err != nil {
		return *new(T), err
	}

	return castValue[T](value)
}

// GetWithTTL returns the object stored in cache and its corresponding TTL
func (c *Cache[T]) GetWithTTL(ctx context.Context, key any) (T, time.Duration, error) {
	cacheKey := c.getCacheKey(key)

	value, duration, err := c.codec.GetWithTTL(ctx, cacheKey)
	if err != nil {
		return *new(T), duration, err
	}

	object, err := castValue[T](value)

	return object, duration, err
}

// Set populates the cache item using the given key
func (c *Cache[T]) Set(ctx context.Context, key any, object T, options ...store.Option) error {
	cacheKey := c.getCacheKey(key)
	return c.codec.Set(ctx, cacheKey, object, options...)
}

// SetIfNotExists populates the cache item using the given key only when it does
// not already exist, and reports whether it has been written.
//
// It relies on an atomic primitive of the underlying store and returns an error
// wrapping store.ErrNotSupported when the store does not provide one.
func (c *Cache[T]) SetIfNotExists(ctx context.Context, key any, object T, options ...store.Option) (bool, error) {
	setter, ok := c.codec.(codecSetIfNotExists)
	if !ok {
		return false, fmt.Errorf("%w: %s", store.ErrNotSupported, c.GetType())
	}

	return setter.SetIfNotExists(ctx, c.getCacheKey(key), object, options...)
}

// Delete removes the cache item using the given key
func (c *Cache[T]) Delete(ctx context.Context, key any) error {
	cacheKey := c.getCacheKey(key)
	return c.codec.Delete(ctx, cacheKey)
}

// Invalidate invalidates cache item from given options
func (c *Cache[T]) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	return c.codec.Invalidate(ctx, options...)
}

// Clear resets all cache data
func (c *Cache[T]) Clear(ctx context.Context) error {
	return c.codec.Clear(ctx)
}

// GetCodec returns the current codec
func (c *Cache[T]) GetCodec() codec.CodecInterface {
	return c.codec
}

// GetType returns the cache type
func (c *Cache[T]) GetType() string {
	return CacheType
}

// Close releases the resources held by the underlying store when it implements
// io.Closer. Stores backed by a client owning goroutines or connections (such as
// Ristretto or Bigcache) need this to be called when the cache is not used anymore.
func (c *Cache[T]) Close() error {
	if closer, ok := c.codec.GetStore().(io.Closer); ok {
		return closer.Close()
	}

	return nil
}

// getCacheKey returns the cache key for the given key object by returning
// the key if type is string or by computing a checksum of key structure
// if its type is other than string
func (c *Cache[T]) getCacheKey(key any) string {
	switch v := key.(type) {
	case string:
		return v
	case CacheKeyGenerator:
		return v.GetCacheKey()
	default:
		return checksum(key)
	}
}

// castValue returns the given store value as a T.
//
// Stores do not all keep the type they have been given: some of them normalize
// values to []byte or to string, so both representations are accepted when T is
// one of these two types. Any other mismatch returns ErrValueTypeMismatch rather
// than silently returning a zero value.
func castValue[T any](value any) (T, error) {
	if value == nil {
		return *new(T), nil
	}

	if v, ok := value.(T); ok {
		return v, nil
	}

	switch v := value.(type) {
	case []byte:
		if converted, ok := any(string(v)).(T); ok {
			return converted, nil
		}
	case string:
		if converted, ok := any([]byte(v)).(T); ok {
			return converted, nil
		}
	}

	return *new(T), fmt.Errorf(
		"%w: got %T, expected %s",
		ErrValueTypeMismatch,
		value,
		reflect.TypeOf(new(T)).Elem(),
	)
}

// checksum hashes a given object into a string
func checksum(object any) string {
	digester := crypto.MD5.New()
	fmt.Fprint(digester, reflect.TypeOf(object))
	fmt.Fprint(digester, object)
	hash := digester.Sum(nil)

	return fmt.Sprintf("%x", hash)
}
