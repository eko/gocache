package codec

import (
	"time"

	"github.com/eko/gache/store"
)

// CodecInterface represents an instance of a cache codec
type CodecInterface interface {
	Get(key interface{}) (interface{}, error)
	Set(key interface{}, value interface{}, expiration time.Duration) error

	GetStore() store.StoreInterface
	GetStats() *Stats
}
