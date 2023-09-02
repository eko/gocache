package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eko/gocache/lib/v4/codec"
	"github.com/eko/gocache/lib/v4/metrics"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewMetric(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	codec1 := codec.NewMockCodecInterface(ctrl)
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().GetCodec().Return(codec1).Times(1)
	metrics := metrics.NewMockMetricsInterface(ctrl)
	metrics.EXPECT().RecordFromCodec(codec1).Times(1)

	// When
	cache := NewMetric[any](metrics, cache1)

	// Then
	assert.IsType(t, new(MetricCache[any]), cache)

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

	codec1 := codec.NewMockCodecInterface(ctrl)
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return(cacheValue, nil)
	cache1.EXPECT().GetCodec().Return(codec1).MinTimes(1)

	metrics := metrics.NewMockMetricsInterface(ctrl)
	metrics.EXPECT().RecordFromCodec(codec1).MinTimes(1)

	cache := NewMetric[any](metrics, cache1)

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

	store1 := store.NewMockStoreInterface(ctrl)
	store1.EXPECT().GetType().AnyTimes().Return("store1")

	codec1 := codec.NewMockCodecInterface(ctrl)
	codec1.EXPECT().GetStore().AnyTimes().Return(store1)

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().GetWithTTL(ctx, "my-key").Return(cacheValue,
		0*time.Second, nil)
	cache1.EXPECT().GetCodec().AnyTimes().Return(codec1)

	chainCache := NewChain[any](cache1)

	metrics := metrics.NewMockMetricsInterface(ctrl)
	metrics.EXPECT().RecordFromCodec(codec1).AnyTimes()

	cache := NewMetric[any](metrics, chainCache)

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

	codec1 := codec.NewMockCodecInterface(ctrl)
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Set(ctx, "my-key", value).Return(nil)
	cache1.EXPECT().GetCodec().Return(codec1).Times(1)

	metrics := metrics.NewMockMetricsInterface(ctrl)
	metrics.EXPECT().RecordFromCodec(codec1).Times(1)

	cache := NewMetric[any](metrics, cache1)

	// When
	err := cache.Set(ctx, "my-key", value)

	// Then
	assert.Nil(t, err)
}

func TestMetricDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	codec1 := codec.NewMockCodecInterface(ctrl)
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(nil)
	cache1.EXPECT().GetCodec().Return(codec1).Times(1)

	metrics := metrics.NewMockMetricsInterface(ctrl)
	metrics.EXPECT().RecordFromCodec(codec1).Times(1)

	cache := NewMetric[any](metrics, cache1)

	// When
	err := cache.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)
}

func TestMetricDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unable to delete key")

	codec1 := codec.NewMockCodecInterface(ctrl)
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(expectedErr)
	cache1.EXPECT().GetCodec().Return(codec1).Times(1)

	metrics := metrics.NewMockMetricsInterface(ctrl)
	metrics.EXPECT().RecordFromCodec(codec1).Times(1)

	cache := NewMetric[any](metrics, cache1)

	// When
	err := cache.Delete(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMetricInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	codec1 := codec.NewMockCodecInterface(ctrl)
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Invalidate(ctx).Return(nil)
	cache1.EXPECT().GetCodec().Return(codec1).Times(1)

	metrics := metrics.NewMockMetricsInterface(ctrl)
	metrics.EXPECT().RecordFromCodec(codec1).Times(1)

	cache := NewMetric[any](metrics, cache1)

	// When
	err := cache.Invalidate(ctx)

	// Then
	assert.Nil(t, err)
}

func TestMetricInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unexpected error while invalidating data")

	codec1 := codec.NewMockCodecInterface(ctrl)
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Invalidate(ctx).Return(expectedErr)
	cache1.EXPECT().GetCodec().Return(codec1).Times(1)

	metrics := metrics.NewMockMetricsInterface(ctrl)
	metrics.EXPECT().RecordFromCodec(codec1).Times(1)

	cache := NewMetric[any](metrics, cache1)

	// When
	err := cache.Invalidate(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMetricClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	codec1 := codec.NewMockCodecInterface(ctrl)
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Clear(ctx).Return(nil)
	cache1.EXPECT().GetCodec().Return(codec1).Times(1)

	metrics := metrics.NewMockMetricsInterface(ctrl)
	metrics.EXPECT().RecordFromCodec(codec1).Times(1)

	cache := NewMetric[any](metrics, cache1)

	// When
	err := cache.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestMetricClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unexpected error while clearing cache")

	codec1 := codec.NewMockCodecInterface(ctrl)
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Clear(ctx).Return(expectedErr)
	cache1.EXPECT().GetCodec().Return(codec1).Times(1)

	metrics := metrics.NewMockMetricsInterface(ctrl)
	metrics.EXPECT().RecordFromCodec(codec1).Times(1)

	cache := NewMetric[any](metrics, cache1)

	// When
	err := cache.Clear(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMetricGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	codec1 := codec.NewMockCodecInterface(ctrl)
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().GetCodec().Return(codec1).Times(1)
	metrics := metrics.NewMockMetricsInterface(ctrl)
	metrics.EXPECT().RecordFromCodec(codec1).Times(1)

	cache := NewMetric[any](metrics, cache1)

	// When - Then
	assert.Equal(t, MetricType, cache.GetType())
}
