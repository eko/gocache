package store

import (
	"context"
	"time"
)

// StoreInterface is the interface for all available stores
type StoreInterface interface {
	Get(ctx context.Context, key any) (any, error)
	GetWithTTL(ctx context.Context, key any) (any, time.Duration, error)
	Set(ctx context.Context, key any, value any, options ...Option) error
	Delete(ctx context.Context, key any) error
	Invalidate(ctx context.Context, options ...InvalidateOption) error
	Clear(ctx context.Context) error
	GetType() string
}

// SetIfNotExistsStore is implemented by stores able to atomically write a value
// only when the key does not exist yet.
//
// It is an optional interface: stores whose client has no such primitive do not
// implement it, and cache.Cache returns ErrNotSupported for them.
type SetIfNotExistsStore interface {
	StoreInterface

	// SetIfNotExists sets the value for the given key only when it does not
	// already exist, and reports whether it has been written.
	SetIfNotExists(ctx context.Context, key any, value any, options ...Option) (bool, error)
}
