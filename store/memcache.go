package store

import (
	"errors"
	"time"

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
}

// NewMemcache creates a new store to Memcache instance(s)
func NewMemcache(client MemcacheClientInterface) *MemcacheStore {
	return &MemcacheStore{
		client: client,
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
func (s *MemcacheStore) Set(key interface{}, value interface{}, expiration time.Duration) error {
	item := &memcache.Item{
		Key:        key.(string),
		Value:      value.([]byte),
		Expiration: int32(expiration.Seconds()),
	}

	return s.client.Set(item)
}

// GetType returns the store type
func (s *MemcacheStore) GetType() string {
	return MemcacheType
}
