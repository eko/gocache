package serializer

import (
	"context"

	"github.com/eko/gocache/lib/v4/cache"
	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/vmihailenco/msgpack/v5"
)

var _ cache.CacheInterface[any] = (*SerializerCache[any])(nil)

type SerializerCache[T any] struct {
	cache      cache.CacheInterface[[]byte]
	serializer Serializer
}

func New[T any](serializer Serializer, cacheInterface cache.CacheInterface[[]byte]) *SerializerCache[T] {
	serializerCache := &SerializerCache[T]{
		cache:      cacheInterface,
		serializer: serializer,
	}
	return serializerCache
}

func (c *SerializerCache[T]) Get(ctx context.Context, key any) (T, error) {
	var zero T

	result, err := c.cache.Get(ctx, key)
	if err != nil {
		return zero, err
	}

	var returnObj T
	err = c.serializer.Unmarshal(result, &returnObj)
	if err != nil {
		return zero, err
	}

	return returnObj, nil
}

func (c *SerializerCache[T]) Set(ctx context.Context, key any, object T, options ...lib_store.Option) error {
	bytes, err := c.serializer.Marshal(object)
	if err != nil {
		return err
	}

	return c.cache.Set(ctx, key, bytes, options...)
}

func (c *SerializerCache[T]) Delete(ctx context.Context, key any) error {
	return c.cache.Delete(ctx, key)
}

func (c *SerializerCache[T]) Invalidate(ctx context.Context, options ...lib_store.InvalidateOption) error {
	return c.cache.Invalidate(ctx, options...)
}

func (c *SerializerCache[T]) Clear(ctx context.Context) error {
	return c.cache.Clear(ctx)
}

func (c *SerializerCache[T]) GetType() string {
	return "marshaler"
}

// TODO tests here

// DefaultSerializer msgpack by default
type DefaultSerializer struct {
	MarshalFn   func(any) ([]byte, error)
	UnmarshalFn func([]byte, any) error
}

func (c DefaultSerializer) Marshal(a any) ([]byte, error) {
	if c.MarshalFn != nil {
		return c.MarshalFn(a)
	}
	return msgpack.Marshal(a)
}

func (c DefaultSerializer) Unmarshal(bytes []byte, a any) error {
	if c.UnmarshalFn != nil {
		return c.UnmarshalFn(bytes, a)
	}
	return msgpack.Unmarshal(bytes, a)
}
