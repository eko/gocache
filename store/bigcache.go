package store

import (
	"errors"
)

// BigcacheClientInterface represents a allegro/bigcache client
type BigcacheClientInterface interface {
	Get(key string) ([]byte, error)
	Set(key string, entry []byte) error
	Delete(key string) error
}

const (
	BigcacheType = "bigcache"
)

// BigcacheStore is a store for Redis
type BigcacheStore struct {
	client  BigcacheClientInterface
	options *Options
}

// NewBigcache creates a new store to Bigcache instance(s)
func NewBigcache(client BigcacheClientInterface, options *Options) *BigcacheStore {
	if options == nil {
		options = &Options{}
	}

	return &BigcacheStore{
		client:  client,
		options: options,
	}
}

// Get returns data stored from a given key
func (s *BigcacheStore) Get(key interface{}) (interface{}, error) {
	item, err := s.client.Get(key.(string))
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("Unable to retrieve data from bigcache")
	}

	return item, err
}

// Set defines data in Redis for given key idntifier
func (s *BigcacheStore) Set(key interface{}, value interface{}, options *Options) error {
	return s.client.Set(key.(string), value.([]byte))
}

// Delete removes data from Redis for given key idntifier
func (s *BigcacheStore) Delete(key interface{}) error {
	return s.client.Delete(key.(string))
}

// GetType returns the store type
func (s *BigcacheStore) GetType() string {
	return BigcacheType
}
