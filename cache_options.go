package gocache

import (
	"github.com/yeqown/gocache/types"
)

type Option func(co *cacheOptions)

type cacheOptions struct {
	// InitStoreOpt the default store options
	InitStoreOpt *types.StoreOptions
}

func WithStoreOption(opt *types.StoreOptions) Option {
	return func(co *cacheOptions) {
		co.InitStoreOpt = opt
	}
}
