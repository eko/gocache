package cache

import (
	"strings"

	"github.com/eko/gache/codec"
	"github.com/eko/gache/store"
)

const (
	CacheType = "cache"
)

// Cache represents the configuration needed by a cache
type Cache struct {
	codec   codec.CodecInterface
	options *Options
}

// New instanciates a new cache entry
func New(store store.StoreInterface, options *Options) *Cache {
	if options == nil {
		options = &Options{}
	}

	return &Cache{
		codec:   codec.New(store),
		options: options,
	}
}

// Get returns the object stored in cache if it exists
func (c *Cache) Get(key interface{}) (interface{}, error) {
	cacheKey := c.getCacheKey(key)
	return c.codec.Get(cacheKey)
}

// Set populates the cache item using the given key
func (c *Cache) Set(key, object interface{}) error {
	cacheKey := c.getCacheKey(key)
	return c.codec.Set(cacheKey, object, c.options.ExpirationValue())
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
