package codec

import (
	"time"

	"github.com/eko/gocache/store"
)

// CodecInterface represents an instance of a cache codec
type CodecInterface interface {
	Get(key interface{}) (interface{}, error)
	GetWithTTL(key interface{}) (interface{}, time.Duration, error)
	Set(key interface{}, value interface{}, options *store.Options) error
	Delete(key interface{}) error
	Invalidate(options store.InvalidateOptions) error
	Clear() error

	GetStore() store.StoreInterface
	GetStats() *Stats
}
