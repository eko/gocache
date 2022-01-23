package cache

import (
	"context"
	"crypto"
	"fmt"
	"reflect"
	"time"

	"github.com/eko/gocache/v2/codec"
	"github.com/eko/gocache/v2/store"
)

const (
	// CacheType represents the cache type as a string value
	CacheType = "cache"
)

type CacheKeyGenerator interface {
	GetCacheKey() string
}

// Cache represents the configuration needed by a cache
type Cache struct {
	codec codec.CodecInterface
}

// New instantiates a new cache entry
func New(store store.StoreInterface) *Cache {
	return &Cache{
		codec: codec.New(store),
	}
}

// Get returns the object stored in cache if it exists
func (c *Cache) Get(ctx context.Context, key interface{}) (interface{}, error) {
	cacheKey := c.getCacheKey(key)
	return c.codec.Get(ctx, cacheKey)
}

// GetWithTTL returns the object stored in cache and its corresponding TTL
func (c *Cache) GetWithTTL(ctx context.Context, key interface{}) (interface{}, time.Duration, error) {
	cacheKey := c.getCacheKey(key)
	return c.codec.GetWithTTL(ctx, cacheKey)
}

// Set populates the cache item using the given key
func (c *Cache) Set(ctx context.Context, key, object interface{}, options *store.Options) error {
	cacheKey := c.getCacheKey(key)
	return c.codec.Set(ctx, cacheKey, object, options)
}

// Delete removes the cache item using the given key
func (c *Cache) Delete(ctx context.Context, key interface{}) error {
	cacheKey := c.getCacheKey(key)
	return c.codec.Delete(ctx, cacheKey)
}

// Invalidate invalidates cache item from given options
func (c *Cache) Invalidate(ctx context.Context, options store.InvalidateOptions) error {
	return c.codec.Invalidate(ctx, options)
}

// Clear resets all cache data
func (c *Cache) Clear(ctx context.Context) error {
	return c.codec.Clear(ctx)
}

// GetCodec returns the current codec
func (c *Cache) GetCodec() codec.CodecInterface {
	return c.codec
}

// GetType returns the cache type
func (c *Cache) GetType() string {
	return CacheType
}

// getCacheKey returns the cache key for the given key object by returning
// the key if type is string or by computing a checksum of key structure
// if its type is other than string
func (c *Cache) getCacheKey(key interface{}) string {
	switch key.(type) {
	case string:
		return key.(string)
	case CacheKeyGenerator:
		return key.(CacheKeyGenerator).GetCacheKey()
	default:
		return checksum(key)
	}
}

// checksum hashes a given object into a string
func checksum(object interface{}) string {
	digester := crypto.MD5.New()
	fmt.Fprint(digester, reflect.TypeOf(object))
	fmt.Fprint(digester, object)
	hash := digester.Sum(nil)

	return fmt.Sprintf("%x", hash)
}
