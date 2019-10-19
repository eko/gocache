package codec

import (
	"github.com/eko/gocache/store"
)

// CodecInterface represents an instance of a cache codec
type CodecInterface interface {
	Get(key interface{}) (interface{}, error)
	Set(key interface{}, value interface{}, options *store.Options) error
	Delete(key interface{}) error
	Invalidate(options store.InvalidateOptions) error

	GetStore() store.StoreInterface
	GetStats() *Stats
}
