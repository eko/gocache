package marshaler

import (
	"github.com/eko/gocache/cache"
	"github.com/eko/gocache/store"
	"github.com/gogo/protobuf/proto"
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
func (c *Marshaler) Get(key interface{}, returnObj interface{}) (interface{}, error) {
	result, err := c.cache.Get(key)
	if err != nil {
		return nil, err
	}

	switch res := result.(type) {
	case []byte:
		switch v := returnObj.(type) {
		case proto.Message:
			err = proto.Unmarshal(result.([]byte), v)
		default:
			err = msgpack.Unmarshal(res, returnObj)
		}
	case string:
		switch v := returnObj.(type) {
		case proto.Message:
			err = proto.Unmarshal(result.([]byte), v)
		default:
			err = msgpack.Unmarshal([]byte(res), returnObj)
		}
	}

	if err != nil {
		return nil, err
	}

	return returnObj, nil
}

// Set sets a value in cache by marshaling value
func (c *Marshaler) Set(key, object interface{}, options *store.Options) error {
	var bytes []byte
	var err error

	switch obj := object.(type) {
	case proto.Message:
		bytes, err = proto.Marshal(obj)
	default:
		bytes, err = msgpack.Marshal(object)
	}
	if err != nil {
		return err
	}

	return c.cache.Set(key, bytes, options)
}

// Delete removes a value from the cache
func (c *Marshaler) Delete(key interface{}) error {
	return c.cache.Delete(key)
}

// Invalidate invalidate cache values using given options
func (c *Marshaler) Invalidate(options store.InvalidateOptions) error {
	return c.cache.Invalidate(options)
}

// Clear reset all cache data
func (c *Marshaler) Clear() error {
	return c.cache.Clear()
}
