package cache

import (
	"log"
	"github.com/eko/gache/store"
)

const (
	LoadableType = "loadable"
)

type loadFunction func(key interface{}) (interface{}, error)

// LoadableCache represents a cache that uses a function to load data
type LoadableCache struct {
	loadFunc loadFunction
	cache    CacheInterface
}

// NewLoadable instanciates a new cache that uses a function to load data
func NewLoadable(loadFunc loadFunction, cache CacheInterface) *LoadableCache {
	return &LoadableCache{
		loadFunc: loadFunc,
		cache:    cache,
	}
}

// Get returns the object stored in cache if it exists
func (c *LoadableCache) Get(key interface{}) (interface{}, error) {
	var err error

	object, err := c.cache.Get(key)
	if err == nil {
		return object, err
	}

	// Unable to find in cache, try to load it from load function
	object, err = c.loadFunc(key)
	if err != nil {
		log.Printf("An error has occured while trying to load item from load function: %v\n", err)
		return object, err
	}

	// Then, put it back in cache
	go c.Set(key, object, nil)

	return object, err
}

// Set sets a value in available caches
func (c *LoadableCache) Set(key, object interface{}, options *store.Options) error {
	return c.cache.Set(key, object, options)
}

// GetType returns the cache type
func (c *LoadableCache) GetType() string {
	return LoadableType
}
