package cache

import (
	"context"
	"fmt"

	"github.com/eko/gocache/lib/v4/store"
	"golang.org/x/sync/singleflight"
)

const (
	// LoadableType represents the loadable cache type as a string value
	LoadableType = "loadable"
)

type loadableKeyValue[T any] struct {
	key   any
	value T
}

type LoadFunction[T any] func(ctx context.Context, key any) (T, error)

// LoadableCache represents a cache that uses a function to load data
type LoadableCache[T any] struct {
	loadFunc  LoadFunction[T]
	cache     CacheInterface[T]
	loadGroup *singleflight.Group
}

// NewLoadable instanciates a new cache that uses a function to load data
func NewLoadable[T any](loadFunc LoadFunction[T], cache CacheInterface[T]) *LoadableCache[T] {
	return &LoadableCache[T]{
		loadFunc:  loadFunc,
		cache:     cache,
		loadGroup: &singleflight.Group{},
	}
}

// Get returns the object stored in cache if it exists
func (c *LoadableCache[T]) Get(ctx context.Context, key any) (T, error) {
	var err error

	object, err := c.cache.Get(ctx, key)
	if err == nil {
		return object, err
	}

	_, err, _ = c.loadGroup.Do(fmt.Sprintf("%v", key), func() (interface{}, error) {
		// Unable to find in cache, try to load it from load function
		object, err = c.loadFunc(ctx, key)
		if err != nil {
			return object, err
		}

		// Then, put it back in cache
		_ = c.Set(context.Background(), key, object)

		return object, nil
	})

	return object, err
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

	return nil
}
