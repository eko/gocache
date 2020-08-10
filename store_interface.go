package gocache

import "github.com/yeqown/gocache/types"

// IStore is the interface for all available stores
type IStore interface {
	// Get .
	Get(key string) ([]byte, error)

	// Set .
	Set(key string, value interface{}, options *types.StoreOptions) error

	// Delete .
	Delete(key string) error

	// Invalidate .
	Invalidate(options types.InvalidateOptions) error

	// Clear .
	Clear() error

	// GetType .
	GetType() string
}

const (
	_tagPrefix = "gocache_tag_"
)

func GenTagKey(tag string) string {
	return _tagPrefix + tag
}
