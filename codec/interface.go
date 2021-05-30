package codec

import (
	"context"
	"time"

	"github.com/eko/gocache/v2/store"
)

// CodecInterface represents an instance of a cache codec
type CodecInterface interface {
	Get(ctx context.Context, key interface{}) (interface{}, error)
	GetWithTTL(ctx context.Context, key interface{}) (interface{}, time.Duration, error)
	Set(ctx context.Context, key interface{}, value interface{}, options *store.Options) error
	Delete(ctx context.Context, key interface{}) error
	Invalidate(ctx context.Context, options store.InvalidateOptions) error
	Clear(ctx context.Context) error

	GetStore() store.StoreInterface
	GetStats() *Stats
}
