package store

import (
	"errors"
	"fmt"
	"strings"
	"time"
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
	TTL(key []byte) (timeLeft uint32, err error)
	Set(key, value []byte, expireSeconds int) (err error)
	SetInt(key int64, value []byte, expireSeconds int) (err error)
	Del(key []byte) (affected bool)
	DelInt(key int64) (affected bool)
	Clear()
}

//FreecacheStore is a store for freecache
type FreecacheStore struct {
	client  FreecacheClientInterface
	options *Options
}

// NewFreecache creates a new store to freecache instance(s)
func NewFreecache(client FreecacheClientInterface, options *Options) *FreecacheStore {
	if options == nil {
		options = &Options{}
	}

	return &FreecacheStore{
		client:  client,
		options: options,
	}
}

// Get returns data stored from a given key. It returns the value or not found error
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

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (f *FreecacheStore) GetWithTTL(key interface{}) (interface{}, time.Duration, error) {
	if k, ok := key.(string); ok {
		result, err := f.client.Get([]byte(k))
		if err != nil {
			return nil, 0, errors.New("value not found in Freecache store")
		}

		ttl, err := f.client.TTL([]byte(k))
		if err != nil {
			return nil, 0, errors.New("value not found in Freecache store")
		}

		return result, time.Duration(ttl) * time.Second, err
	}

	return nil, 0, errors.New("key type not supported by Freecache store")
}

// Set sets a key, value and expiration for a cache entry and stores it in the cache.
// If the key is larger than 65535 or value is larger than 1/1024 of the cache size,
// the entry will not be written to the cache. expireSeconds <= 0 means no expire,
// but it can be evicted when cache is full.
func (f *FreecacheStore) Set(key interface{}, value interface{}, options *Options) error {
	var err error
	var val []byte

	// Using default options set during cache initialization
	if options == nil {
		options = f.options
	}

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
		if tags := options.TagsValue(); len(tags) > 0 {
			f.setTags(key, tags)
		}
		return nil
	}
	return errors.New("key type not supported by Freecache store")
}

func (f *FreecacheStore) setTags(key interface{}, tags []string) {
	for _, tag := range tags {
		var tagKey = fmt.Sprintf(FreecacheTagPattern, tag)
		var cacheKeys = f.getCacheKeysForTag(tagKey)

		var alreadyInserted = false
		for _, cacheKey := range cacheKeys {
			if cacheKey == key.(string) {
				alreadyInserted = true
				break
			}
		}

		if !alreadyInserted {
			cacheKeys = append(cacheKeys, key.(string))
		}

		f.Set(tagKey, []byte(strings.Join(cacheKeys, ",")), &Options{Expiration: 720 * time.Hour})
	}
}

func (f *FreecacheStore) getCacheKeysForTag(tagKey string) []string {
	var cacheKeys = []string{}
	if result, err := f.Get(tagKey); err == nil && result != nil {
		if str, ok := result.([]byte); ok {
			cacheKeys = strings.Split(string(str), ",")
		}
	}
	return cacheKeys
}

// Delete deletes an item in the cache by key and returns err or nil if a delete occurred
func (f *FreecacheStore) Delete(key interface{}) error {
	if v, ok := key.(string); ok {
		if f.client.Del([]byte(v)) {
			return nil
		}
		return fmt.Errorf("failed to delete key %v", key)
	}
	return errors.New("key type not supported by Freecache store")

}

// Invalidate invalidates some cache data in freecache for given options
func (f *FreecacheStore) Invalidate(options InvalidateOptions) error {
	if tags := options.TagsValue(); len(tags) > 0 {
		for _, tag := range tags {
			var tagKey = fmt.Sprintf(FreecacheTagPattern, tag)
			var cacheKeys = f.getCacheKeysForTag(tagKey)

			for _, cacheKey := range cacheKeys {
				err := f.Delete(cacheKey)
				if err != nil {
					return err
				}
			}

			err := f.Delete(tagKey)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Clear resets all data in the store
func (f *FreecacheStore) Clear() error {
	f.client.Clear()
	return nil
}

// GetType returns the store type
func (f *FreecacheStore) GetType() string {
	return FreecacheType
}
