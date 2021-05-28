package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eko/gocache/store"
	mocksCache "github.com/eko/gocache/test/mocks/cache"
	mocksCodec "github.com/eko/gocache/test/mocks/codec"
	mocksMetrics "github.com/eko/gocache/test/mocks/metrics"
	mocksStore "github.com/eko/gocache/test/mocks/store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewMetric(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	// cmsClient := cmsMocks.NewMockContentClient(ctrl)

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	metrics := mocksMetrics.NewMockMetricsInterface(ctrl)

	// When
	cache := NewMetric(metrics, cache1)

	// Then
	assert.IsType(t, new(MetricCache), cache)

	assert.Equal(t, cache1, cache.cache)
	assert.Equal(t, metrics, cache.metrics)
}

func TestMetricGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	codec1 := mocksCodec.NewMockCodecInterface(ctrl)
	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return(cacheValue, nil)
	cache1.EXPECT().GetCodec().Return(codec1)

	metrics := mocksMetrics.NewMockMetricsInterface(ctrl)
	metrics.EXPECT().RecordFromCodec(codec1).AnyTimes()

	cache := NewMetric(metrics, cache1)

	// When
	value, err := cache.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestMetricGetWhenChainCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store1 := mocksStore.NewMockStoreInterface(ctrl)
	store1.EXPECT().GetType().AnyTimes().Return("store1")

	codec1 := mocksCodec.NewMockCodecInterface(ctrl)
	codec1.EXPECT().GetStore().AnyTimes().Return(store1)

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().GetWithTTL(ctx, "my-key").Return(cacheValue,
		0*time.Second, nil)
	cache1.EXPECT().GetCodec().AnyTimes().Return(codec1)

	chainCache := NewChain(cache1)

	metrics := mocksMetrics.NewMockMetricsInterface(ctrl)
	metrics.EXPECT().RecordFromCodec(codec1).AnyTimes()

	cache := NewMetric(metrics, chainCache)

	// When
	value, err := cache.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestMetricSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	value := &struct {
		Hello string
	}{
		Hello: "world",
	}

	options := &store.Options{
		Expiration: 5 * time.Second,
	}

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Set(ctx, "my-key", value, options).Return(nil)

	metrics := mocksMetrics.NewMockMetricsInterface(ctrl)

	cache := NewMetric(metrics, cache1)

	// When
	err := cache.Set(ctx, "my-key", value, options)

	// Then
	assert.Nil(t, err)
}

func TestMetricDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(nil)

	metrics := mocksMetrics.NewMockMetricsInterface(ctrl)

	cache := NewMetric(metrics, cache1)

	// When
	err := cache.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)
}

func TestMetricDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("Unable to delete key")

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(expectedErr)

	metrics := mocksMetrics.NewMockMetricsInterface(ctrl)

	cache := NewMetric(metrics, cache1)

	// When
	err := cache.Delete(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMetricInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Invalidate(ctx, options).Return(nil)

	metrics := mocksMetrics.NewMockMetricsInterface(ctrl)

	cache := NewMetric(metrics, cache1)

	// When
	err := cache.Invalidate(ctx, options)

	// Then
	assert.Nil(t, err)
}

func TestMetricInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	expectedErr := errors.New("Unexpected error while invalidating data")

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Invalidate(ctx, options).Return(expectedErr)

	metrics := mocksMetrics.NewMockMetricsInterface(ctrl)

	cache := NewMetric(metrics, cache1)

	// When
	err := cache.Invalidate(ctx, options)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMetricClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Clear(ctx).Return(nil)

	metrics := mocksMetrics.NewMockMetricsInterface(ctrl)

	cache := NewMetric(metrics, cache1)

	// When
	err := cache.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestMetricClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("Unexpected error while clearing cache")

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	cache1.EXPECT().Clear(ctx).Return(expectedErr)

	metrics := mocksMetrics.NewMockMetricsInterface(ctrl)

	cache := NewMetric(metrics, cache1)

	// When
	err := cache.Clear(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMetricGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := mocksCache.NewMockSetterCacheInterface(ctrl)
	metrics := mocksMetrics.NewMockMetricsInterface(ctrl)

	cache := NewMetric(metrics, cache1)

	// When - Then
	assert.Equal(t, MetricType, cache.GetType())
}
