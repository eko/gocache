package cache

import (
	"fmt"
	"log"

	"github.com/eko/gocache/store"
)

const (
	ChainType = "chain"
)

// ChainCache represents the configuration needed by a cache aggregator
type ChainCache struct {
	caches []SetterCacheInterface
}

// NewChain instanciates a new cache aggregator
func NewChain(caches ...SetterCacheInterface) *ChainCache {
	return &ChainCache{
		caches: caches,
	}
}

// Get returns the object stored in cache if it exists
func (c *ChainCache) Get(key interface{}) (interface{}, error) {
	var object interface{}
	var err error

	for _, cache := range c.caches {
		storeType := cache.GetCodec().GetStore().GetType()
		object, err = cache.Get(key)
		if err == nil {
			// Set the value back until this cache layer
			go c.setUntil(key, object, &storeType)
			return object, nil
		}

		log.Printf("Unable to retrieve item from cache with store '%s': %v\n", storeType, err)
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

// setUntil sets a value in available caches, eventually until a given cache layer
func (c *ChainCache) setUntil(key, object interface{}, until *string) error {
	for _, cache := range c.caches {
		if until != nil && *until == cache.GetCodec().GetStore().GetType() {
			break
		}

		err := cache.Set(key, object, nil)
		if err != nil {
			storeType := cache.GetCodec().GetStore().GetType()
			return fmt.Errorf("Unable to set item into cache with store '%s': %v", storeType, err)
		}
	}

	return nil
}

// GetCaches returns all Chaind caches
func (c *ChainCache) GetCaches() []SetterCacheInterface {
	return c.caches
}

// GetType returns the cache type
func (c *ChainCache) GetType() string {
	return ChainType
}
