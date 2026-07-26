package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/eko/gocache/lib/v4/store"
)

const (
	// ChainType represents the chain cache type as a string value
	ChainType = "chain"
)

type chainKeyValue[T any] struct {
	key          any
	value        T
	ttl          time.Duration
	cacheAddress *string
}

// ChainCache represents the configuration needed by a cache aggregator
type ChainCache[T any] struct {
	caches     []SetterCacheInterface[T]
	setChannel chan *chainKeyValue[T]
	done       chan struct{}
	closeOnce  sync.Once
	setterWg   sync.WaitGroup
}

// NewChain instantiates a new cache aggregator.
//
// It starts a background goroutine responsible for setting values back into the
// upper cache layers: call Close when the chain is not used anymore to release it.
func NewChain[T any](caches ...SetterCacheInterface[T]) *ChainCache[T] {
	chain := &ChainCache[T]{
		caches:     caches,
		setChannel: make(chan *chainKeyValue[T], 10000),
		done:       make(chan struct{}),
	}

	chain.setterWg.Add(1)
	go chain.setter()

	return chain
}

// setter sets values in available caches until the chain is closed
func (c *ChainCache[T]) setter() {
	defer c.setterWg.Done()

	for {
		select {
		case item := <-c.setChannel:
			c.setUntilCacheAddress(item)
		case <-c.done:
			c.drain()
			return
		}
	}
}

// drain sets the items still buffered when the chain has been closed
func (c *ChainCache[T]) drain() {
	for {
		select {
		case item := <-c.setChannel:
			c.setUntilCacheAddress(item)
		default:
			return
		}
	}
}

// setUntilCacheAddress sets a value in available caches, until a given cache layer
func (c *ChainCache[T]) setUntilCacheAddress(item *chainKeyValue[T]) {
	for _, cache := range c.caches {
		cacheAddress := fmt.Sprintf("%p", cache)

		if item.cacheAddress != nil && *item.cacheAddress == cacheAddress {
			return
		}

		cache.Set(context.Background(), item.key, item.value, store.WithExpiration(item.ttl))
	}
}

// Get returns the object stored in cache if it exists
func (c *ChainCache[T]) Get(ctx context.Context, key any) (T, error) {
	var object T
	var err error
	var ttl time.Duration

	for _, cache := range c.caches {
		cacheAddress := fmt.Sprintf("%p", cache)

		object, ttl, err = cache.GetWithTTL(ctx, key)
		if err == nil {
			// Set the value back until this cache layer
			select {
			case c.setChannel <- &chainKeyValue[T]{key, object, ttl, &cacheAddress}:
			case <-c.done:
			}
			return object, nil
		}
	}

	return object, err
}

// Set sets a value in available caches
func (c *ChainCache[T]) Set(ctx context.Context, key any, object T, options ...store.Option) error {
	errs := []error{}
	for _, cache := range c.caches {
		err := cache.Set(ctx, key, object, options...)
		if err != nil {
			storeType := cache.GetCodec().GetStore().GetType()
			errs = append(errs, fmt.Errorf("unable to set item into cache with store '%s': %w", storeType, err))
		}
	}
	return errors.Join(errs...)
}

// Delete removes a value from all available caches
func (c *ChainCache[T]) Delete(ctx context.Context, key any) error {
	for _, cache := range c.caches {
		cache.Delete(ctx, key)
	}

	return nil
}

// Invalidate invalidates cache item from given options
func (c *ChainCache[T]) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	for _, cache := range c.caches {
		cache.Invalidate(ctx, options...)
	}

	return nil
}

// Clear resets all cache data
func (c *ChainCache[T]) Clear(ctx context.Context) error {
	for _, cache := range c.caches {
		cache.Clear(ctx)
	}

	return nil
}

// GetCaches returns all Chained caches
func (c *ChainCache[T]) GetCaches() []SetterCacheInterface[T] {
	return c.caches
}

// Close releases the background goroutine started by NewChain, after having set
// the values that were still waiting to be propagated to the upper cache layers.
// It is safe to call Close multiple times.
func (c *ChainCache[T]) Close() error {
	c.closeOnce.Do(func() {
		close(c.done)
	})

	c.setterWg.Wait()

	return nil
}

// GetType returns the cache type
func (c *ChainCache[T]) GetType() string {
	return ChainType
}
