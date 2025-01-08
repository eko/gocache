package cache

import (
	"context"
	"errors"
	"sync"

	"github.com/eko/gocache/lib/v4/store"
	"golang.org/x/sync/singleflight"
)

const (
	// LoadableType represents the loadable cache type as a string value
	LoadableType = "loadable"
)

type loadableKeyValue[T any] struct {
	key     any
	value   T
	options []store.Option
}

type LoadFunction[T any] func(ctx context.Context, key any) (T, []store.Option, error)

// LoadableCache represents a cache that uses a function to load data
type LoadableCache[T any] struct {
	singleFlight singleflight.Group
	loadFunc     LoadFunction[T]
	cache        CacheInterface[T]
	setChannel   chan *loadableKeyValue[T]
	setCache     sync.Map
	setterWg     *sync.WaitGroup
}

// NewLoadable instantiates a new cache that uses a function to load data
func NewLoadable[T any](loadFunc LoadFunction[T], cache CacheInterface[T]) *LoadableCache[T] {
	loadable := &LoadableCache[T]{
		singleFlight: singleflight.Group{},
		loadFunc:     loadFunc,
		cache:        cache,
		setChannel:   make(chan *loadableKeyValue[T], 10000),
		setterWg:     &sync.WaitGroup{},
	}

	loadable.setterWg.Add(1)
	go loadable.setter()

	return loadable
}

func (c *LoadableCache[T]) setter() {
	defer c.setterWg.Done()

	for item := range c.setChannel {
		c.Set(context.Background(), item.key, item.value, item.options...)

		cacheKey := c.getCacheKey(item.key)
		c.singleFlight.Forget(cacheKey)
		c.setCache.Delete(cacheKey)
	}
}

// Get returns the object stored in cache if it exists
func (c *LoadableCache[T]) Get(ctx context.Context, key any) (T, error) {
	var err error

	object, err := c.cache.Get(ctx, key)
	if err == nil {
		return object, err
	}

	// Unable to find in cache, try to load it from load function
	cacheKey := c.getCacheKey(key)
	if v, ok := c.setCache.Load(cacheKey); ok {
		return v.(T), nil
	}

	zero := *new(T)

	rawLoadedResult, err, _ := c.singleFlight.Do(
		cacheKey,
		func() (any, error) {
			value, options, innerErr := c.loadFunc(ctx, key)

			return &loadableKeyValue[T]{
				key:     key,
				value:   value,
				options: options,
			}, innerErr
		},
	)
	if err != nil {
		return zero, err
	}

	loadedKeyValue, ok := rawLoadedResult.(*loadableKeyValue[T])
	if !ok {
		return zero, errors.New(
			"returned value can't be cast to *loadableKeyValue[T]",
		)
	}

	// Then, put it back in cache
	c.setCache.Store(cacheKey, loadedKeyValue.value)
	c.setChannel <- loadedKeyValue

	return loadedKeyValue.value, err
}

// Set sets a value in available caches
func (c *LoadableCache[T]) Set(ctx context.Context, key any, object T, options ...store.Option) error {
	return c.cache.Set(ctx, key, object, options...)
}

// Delete removes a value from cache
func (c *LoadableCache[T]) Delete(ctx context.Context, key any) error {
	return c.cache.Delete(ctx, key)
}

// Invalidate invalidates cache item from given options
func (c *LoadableCache[T]) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	return c.cache.Invalidate(ctx, options...)
}

// Clear resets all cache data
func (c *LoadableCache[T]) Clear(ctx context.Context) error {
	return c.cache.Clear(ctx)
}

// GetType returns the cache type
func (c *LoadableCache[T]) GetType() string {
	return LoadableType
}

func (c *LoadableCache[T]) Close() error {
	close(c.setChannel)
	c.setterWg.Wait()

	return nil
}

// getCacheKey returns the cache key for the given key object by returning
// the key if type is string or by computing a checksum of key structure
// if its type is other than string
func (c *LoadableCache[T]) getCacheKey(key any) string {
	switch v := key.(type) {
	case string:
		return v
	case CacheKeyGenerator:
		return v.GetCacheKey()
	default:
		return checksum(key)
	}
}
