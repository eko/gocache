package store

import (
	"time"
)

// StoreInterface is the interface for all available stores
type StoreInterface interface {
	Get(key interface{}) (interface{}, error)
	Set(key interface{}, value interface{}, expiration time.Duration) error
	GetType() string
}
