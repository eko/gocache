package codec

import (
	"github.com/eko/gache/store"
)

// CodecInterface represents an instance of a cache codec
type CodecInterface interface {
	Get(key interface{}) (interface{}, error)
	Set(key interface{}, value interface{}, options *store.Options) error

	GetStore() store.StoreInterface
	GetStats() *Stats
}
