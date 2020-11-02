package cache

import (
	"crypto"
	"fmt"
	"reflect"
	"time"

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

// New instantiates a new cache entry
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

// GetWithTTL returns the object stored in cache and its corresponding TTL
func (c *Cache) GetWithTTL(key interface{}) (interface{}, time.Duration, error) {
	cacheKey := c.getCacheKey(key)
	return c.codec.GetWithTTL(cacheKey)
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

// Clear resets all cache data
func (c *Cache) Clear() error {
	return c.codec.Clear()
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
