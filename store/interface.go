package store

import (
	"context"
	"time"
)

// StoreInterface is the interface for all available stores
type StoreInterface interface {
	Get(ctx context.Context, key interface{}) (interface{}, error)
	GetWithTTL(ctx context.Context, key interface{}) (interface{}, time.Duration, error)
	Set(ctx context.Context, key interface{}, value interface{}, options *Options) error
	Delete(ctx context.Context, key interface{}) error
	Invalidate(ctx context.Context, options InvalidateOptions) error
	Clear(ctx context.Context) error
	GetType() string
}
