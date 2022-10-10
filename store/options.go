package store

import (
	"time"
)

// Option represents a store option function.
type Option func(o *Options)

type Options struct {
	cost       int64
	expiration time.Duration
	tags       []string
}

func (o *Options) isEmpty() bool {
	return o.cost == 0 && o.expiration == 0 && len(o.tags) == 0
}

func applyOptionsWithDefault(defaultOptions *Options, opts ...Option) *Options {
	returnedOptions := &Options{}
	*returnedOptions = *defaultOptions

	for _, opt := range opts {
		opt(returnedOptions)
	}

	return returnedOptions
}

func applyOptions(opts ...Option) *Options {
	o := &Options{}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// WithCost allows setting the memory capacity used by the item when setting a value.
// Actually it seems to be used by Ristretto library only.
func WithCost(cost int64) Option {
	return func(o *Options) {
		o.cost = cost
	}
}

// WithExpiration allows to specify an expiration time when setting a value.
func WithExpiration(expiration time.Duration) Option {
	return func(o *Options) {
		o.expiration = expiration
	}
}

// WithTags allows to specify associated tags to the current value.
func WithTags(tags []string) Option {
	return func(o *Options) {
		o.tags = tags
	}
}
