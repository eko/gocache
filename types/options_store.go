package types

import (
	"time"
)

// StoreOptions represents the cache store available options
type StoreOptions struct {
	// Expiration allows to specify an expiration time when setting a value
	Expiration time.Duration

	// Tags allows to specify associated tags to the current value
	Tags []string
}

// ExpirationValue returns the expiration option value
func (o StoreOptions) ExpirationValue() time.Duration {
	return o.Expiration
}

// TagsValue returns the tags option value
func (o StoreOptions) TagsValue() []string {
	return o.Tags
}
