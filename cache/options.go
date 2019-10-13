package cache

import (
	"time"
)

// Options represents the cache available options
type Options struct {
	Expiration time.Duration
}

// ExpirationValue returns the expiration option value
func (o Options) ExpirationValue() time.Duration {
	return o.Expiration
}
