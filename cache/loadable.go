package cache

import (
	"context"
	"sync"

	"github.com/eko/gocache/v2/store"
)

const (
	// LoadableType represents the loadable cache type as a string value
	LoadableType = "loadable"
)

type loadableKeyValue struct {
	key   interface{}
	value interface{}
}

type loadFunction func(ctx context.Context, key interface{}) (interface{}, error)

// LoadableCache represents a cache that uses a function to load data
type LoadableCache struct {
	loadFunc   loadFunction
	cache      CacheInterface
	setChannel chan *loadableKeyValue
	setterWg   *sync.WaitGroup
}

// NewLoadable instanciates a new cache that uses a function to load data
func NewLoadable(loadFunc loadFunction, cache CacheInterface) *LoadableCache {
	loadable := &LoadableCache{
		loadFunc:   loadFunc,
		cache:      cache,
		setChannel: make(chan *loadableKeyValue, 10000),
		setterWg:   &sync.WaitGroup{},
	}

	loadable.setterWg.Add(1)
	go loadable.setter()

	return loadable
}

func (c *LoadableCache) setter() {
	defer c.setterWg.Done()

	for item := range c.setChannel {
		c.Set(context.Background(), item.key, item.value, nil)
	}
}

// Get returns the object stored in cache if it exists
func (c *LoadableCache) Get(ctx context.Context, key interface{}) (interface{}, error) {
	var err error

	object, err := c.cache.Get(ctx, key)
	if err == nil {
		return object, err
	}

	// Unable to find in cache, try to load it from load function
	object, err = c.loadFunc(ctx, key)
	if err != nil {
		return object, err
	}

	// Then, put it back in cache
	c.setChannel <- &loadableKeyValue{key, object}

	return object, err
}

// Set sets a value in available caches
func (c *LoadableCache) Set(ctx context.Context, key, object interface{}, options *store.Options) error {
	return c.cache.Set(ctx, key, object, options)
}

// Delete removes a value from cache
func (c *LoadableCache) Delete(ctx context.Context, key interface{}) error {
	return c.cache.Delete(ctx, key)
}

// Invalidate invalidates cache item from given options
func (c *LoadableCache) Invalidate(ctx context.Context, options store.InvalidateOptions) error {
	return c.cache.Invalidate(ctx, options)
}

// Clear resets all cache data
func (c *LoadableCache) Clear(ctx context.Context) error {
	return c.cache.Clear(ctx)
}

// GetType returns the cache type
func (c *LoadableCache) GetType() string {
	return LoadableType
}

func (c *LoadableCache) Close() error {
	close(c.setChannel)
	c.setterWg.Wait()

	return nil
}
