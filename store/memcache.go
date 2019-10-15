package store

import (
	"errors"

	"github.com/bradfitz/gomemcache/memcache"
)

// MemcacheClientInterface represents a bradfitz/gomemcache client
type MemcacheClientInterface interface {
	Get(key string) (item *memcache.Item, err error)
	Set(item *memcache.Item) error
}

const (
	MemcacheType = "memcache"
)

// MemcacheStore is a store for Redis
type MemcacheStore struct {
	client MemcacheClientInterface
	options *Options
}

// NewMemcache creates a new store to Memcache instance(s)
func NewMemcache(client MemcacheClientInterface, options *Options) *MemcacheStore {
	if options == nil {
		options = &Options{}
	}

	return &MemcacheStore{
		client: client,
		options: options,
	}
}

// Get returns data stored from a given key
func (s *MemcacheStore) Get(key interface{}) (interface{}, error) {
	item, err := s.client.Get(key.(string))
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("Unable to retrieve data from memcache")
	}

	return item.Value, err
}

// Set defines data in Redis for given key idntifier
func (s *MemcacheStore) Set(key interface{}, value interface{}, options *Options) error {
	if options == nil {
		options = s.options
	}

	item := &memcache.Item{
		Key:        key.(string),
		Value:      value.([]byte),
		Expiration: int32(options.ExpirationValue().Seconds()),
	}

	return s.client.Set(item)
}

// GetType returns the store type
func (s *MemcacheStore) GetType() string {
	return MemcacheType
}
