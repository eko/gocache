package cache

import (
	"fmt"
	"time"

	"github.com/eko/gocache/store"
)

const (
	// ChainType represents the chain cache type as a string value
	ChainType = "chain"
)

type chainKeyValue struct {
	key       interface{}
	value     interface{}
	ttl       time.Duration
	storeType *string
}

// ChainCache represents the configuration needed by a cache aggregator
type ChainCache struct {
	caches     []SetterCacheInterface
	setChannel chan *chainKeyValue
}

// NewChain instantiates a new cache aggregator
func NewChain(caches ...SetterCacheInterface) *ChainCache {
	chain := &ChainCache{
		caches:     caches,
		setChannel: make(chan *chainKeyValue, 10000),
	}

	go chain.setter()

	return chain
}

// setter sets a value in available caches, until a given cache layer
func (c *ChainCache) setter() {
	for item := range c.setChannel {
		for _, cache := range c.caches {
			if item.storeType != nil && *item.storeType == cache.GetCodec().GetStore().GetType() {
				break
			}

			cache.Set(item.key, item.value, &store.Options{Expiration: item.ttl})
		}
	}
}

// Get returns the object stored in cache if it exists
func (c *ChainCache) Get(key interface{}) (interface{}, error) {
	var object interface{}
	var err error
	var ttl time.Duration

	for _, cache := range c.caches {
		storeType := cache.GetCodec().GetStore().GetType()
		object, ttl, err = cache.GetWithTTL(key)
		if err == nil {
			// Set the value back until this cache layer
			c.setChannel <- &chainKeyValue{key, object, ttl, &storeType}
			return object, nil
		}
	}

	return object, err
}

// Set sets a value in available caches
func (c *ChainCache) Set(key, object interface{}, options *store.Options) error {
	for _, cache := range c.caches {
		err := cache.Set(key, object, options)
		if err != nil {
			storeType := cache.GetCodec().GetStore().GetType()
			return fmt.Errorf("Unable to set item into cache with store '%s': %v", storeType, err)
		}
	}

	return nil
}

// Delete removes a value from all available caches
func (c *ChainCache) Delete(key interface{}) error {
	for _, cache := range c.caches {
		cache.Delete(key)
	}

	return nil
}

// Invalidate invalidates cache item from given options
func (c *ChainCache) Invalidate(options store.InvalidateOptions) error {
	for _, cache := range c.caches {
		cache.Invalidate(options)
	}

	return nil
}

// Clear resets all cache data
func (c *ChainCache) Clear() error {
	for _, cache := range c.caches {
		cache.Clear()
	}

	return nil
}

// GetCaches returns all Chained caches
func (c *ChainCache) GetCaches() []SetterCacheInterface {
	return c.caches
}

// GetType returns the cache type
func (c *ChainCache) GetType() string {
	return ChainType
}
