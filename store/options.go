package store

import (
	"context"
	"time"
)

// Options represents the cache store available options
type Options struct {
	// Cost corresponds to the memory capacity used by the item when setting a value
	// Actually it seems to be used by Ristretto library only
	Cost int64

	// Expiration allows to specify an expiration time when setting a value
	Expiration time.Duration

	// Tags allows to specify associated tags to the current value
	Tags []string

	// Ctx pass context for control timeout for all operations
	Ctx context.Context
}

// CostValue returns the allocated memory capacity
func (o Options) CostValue() int64 {
	return o.Cost
}

// ExpirationValue returns the expiration option value
func (o Options) ExpirationValue() time.Duration {
	return o.Expiration
}

// TagsValue returns the tags option value
func (o Options) TagsValue() []string {
	return o.Tags
}

// CtxValue returns the ctx option value
func (o Options) CtxValue() context.Context {
	return o.Ctx
}
