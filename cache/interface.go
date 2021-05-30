package cache

import (
	"context"
	"time"

	"github.com/eko/gocache/v2/codec"
	"github.com/eko/gocache/v2/store"
)

// CacheInterface represents the interface for all caches (aggregates, metric, memory, redis, ...)
type CacheInterface interface {
	Get(ctx context.Context, key interface{}) (interface{}, error)
	Set(ctx context.Context, key, object interface{}, options *store.Options) error
	Delete(ctx context.Context, key interface{}) error
	Invalidate(ctx context.Context, options store.InvalidateOptions) error
	Clear(ctx context.Context) error
	GetType() string
}

// SetterCacheInterface represents the interface for caches that allows
// storage (for instance: memory, redis, ...)
type SetterCacheInterface interface {
	CacheInterface
	GetWithTTL(ctx context.Context, key interface{}) (interface{}, time.Duration, error)

	GetCodec() codec.CodecInterface
}
