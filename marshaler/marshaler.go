package marshaler

import (
	"context"

	"github.com/eko/gocache/v2/cache"
	"github.com/eko/gocache/v2/store"
	"github.com/vmihailenco/msgpack"
)

// Marshaler is the struct that marshal and unmarshal cache values
type Marshaler struct {
	cache cache.CacheInterface
}

// New creates a new marshaler that marshals/unmarshals cache values
func New(cache cache.CacheInterface) *Marshaler {
	return &Marshaler{
		cache: cache,
	}
}

// Get obtains a value from cache and unmarshal value with given object
func (c *Marshaler) Get(ctx context.Context, key interface{}, returnObj interface{}) (interface{}, error) {
	result, err := c.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	switch v := result.(type) {
	case []byte:
		err = msgpack.Unmarshal(v, returnObj)
	case string:
		err = msgpack.Unmarshal([]byte(v), returnObj)
	}

	if err != nil {
		return nil, err
	}

	return returnObj, nil
}

// Set sets a value in cache by marshaling value
func (c *Marshaler) Set(ctx context.Context, key, object interface{}, options *store.Options) error {
	bytes, err := msgpack.Marshal(object)
	if err != nil {
		return err
	}

	return c.cache.Set(ctx, key, bytes, options)
}

// Delete removes a value from the cache
func (c *Marshaler) Delete(ctx context.Context, key interface{}) error {
	return c.cache.Delete(ctx, key)
}

// Invalidate invalidate cache values using given options
func (c *Marshaler) Invalidate(ctx context.Context, options store.InvalidateOptions) error {
	return c.cache.Invalidate(ctx, options)
}

// Clear reset all cache data
func (c *Marshaler) Clear(ctx context.Context) error {
	return c.cache.Clear(ctx)
}
