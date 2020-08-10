package gocache

import (
	"github.com/yeqown/gocache/types"

	"github.com/pkg/errors"
)

const (
	// CacheType represents the cache type as a string value
	PureCacheType = "pure-cache"
)

var (
	_ ICache = &cache{}

	// ErrEmptyStore .
	ErrEmptyStore = errors.New("store is empty")
)

// cache is a pure component which implements ICache with store.
type cache struct {
	store IStore
}

// New construct an ICache with `s` and `otps`
func New(s IStore, opts ...Option) (ICache, error) {
	if s == nil {
		return nil, ErrEmptyStore
	}

	// TODO: use with co
	var co = new(cacheOptions)

	// apply all `opt` to `co`
	for _, opt := range opts {
		opt(co)
	}

	return &cache{
		store: s,
	}, nil
}

func (c *cache) Set(key string, object interface{}, options *types.StoreOptions) error {
	return c.store.Set(key, object, options)
}                                                             // insert or update
func (c *cache) Get(key string) ([]byte, error)               { return c.store.Get(key) }        // query
func (c *cache) Invalidate(opt types.InvalidateOptions) error { return c.store.Invalidate(opt) } // invalidate multi keys
func (c *cache) Delete(key string) error                      { return c.store.Delete(key) }     // delete
func (c *cache) Clear() error                                 { return c.store.Clear() }         // clear
func (c *cache) GetType() string                              { return PureCacheType }           //  indicate the store which is used
