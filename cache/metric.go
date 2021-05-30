package cache

import (
	"context"

	"github.com/eko/gocache/v2/metrics"
	"github.com/eko/gocache/v2/store"
)

const (
	// MetricType represents the metric cache type as a string value
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
func (c *MetricCache) Get(ctx context.Context, key interface{}) (interface{}, error) {
	result, err := c.cache.Get(ctx, key)

	c.updateMetrics(c.cache)

	return result, err
}

// Set sets a value from the cache
func (c *MetricCache) Set(ctx context.Context, key, object interface{}, options *store.Options) error {
	return c.cache.Set(ctx, key, object, options)
}

// Delete removes a value from the cache
func (c *MetricCache) Delete(ctx context.Context, key interface{}) error {
	return c.cache.Delete(ctx, key)
}

// Invalidate invalidates cache item from given options
func (c *MetricCache) Invalidate(ctx context.Context, options store.InvalidateOptions) error {
	return c.cache.Invalidate(ctx, options)
}

// Clear resets all cache data
func (c *MetricCache) Clear(ctx context.Context) error {
	return c.cache.Clear(ctx)
}

// Get obtains a value from cache and also records metrics
func (c *MetricCache) updateMetrics(cache CacheInterface) {
	switch current := cache.(type) {
	case *ChainCache:
		for _, cache := range current.GetCaches() {
			c.updateMetrics(cache)
		}

	case SetterCacheInterface:
		c.metrics.RecordFromCodec(current.GetCodec())
	}
}

// GetType returns the cache type
func (c *MetricCache) GetType() string {
	return MetricType
}
