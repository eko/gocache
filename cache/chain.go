package cache

import (
	"fmt"
	"log"

	"github.com/yeqown/gocache/store"
)

const (
	// ChainType represents the chain cache type as a string value
	ChainType = "chain"
)

var (
	_ ICache = &ChainCache{}
)

// TODO: 针对链式缓存组件，设计一个错误类型
type chainKeyValue struct {
	key       interface{}
	value     interface{}
	storeType *string
}

// ChainCache represents the configuration needed by a cache aggregator
type ChainCache struct {
	caches     []ICache
	setChannel chan *chainKeyValue
}

func (c *ChainCache) GetStats() *Stats {
	for _, c := range c.caches {
		if stat := c.GetStats(); stat != nil {
			return stat
		}
	}

	return nil
}

// NewChain instanciates a new cache aggregator
func NewChain(caches ...ICache) *ChainCache {
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
			if item.storeType != nil && *item.storeType == cache.GetType() {
				break
			}

			cache.Set(item.key, item.value, nil)
		}
	}
}

// Get returns the object stored in cache if it exists
func (c *ChainCache) Get(key interface{}, returnObj interface{}) (interface{}, error) {
	var object interface{}
	var err error

	for _, cache := range c.caches {
		storeType := cache.GetType()
		object, err = cache.Get(key, returnObj)
		if err == nil {
			// Set the value back until this cache layer
			c.setChannel <- &chainKeyValue{key, object, &storeType}
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
			storeType := cache.GetType()
			return fmt.Errorf("Unable to set item into cache with store %s:%v", storeType, err)
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

// GetCaches returns all Chaind caches
func (c *ChainCache) GetCaches() []ICache {
	return c.caches
}

// GetType returns the cache type
func (c *ChainCache) GetType() string {
	return ChainType
}
