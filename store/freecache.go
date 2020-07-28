package store

import (
	"errors"
	"fmt"
)

const (
	// FreecacheType represents the storage type as a string value
	FreecacheType = "freecache"
	// FreecacheTagPattern represents the tag pattern to be used as a key in specified storage
	FreecacheTagPattern = "freecache_tag_%s"
)

// FreecacheClientInterface represents a coocood/freecache client
type FreecacheClientInterface interface {
	Get(key []byte) (value []byte, err error)
	GetInt(key int64) (value []byte, err error)
	Set(key, value []byte, expireSeconds int) (err error)
	SetInt(key int64, value []byte, expireSeconds int) (err error)
	Del(key []byte) (affected bool)
	DelInt(key int64) (affected bool)
	Clear()
}

type FreecacheStore struct {
	client  FreecacheClientInterface
	options *Options
}

func NewFreecache(client FreecacheClientInterface, options *Options) *FreecacheStore {
	if options == nil {
		options = &Options{}
	}

	return &FreecacheStore{
		client:  client,
		options: options,
	}
}
func (f *FreecacheStore) Get(key interface{}) (interface{}, error) {
	var err error
	var result interface{}
	if k, ok := key.(string); ok {
		result, err = f.client.Get([]byte(k))
		if err != nil {
			return nil, errors.New("value not found in Freecache store")
		}
		return result, err
	}

	return nil, errors.New("key type not supported by Freecache store")

}

// Set sets a key, value and expiration for a cache entry and stores it in the cache.
// If the key is larger than 65535 or value is larger than 1/1024 of the cache size,
// the entry will not be written to the cache. expireSeconds <= 0 means no expire,
// but it can be evicted when cache is full.
func (f *FreecacheStore) Set(key interface{}, value interface{}, options *Options) error {
	var err error
	var val []byte

	//type check for value, as freecache only supports value of type []byte
	switch v := value.(type) {
	case []byte:
		val = v
	default:
		return errors.New("value type not supported by Freecache store")
	}

	if k, ok := key.(string); ok {
		err = f.client.Set([]byte(k), val, int(options.Expiration.Seconds()))
		if err != nil {
			return fmt.Errorf("size of key: %v, value: %v, err: %v", k, len(val), err)
		}
		return err
	}
	return errors.New("key type not supported by Freecache store")
}

func (f *FreecacheStore) Delete(key interface{}) error {
	if v, ok := key.(string); ok {
		if f.client.Del([]byte(v)) {
			return nil
		}
		return fmt.Errorf("failed to delete key %v", key)
	}
	return errors.New("key type not supported by Freecache store")

}

func (f *FreecacheStore) Invalidate(options InvalidateOptions) error {
	if tags := options.TagsValue(); len(tags) > 0 {
		for _, tag := range tags {
			var tagKey = fmt.Sprintf(FreecacheTagPattern, tag)
			return f.Delete([]byte(tagKey))
		}
	}

	return nil
}

func (f *FreecacheStore) Clear() error {
	f.client.Clear()
	return nil
}

func (f *FreecacheStore) GetType() string {
	return FreecacheType
}
