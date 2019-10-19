package cache

import (
	"strings"

	"github.com/eko/gocache/codec"
	"github.com/eko/gocache/store"
)

const (
	// CacheType represents the cache type as a string value
	CacheType = "cache"
)

// Cache represents the configuration needed by a cache
type Cache struct {
	codec codec.CodecInterface
}

// New instanciates a new cache entry
func New(store store.StoreInterface) *Cache {
	return &Cache{
		codec: codec.New(store),
	}
}

// Get returns the object stored in cache if it exists
func (c *Cache) Get(key interface{}) (interface{}, error) {
	cacheKey := c.getCacheKey(key)
	return c.codec.Get(cacheKey)
}

// Set populates the cache item using the given key
func (c *Cache) Set(key, object interface{}, options *store.Options) error {
	cacheKey := c.getCacheKey(key)
	return c.codec.Set(cacheKey, object, options)
}

// Delete removes the cache item using the given key
func (c *Cache) Delete(key interface{}) error {
	cacheKey := c.getCacheKey(key)
	return c.codec.Delete(cacheKey)
}

// Invalidate invalidates cache item from given options
func (c *Cache) Invalidate(options store.InvalidateOptions) error {
	return c.codec.Invalidate(options)
}

// GetCodec returns the current codec
func (c *Cache) GetCodec() codec.CodecInterface {
	return c.codec
}

// GetType returns the cache type
func (c *Cache) GetType() string {
	return CacheType
}

// getCacheKey returns the cache key for the given key object by computing a
// checksum of key struct
func (c *Cache) getCacheKey(key interface{}) string {
	return strings.ToLower(checksum(key))
}
