package store

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
)

const (
	RistrettoType = "ristretto"
)

// RistrettoStore is a store for Ristretto (memory) library
type RistrettoStore struct {
	client *ristretto.Cache
}

// NewRistretto creates a new store to Ristretto (memory) library instance
func NewRistretto(client *ristretto.Cache) *RistrettoStore {
	return &RistrettoStore{
		client: client,
	}
}

// Get returns data stored from a given key
func (s *RistrettoStore) Get(key interface{}) (interface{}, error) {
	var err error

	value, exists := s.client.Get(key)
	if !exists {
		err = errors.New("Value not found in Ristretto store")
	}

	return value, err
}

// Set defines data in Ristretto memoey cache for given key idntifier
func (s *RistrettoStore) Set(key interface{}, value interface{}, expiration time.Duration) error {
	var err error

	if set := s.client.Set(key, value, 1); !set {
		err = fmt.Errorf("An error has occured while setting value '%v' on key '%v'", value, key)
	}

	return err
}

// GetType returns the store type
func (s *RistrettoStore) GetType() string {
	return RistrettoType
}
