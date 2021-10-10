package marshaler

import (
	"context"

	"github.com/eko/gocache/v2/cache"
	"github.com/eko/gocache/v2/store"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Marshaler is the struct that marshal and unmarshal cache values
type ProtoMarshaler struct {
	marshalOpts   protojson.MarshalOptions
	unmarshalOpts protojson.UnmarshalOptions
	cache         cache.CacheInterface
}

// New creates a new marshaler that marshals/unmarshals cache values
func NewProtoMarshaler(cache cache.CacheInterface, opts ...ProtoMarshalerOption) *ProtoMarshaler {
	m := &ProtoMarshaler{
		cache:         cache,
		marshalOpts:   protojson.MarshalOptions{},
		unmarshalOpts: protojson.UnmarshalOptions{},
	}

	for _, o := range opts {
		o(m)
	}

	return m
}

// Get obtains a value from cache and unmarshal value with given object
func (c *ProtoMarshaler) Get(ctx context.Context, key interface{}, returnObj proto.Message) (interface{}, error) {
	result, err := c.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	switch v := result.(type) {
	case []byte:
		err = c.unmarshalOpts.Unmarshal(v, returnObj)
	case string:
		err = c.unmarshalOpts.Unmarshal([]byte(v), returnObj)
	}

	if err != nil {
		return nil, err
	}

	return returnObj, nil
}

// Set sets a value in cache by marshaling value
func (c *ProtoMarshaler) Set(ctx context.Context, key interface{}, object proto.Message, options *store.Options) error {
	bytes, err := c.marshalOpts.Marshal(object)
	if err != nil {
		return err
	}

	return c.cache.Set(ctx, key, bytes, options)
}

// Delete removes a value from the cache
func (c *ProtoMarshaler) Delete(ctx context.Context, key interface{}) error {
	return c.cache.Delete(ctx, key)
}

// Invalidate invalidate cache values using given options
func (c *ProtoMarshaler) Invalidate(ctx context.Context, options store.InvalidateOptions) error {
	return c.cache.Invalidate(ctx, options)
}

// Clear reset all cache data
func (c *ProtoMarshaler) Clear(ctx context.Context) error {
	return c.cache.Clear(ctx)
}
