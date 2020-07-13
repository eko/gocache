package cache

import (
	"github.com/yeqown/gocache/store"
)

// ICache represents the interface for all caches (aggregates, metric, memory, redis, ...)
type ICache interface {
	// Get 从cache中获取缓存值
	// 如果是经过序列化的数据，需要通过 returnObj 指定接收值对象
	Get(key interface{}, returnObj interface{}) (interface{}, error)

	// Set .
	Set(key, object interface{}, options *store.Options) error

	// Delete .
	Delete(key interface{}) error

	// Invalidate .
	Invalidate(options store.InvalidateOptions) error

	// Clear .
	Clear() error

	// GetType .
	GetType() string

	GetStats() *Stats
}

//// SetterCacheInterface represents the interface for caches that allows
//// storage (for instance: memory, redis, ...)
//type SetterCacheInterface interface {
//	ICache
//
//	GetCodec() codec.CodecInterface
//}
