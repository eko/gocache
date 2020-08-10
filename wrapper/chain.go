package wrapper

import (
	"log"

	"github.com/yeqown/gocache"

	"github.com/yeqown/gocache/types"

	"github.com/pkg/errors"
)

const (
	// ChainType represents the chain cache type as a string value
	ChainType = "chain"
)

var (
	_ gocache.ICache = &ChainCache{}
)

type chainKeyValue struct {
	key       string
	value     interface{}
	storeType *string
}

// ChainCache represents the configuration needed by a cache aggregator
type ChainCache struct {
	caches     []gocache.ICache
	setChannel chan *chainKeyValue
}

// WrapAsChain instantiate a new cache aggregator
func WrapAsChain(caches ...gocache.ICache) gocache.ICache {
	chain := &ChainCache{
		caches:     caches,
		setChannel: make(chan *chainKeyValue, 10000),
	}

	go chain.setter()

	return chain
}

// setter sets a value in available caches, until a given cache layer
func (c *ChainCache) setter() {
	var multiErr = new(types.MultiError)

	for item := range c.setChannel {
		for _, cc := range c.caches {
			if item.storeType != nil && *item.storeType == cc.GetType() {
				break
			}

			if err := cc.Set(item.key, item.value, nil); err != nil {
				multiErr.Add(err)
			}
		}

		// Log error
		if multiErr != nil && !multiErr.Nil() {
			log.Printf("[Error] chainCache set value failed, err=%s", multiErr)
			multiErr.Reset()
		}
	}
}

// Get returns the object stored in cache if it exists
func (c *ChainCache) Get(key string) (data []byte, err error) {
	for _, cc := range c.caches {
		storeType := cc.GetType()
		data, err = cc.Get(key)
		if err == nil {
			// Set the value back until this cache layer
			c.setChannel <- &chainKeyValue{key, data, &storeType}
			return data, nil
		}

		log.Printf("could not to locate item in store=%s, with data=%v err=%v\n",
			storeType, data, err)
	}

	return nil, err
}

// Set sets a value in available caches
func (c *ChainCache) Set(key string, object interface{}, options *types.StoreOptions) error {
	var multiErr = new(types.MultiError)

	for _, cc := range c.caches {
		if err := cc.Set(key, object, options); err != nil {
			storeType := cc.GetType()
			err = errors.Wrapf(err, "Unable to set item into cache with store %s", storeType)
			multiErr.Add(err)
		}
	}

	if !multiErr.Nil() {
		return multiErr
	}

	return nil
}

// Delete removes a value from all available caches
func (c *ChainCache) Delete(key string) error {
	var multiErr = new(types.MultiError)

	for _, cc := range c.caches {
		if err := cc.Delete(key); err != nil {
			multiErr.Add(err)
		}
	}

	if !multiErr.Nil() {
		return multiErr
	}

	return nil
}

// Invalidate invalidates cache item from given options
func (c *ChainCache) Invalidate(opt types.InvalidateOptions) error {
	var multiErr = new(types.MultiError)

	for _, cc := range c.caches {
		if err := cc.Invalidate(opt); err != nil {
			multiErr.Add(err)
		}
	}

	if !multiErr.Nil() {
		return multiErr
	}

	return nil
}

//
//// Clear resets all cache data
//func (c *ChainCache) Clear() error {
//	var multiErr = new(types.MultiError)
//
//	for _, cc := range c.caches {
//		if err := cc.Clear(); err != nil {
//			multiErr.Add(err)
//		}
//	}
//
//	if !multiErr.Nil() {
//		return multiErr
//	}
//
//	return nil
//}

// GetCaches returns all chained caches
func (c *ChainCache) GetCaches() []gocache.ICache {
	return c.caches
}

// GetType returns the cache type
func (c *ChainCache) GetType() string {
	return ChainType
}
