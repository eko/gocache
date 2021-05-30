package codec

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eko/gocache/v2/store"
	mocksStore "github.com/eko/gocache/v2/test/mocks/store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	store := mocksStore.NewMockStoreInterface(ctrl)

	// When
	codec := New(store)

	// Then
	assert.IsType(t, new(Codec), codec)
}

func TestGetWhenHit(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Get(ctx, "my-key").Return(cacheValue, nil)

	codec := New(store)

	// When
	value, err := codec.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)

	assert.Equal(t, 1, codec.GetStats().Hits)
	assert.Equal(t, 0, codec.GetStats().Miss)
	assert.Equal(t, 0, codec.GetStats().SetSuccess)
	assert.Equal(t, 0, codec.GetStats().SetError)
	assert.Equal(t, 0, codec.GetStats().DeleteSuccess)
	assert.Equal(t, 0, codec.GetStats().DeleteError)
	assert.Equal(t, 0, codec.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, codec.GetStats().InvalidateError)
	assert.Equal(t, 0, codec.GetStats().ClearSuccess)
	assert.Equal(t, 0, codec.GetStats().ClearError)
}

func TestGetWithTTLWhenHit(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().GetWithTTL(ctx, "my-key").Return(cacheValue, 1*time.Second, nil)

	codec := New(store)

	// When
	value, ttl, err := codec.GetWithTTL(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
	assert.Equal(t, 1*time.Second, ttl)

	assert.Equal(t, 1, codec.GetStats().Hits)
	assert.Equal(t, 0, codec.GetStats().Miss)
	assert.Equal(t, 0, codec.GetStats().SetSuccess)
	assert.Equal(t, 0, codec.GetStats().SetError)
	assert.Equal(t, 0, codec.GetStats().DeleteSuccess)
	assert.Equal(t, 0, codec.GetStats().DeleteError)
	assert.Equal(t, 0, codec.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, codec.GetStats().InvalidateError)
	assert.Equal(t, 0, codec.GetStats().ClearSuccess)
	assert.Equal(t, 0, codec.GetStats().ClearError)
}

func TestGetWithTTLWhenMiss(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("Unable to find in store")

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().GetWithTTL(ctx, "my-key").Return(nil, 0*time.Second, expectedErr)

	codec := New(store)

	// When
	value, ttl, err := codec.GetWithTTL(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, value)
	assert.Equal(t, 0*time.Second, ttl)

	assert.Equal(t, 0, codec.GetStats().Hits)
	assert.Equal(t, 1, codec.GetStats().Miss)
	assert.Equal(t, 0, codec.GetStats().SetSuccess)
	assert.Equal(t, 0, codec.GetStats().SetError)
	assert.Equal(t, 0, codec.GetStats().DeleteSuccess)
	assert.Equal(t, 0, codec.GetStats().DeleteError)
	assert.Equal(t, 0, codec.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, codec.GetStats().InvalidateError)
	assert.Equal(t, 0, codec.GetStats().ClearSuccess)
	assert.Equal(t, 0, codec.GetStats().ClearError)
}

func TestGetWhenMiss(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("Unable to find in store")

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Get(ctx, "my-key").Return(nil, expectedErr)

	codec := New(store)

	// When
	value, err := codec.Get(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, value)

	assert.Equal(t, 0, codec.GetStats().Hits)
	assert.Equal(t, 1, codec.GetStats().Miss)
	assert.Equal(t, 0, codec.GetStats().SetSuccess)
	assert.Equal(t, 0, codec.GetStats().SetError)
	assert.Equal(t, 0, codec.GetStats().DeleteSuccess)
	assert.Equal(t, 0, codec.GetStats().DeleteError)
	assert.Equal(t, 0, codec.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, codec.GetStats().InvalidateError)
	assert.Equal(t, 0, codec.GetStats().ClearSuccess)
	assert.Equal(t, 0, codec.GetStats().ClearError)
}

func TestSetWhenSuccess(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	options := &store.Options{
		Expiration: 5 * time.Second,
	}

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Set(ctx, "my-key", cacheValue, options).Return(nil)

	codec := New(store)

	// When
	err := codec.Set(ctx, "my-key", cacheValue, options)

	// Then
	assert.Nil(t, err)

	assert.Equal(t, 0, codec.GetStats().Hits)
	assert.Equal(t, 0, codec.GetStats().Miss)
	assert.Equal(t, 1, codec.GetStats().SetSuccess)
	assert.Equal(t, 0, codec.GetStats().SetError)
	assert.Equal(t, 0, codec.GetStats().DeleteSuccess)
	assert.Equal(t, 0, codec.GetStats().DeleteError)
	assert.Equal(t, 0, codec.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, codec.GetStats().InvalidateError)
	assert.Equal(t, 0, codec.GetStats().ClearSuccess)
	assert.Equal(t, 0, codec.GetStats().ClearError)
}

func TestSetWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	options := &store.Options{
		Expiration: 5 * time.Second,
	}

	expectedErr := errors.New("Unable to set value in store")

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Set(ctx, "my-key", cacheValue, options).Return(expectedErr)

	codec := New(store)

	// When
	err := codec.Set(ctx, "my-key", cacheValue, options)

	// Then
	assert.Equal(t, expectedErr, err)

	assert.Equal(t, 0, codec.GetStats().Hits)
	assert.Equal(t, 0, codec.GetStats().Miss)
	assert.Equal(t, 0, codec.GetStats().SetSuccess)
	assert.Equal(t, 1, codec.GetStats().SetError)
	assert.Equal(t, 0, codec.GetStats().DeleteSuccess)
	assert.Equal(t, 0, codec.GetStats().DeleteError)
	assert.Equal(t, 0, codec.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, codec.GetStats().InvalidateError)
	assert.Equal(t, 0, codec.GetStats().ClearSuccess)
	assert.Equal(t, 0, codec.GetStats().ClearError)
}

func TestDeleteWhenSuccess(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Delete(ctx, "my-key").Return(nil)

	codec := New(store)

	// When
	err := codec.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)

	assert.Equal(t, 0, codec.GetStats().Hits)
	assert.Equal(t, 0, codec.GetStats().Miss)
	assert.Equal(t, 0, codec.GetStats().SetSuccess)
	assert.Equal(t, 0, codec.GetStats().SetError)
	assert.Equal(t, 1, codec.GetStats().DeleteSuccess)
	assert.Equal(t, 0, codec.GetStats().DeleteError)
	assert.Equal(t, 0, codec.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, codec.GetStats().InvalidateError)
	assert.Equal(t, 0, codec.GetStats().ClearSuccess)
	assert.Equal(t, 0, codec.GetStats().ClearError)
}

func TesDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("Unable to delete key")

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Delete(ctx, "my-key").Return(expectedErr)

	codec := New(store)

	// When
	err := codec.Delete(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)

	assert.Equal(t, 0, codec.GetStats().Hits)
	assert.Equal(t, 0, codec.GetStats().Miss)
	assert.Equal(t, 0, codec.GetStats().SetSuccess)
	assert.Equal(t, 0, codec.GetStats().SetError)
	assert.Equal(t, 0, codec.GetStats().DeleteSuccess)
	assert.Equal(t, 1, codec.GetStats().DeleteError)
	assert.Equal(t, 0, codec.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, codec.GetStats().InvalidateError)
	assert.Equal(t, 0, codec.GetStats().ClearSuccess)
	assert.Equal(t, 0, codec.GetStats().ClearError)
}

func TestInvalidateWhenSuccess(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Invalidate(ctx, options).Return(nil)

	codec := New(store)

	// When
	err := codec.Invalidate(ctx, options)

	// Then
	assert.Nil(t, err)

	assert.Equal(t, 0, codec.GetStats().Hits)
	assert.Equal(t, 0, codec.GetStats().Miss)
	assert.Equal(t, 0, codec.GetStats().SetSuccess)
	assert.Equal(t, 0, codec.GetStats().SetError)
	assert.Equal(t, 0, codec.GetStats().DeleteSuccess)
	assert.Equal(t, 0, codec.GetStats().DeleteError)
	assert.Equal(t, 1, codec.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, codec.GetStats().InvalidateError)
	assert.Equal(t, 0, codec.GetStats().ClearSuccess)
	assert.Equal(t, 0, codec.GetStats().ClearError)
}

func TestInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	expectedErr := errors.New("Unexpected error when invalidating data")

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Invalidate(ctx, options).Return(expectedErr)

	codec := New(store)

	// When
	err := codec.Invalidate(ctx, options)

	// Then
	assert.Equal(t, expectedErr, err)

	assert.Equal(t, 0, codec.GetStats().Hits)
	assert.Equal(t, 0, codec.GetStats().Miss)
	assert.Equal(t, 0, codec.GetStats().SetSuccess)
	assert.Equal(t, 0, codec.GetStats().SetError)
	assert.Equal(t, 0, codec.GetStats().DeleteSuccess)
	assert.Equal(t, 0, codec.GetStats().DeleteError)
	assert.Equal(t, 0, codec.GetStats().InvalidateSuccess)
	assert.Equal(t, 1, codec.GetStats().InvalidateError)
	assert.Equal(t, 0, codec.GetStats().ClearSuccess)
	assert.Equal(t, 0, codec.GetStats().ClearError)
}

func TestClearWhenSuccess(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Clear(ctx).Return(nil)

	codec := New(store)

	// When
	err := codec.Clear(ctx)

	// Then
	assert.Nil(t, err)

	assert.Equal(t, 0, codec.GetStats().Hits)
	assert.Equal(t, 0, codec.GetStats().Miss)
	assert.Equal(t, 0, codec.GetStats().SetSuccess)
	assert.Equal(t, 0, codec.GetStats().SetError)
	assert.Equal(t, 0, codec.GetStats().DeleteSuccess)
	assert.Equal(t, 0, codec.GetStats().DeleteError)
	assert.Equal(t, 0, codec.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, codec.GetStats().InvalidateError)
	assert.Equal(t, 1, codec.GetStats().ClearSuccess)
	assert.Equal(t, 0, codec.GetStats().ClearError)
}

func TestClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("Unexpected error when clearing cache")

	store := mocksStore.NewMockStoreInterface(ctrl)
	store.EXPECT().Clear(ctx).Return(expectedErr)

	codec := New(store)

	// When
	err := codec.Clear(ctx)

	// Then
	assert.Equal(t, expectedErr, err)

	assert.Equal(t, 0, codec.GetStats().Hits)
	assert.Equal(t, 0, codec.GetStats().Miss)
	assert.Equal(t, 0, codec.GetStats().SetSuccess)
	assert.Equal(t, 0, codec.GetStats().SetError)
	assert.Equal(t, 0, codec.GetStats().DeleteSuccess)
	assert.Equal(t, 0, codec.GetStats().DeleteError)
	assert.Equal(t, 0, codec.GetStats().InvalidateSuccess)
	assert.Equal(t, 0, codec.GetStats().InvalidateError)
	assert.Equal(t, 0, codec.GetStats().ClearSuccess)
	assert.Equal(t, 1, codec.GetStats().ClearError)
}

func TestGetStore(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	store := mocksStore.NewMockStoreInterface(ctrl)

	codec := New(store)

	// When - Then
	assert.Equal(t, store, codec.GetStore())
}

func TestGetStats(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	store := mocksStore.NewMockStoreInterface(ctrl)

	codec := New(store)

	// When - Then
	expectedStats := &Stats{}
	assert.Equal(t, expectedStats, codec.GetStats())
}
