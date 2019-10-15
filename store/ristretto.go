package store

import (
	"errors"
	"fmt"
)

const (
	RistrettoType = "ristretto"
)

// RistrettoClientInterface represents a dgraph-io/ristretto client
type RistrettoClientInterface interface {
	Get(key interface{}) (interface{}, bool)
	Set(key, value interface{}, cost int64) bool
}

// RistrettoStore is a store for Ristretto (memory) library
type RistrettoStore struct {
	client RistrettoClientInterface
	options *Options
}

// NewRistretto creates a new store to Ristretto (memory) library instance
func NewRistretto(client RistrettoClientInterface, options *Options) *RistrettoStore {
	if options == nil {
		options = &Options{}
	}

	return &RistrettoStore{
		client: client,
		options: options,
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
func (s *RistrettoStore) Set(key interface{}, value interface{}, options *Options) error {
	var err error

	if options == nil {
		options = s.options
	}

	if set := s.client.Set(key, value, options.CostValue()); !set {
		err = fmt.Errorf("An error has occured while setting value '%v' on key '%v'", value, key)
	}

	return err
}

// GetType returns the store type
func (s *RistrettoStore) GetType() string {
	return RistrettoType
}
