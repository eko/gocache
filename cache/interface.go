package cache

import (
	"github.com/eko/gocache/codec"
	"github.com/eko/gocache/store"
)

// CacheInterface represents the interface for all caches (aggregates, metric, memory, redis, ...)
type CacheInterface interface {
	Get(key interface{}) (interface{}, error)
	Set(key, object interface{}, options *store.Options) error
	Delete(key interface{}) error
	Invalidate(options store.InvalidateOptions) error
	Clear() error
	GetType() string
}

// SetterCacheInterface represents the interface for caches that allows
// storage (for instance: memory, redis, ...)
type SetterCacheInterface interface {
	CacheInterface

	GetCodec() codec.CodecInterface
}
