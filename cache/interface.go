package cache

import (
	"context"
	"time"

	"github.com/eko/gocache/v3/codec"
	"github.com/eko/gocache/v3/store"
)

// CacheInterface represents the interface for all caches (aggregates, metric, memory, redis, ...)
type CacheInterface[T any] interface {
	Get(ctx context.Context, key any) (T, error)
	Set(ctx context.Context, key any, object T, options *store.Options) error
	Delete(ctx context.Context, key any) error
	Invalidate(ctx context.Context, options store.InvalidateOptions) error
	Clear(ctx context.Context) error
	GetType() string
}

type CacheKeyGenerator interface {
	GetCacheKey() string
}

// SetterCacheInterface represents the interface for caches that allows
// storage (for instance: memory, redis, ...)
type SetterCacheInterface[T any] interface {
	CacheInterface[T]
	GetWithTTL(ctx context.Context, key any) (T, time.Duration, error)

	GetCodec() codec.CodecInterface
}
