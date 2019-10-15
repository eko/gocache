package cache

import (
	"github.com/eko/gache/codec"
	"github.com/eko/gache/store"
)

// CacheInterface represents the interface for all caches (aggregates, metric, memory, redis, ...)
type CacheInterface interface {
	Get(key interface{}) (interface{}, error)
	Set(key, object interface{}, options *store.Options) error
	GetType() string
}

// SetterCacheInterface represents the interface for caches that allows
// storage (for instance: memory, redis, ...)
type SetterCacheInterface interface {
	CacheInterface

	GetCodec() codec.CodecInterface
}
