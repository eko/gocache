package cache

import (
	"github.com/eko/gocache/metrics"
	"github.com/eko/gocache/store"
)

const (
	MetricType = "metric"
)

// MetricCache is the struct that specifies metrics available for different caches
type MetricCache struct {
	metrics metrics.MetricsInterface
	cache   CacheInterface
}

// NewMetric creates a new cache with metrics and a given cache storage
func NewMetric(metrics metrics.MetricsInterface, cache CacheInterface) *MetricCache {
	return &MetricCache{
		metrics: metrics,
		cache:   cache,
	}
}

// Get obtains a value from cache and also records metrics
func (c *MetricCache) Get(key interface{}) (interface{}, error) {
	result, err := c.cache.Get(key)

	c.updateMetrics(c.cache)

	return result, err
}

// Set sets a value in cache and also records set metric
func (c *MetricCache) Set(key, object interface{}, options *store.Options) error {
	return c.cache.Set(key, object, options)
}

// Get obtains a value from cache and also records metrics
func (c *MetricCache) updateMetrics(cache CacheInterface) {
	switch current := cache.(type) {
	case *ChainCache:
		for _, cache := range current.GetCaches() {
			c.updateMetrics(cache)
		}

	case SetterCacheInterface:
		go c.metrics.RecordFromCodec(current.GetCodec())
	}
}

// GetType returns the cache type
func (c *MetricCache) GetType() string {
	return MetricType
}
