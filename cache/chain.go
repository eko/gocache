package cache

import (
	"fmt"

	"log"
)

const (
	ChainType = "chain"
)

type loadFunction func(key interface{}) (interface{}, error)

// ChainCache represents the configuration needed by a cache aggregator
type ChainCache struct {
	loadFunc loadFunction
	caches   []SetterCacheInterface
}

// NewChain instanciates a new cache aggregator
func NewChain(loadFunc loadFunction, caches ...SetterCacheInterface) *ChainCache {
	return &ChainCache{
		loadFunc: loadFunc,
		caches:   caches,
	}
}

// Get returns the object stored in cache if it exists
func (c *ChainCache) Get(key interface{}) (interface{}, error) {
	var err error

	for _, cache := range c.caches {
		storeType := cache.GetCodec().GetStore().GetType()
		object, err := cache.Get(key)
		if err == nil {
			// Set the value back until this cache layer
			go c.setUntil(key, object, &storeType)
			return object, nil
		}

		log.Printf("Unable to retrieve item from cache with store '%s': %v\n", storeType, err)
	}

	// Unable to find in all caches, load it from remote source
	object, err := c.loadFunc(key)
	if err != nil {
		log.Printf("An error has occured while trying to load item from load function: %v\n", err)
		return object, err
	}

	// Then, put it in all available caches
	go c.setUntil(key, object, nil)

	return object, err
}

// Set sets a value in available caches
func (c *ChainCache) Set(key, object interface{}) error {
	for _, cache := range c.caches {
		err := cache.Set(key, object)
		if err != nil {
			storeType := cache.GetCodec().GetStore().GetType()
			return fmt.Errorf("Unable to set item into cache with store '%s': %v", storeType, err)
		}
	}

	return nil
}

// setUntil sets a value in available caches, eventually until a given cache layer
func (c *ChainCache) setUntil(key, object interface{}, until *string) error {
	for _, cache := range c.caches {
		if until != nil && *until == cache.GetCodec().GetStore().GetType() {
			break
		}

		err := cache.Set(key, object)
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
