package cache

import (
	"errors"
	"testing"
	"time"

	"github.com/eko/gocache/store"
	mocksCache "github.com/eko/gocache/test/mocks/cache"
	mocksCodec "github.com/eko/gocache/test/mocks/codec"
	mocksMetrics "github.com/eko/gocache/test/mocks/metrics"
	mocksStore "github.com/eko/gocache/test/mocks/store"
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

	chainCache := NewChain(cache1)

	metrics := &mocksMetrics.MetricsInterface{}
	metrics.On("RecordFromCodec", codec1)

	cache := NewMetric(metrics, chainCache)

	// When
	value, err := cache.Get("my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestMetricSet(t *testing.T) {
	// Given
	value := &struct {
		Hello string
	}{
		Hello: "world",
	}

	options := &store.Options{
		Expiration: 5 * time.Second,
	}

	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Set", "my-key", value, options).Return(nil)

	metrics := &mocksMetrics.MetricsInterface{}

	cache := NewMetric(metrics, cache1)

	// When
	err := cache.Set("my-key", value, options)

	// Then
	assert.Nil(t, err)
}

func TestMetricDelete(t *testing.T) {
	// Given
	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Delete", "my-key").Return(nil)

	metrics := &mocksMetrics.MetricsInterface{}

	cache := NewMetric(metrics, cache1)

	// When
	err := cache.Delete("my-key")

	// Then
	assert.Nil(t, err)
}

func TestMetricDeleteWhenError(t *testing.T) {
	// Given
	expectedErr := errors.New("Unable to delete key")

	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Delete", "my-key").Return(expectedErr)

	metrics := &mocksMetrics.MetricsInterface{}

	cache := NewMetric(metrics, cache1)

	// When
	err := cache.Delete("my-key")

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMetricInvalidate(t *testing.T) {
	// Given
	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Invalidate", options).Return(nil)

	metrics := &mocksMetrics.MetricsInterface{}

	cache := NewMetric(metrics, cache1)

	// When
	err := cache.Invalidate(options)

	// Then
	assert.Nil(t, err)
}

func TestMetricInvalidateWhenError(t *testing.T) {
	// Given
	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	expectedErr := errors.New("Unexpected error while invalidating data")

	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Invalidate", options).Return(expectedErr)

	metrics := &mocksMetrics.MetricsInterface{}

	cache := NewMetric(metrics, cache1)

	// When
	err := cache.Invalidate(options)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMetricClear(t *testing.T) {
	// Given
	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Clear").Return(nil)

	metrics := &mocksMetrics.MetricsInterface{}

	cache := NewMetric(metrics, cache1)

	// When
	err := cache.Clear()

	// Then
	assert.Nil(t, err)
}

func TestMetricClearWhenError(t *testing.T) {
	// Given
	expectedErr := errors.New("Unexpected error while clearing cache")

	cache1 := &mocksCache.SetterCacheInterface{}
	cache1.On("Clear").Return(expectedErr)

	metrics := &mocksMetrics.MetricsInterface{}

	cache := NewMetric(metrics, cache1)

	// When
	err := cache.Clear()

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMetricGetType(t *testing.T) {
	// Given
	cache1 := &mocksCache.SetterCacheInterface{}
	metrics := &mocksMetrics.MetricsInterface{}

	cache := NewMetric(metrics, cache1)

	// When - Then
	assert.Equal(t, MetricType, cache.GetType())
}
