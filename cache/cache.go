package cache

import (
	"crypto"
	"fmt"
	"reflect"
	"strings"

	"github.com/yeqown/gocache/store"
)

const (
	// CacheType represents the cache type as a string value
	CacheType = "cache"
)

var (
	_ ICache = &Cache{}
)

// Cache represents the configuration needed by a cache
type Cache struct {
	store store.StoreInterface
}

// New instantiate a new cache entry
func New(store store.StoreInterface) ICache {
	return &Cache{
		store: store,
	}
}

// Get returns the object stored in cache if it exists
func (c *Cache) Get(key interface{}, returnObj interface{}) (interface{}, error) {
	cacheKey := c.getCacheKey(key)
	return c.store.Get(cacheKey)
}

// Set populates the cache item using the given key
func (c *Cache) Set(key, object interface{}, options *store.Options) error {
	cacheKey := c.getCacheKey(key)
	return c.store.Set(cacheKey, object, options)
}

// Delete removes the cache item using the given key
func (c *Cache) Delete(key interface{}) error {
	cacheKey := c.getCacheKey(key)
	return c.store.Delete(cacheKey)
}

// Invalidate invalidates cache item from given options
func (c *Cache) Invalidate(options store.InvalidateOptions) error {
	return c.store.Invalidate(options)
}

// Clear resets all cache data
func (c *Cache) Clear() error {
	return c.store.Clear()
}

// GetType returns the cache type
func (c *Cache) GetType() string {
	return CacheType
}

func (c *Cache) GetStats() *Stats {
	panic("implement me")
}

// getCacheKey returns the cache key for the given key object by computing a
// checksum of key struct
func (c *Cache) getCacheKey(key interface{}) string {
	return strings.ToLower(checksum(key))
}

// checksum hashes a given object into a string
func checksum(object interface{}) string {
	digester := crypto.MD5.New()
	fmt.Fprint(digester, reflect.TypeOf(object))
	fmt.Fprint(digester, object)
	hash := digester.Sum(nil)

	return fmt.Sprintf("%x", hash)
}
