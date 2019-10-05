package cache

import (
	"testing"

	mocksCache "github.com/eko/gache/test/mocks/cache"
	mocksCodec "github.com/eko/gache/test/mocks/codec"
	mocksMetrics "github.com/eko/gache/test/mocks/metrics"
	mocksStore "github.com/eko/gache/test/mocks/store"
	"github.com/stretchr/testify/assert"
)

func TestNewMetric(t *testing.T) {
	// Given
	cache1 := &mocksCache.SetterCacheInterface{}
	metrics := &mocksMetrics.MetricsInterface{}

	// When
	cache := NewMetric(metrics, cache1)

	// Then
	assert.IsType(t, new(MetricCache), cache)

	assert.Equal(t, cache1, cache.cache)
	assert.Equal(t, metrics, cache.metrics)
}

func TestMetricGet(t *testing.T) {
	// Given
	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	codec1 := &mocksCodec.CodecInterface{}
	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Get", "my-key").Return(cacheValue, nil)
	cache1.On("GetCodec").Return(codec1)

	metrics := &mocksMetrics.MetricsInterface{}
	metrics.On("RecordFromCodec", codec1)

	cache := NewMetric(metrics, cache1)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestMetricGetWhenChainCache(t *testing.T) {
	// Given
	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store1 := &mocksStore.StoreInterface{}
	store1.On("GetType").Return("store1")

	codec1 := &mocksCodec.CodecInterface{}
	codec1.On("GetStore").Return(store1)

	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Get", "my-key").Return(cacheValue, nil)
	cache1.On("GetCodec").Return(codec1)

	loadFunc := func(key interface{}) (interface{}, error) {
		return cacheValue, nil
	}

	chainCache := NewChain(loadFunc, cache1)

	metrics := &mocksMetrics.MetricsInterface{}
	metrics.On("RecordFromCodec", codec1)

	cache := NewMetric(metrics, chainCache)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestMetricGetType(t *testing.T) {
	// Given
	cache1 := &mocksCache.SetterCacheInterface{}
	metrics := &mocksMetrics.MetricsInterface{}

	cache := NewMetric(metrics, cache1)

	// When - Then
	assert.Equal(t, MetricType, cache.GetType())
}
