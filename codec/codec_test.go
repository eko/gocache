package codec

import (
	"errors"
	"testing"
	"time"

	"github.com/eko/gocache/store"
	mocksStore "github.com/eko/gocache/test/mocks/store"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Given
	store := &mocksStore.StoreInterface{}

	// When
	codec := New(store)

	// Then
	assert.IsType(t, new(Codec), codec)
}

func TestGetWhenHit(t *testing.T) {
	// Given
	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store := &mocksStore.StoreInterface{}
	store.On("Get", "my-key").Return(cacheValue, nil)

	codec := New(store)

	// When
	value, err := codec.Get("my-key")

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

func TestGetWhenMiss(t *testing.T) {
	// Given
	expectedErr := errors.New("Unable to find in store")

	store := &mocksStore.StoreInterface{}
	store.On("Get", "my-key").Return(nil, expectedErr)

	codec := New(store)

	// When
	value, err := codec.Get("my-key")

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
	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	options := &store.Options{
		Expiration: 5 * time.Second,
	}

	store := &mocksStore.StoreInterface{}
	store.On("Set", "my-key", cacheValue, options).Return(nil)

	codec := New(store)

	// When
	err := codec.Set("my-key", cacheValue, options)

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
	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	options := &store.Options{
		Expiration: 5 * time.Second,
	}

	expectedErr := errors.New("Unable to set value in store")

	store := &mocksStore.StoreInterface{}
	store.On("Set", "my-key", cacheValue, options).Return(expectedErr)

	codec := New(store)

	// When
	err := codec.Set("my-key", cacheValue, options)

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
	store := &mocksStore.StoreInterface{}
	store.On("Delete", "my-key").Return(nil)

	codec := New(store)

	// When
	err := codec.Delete("my-key")

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
	expectedErr := errors.New("Unable to delete key")

	store := &mocksStore.StoreInterface{}
	store.On("Delete", "my-key").Return(expectedErr)

	codec := New(store)

	// When
	err := codec.Delete("my-key")

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
	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	store := &mocksStore.StoreInterface{}
	store.On("Invalidate", options).Return(nil)

	codec := New(store)

	// When
	err := codec.Invalidate(options)

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
	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	expectedErr := errors.New("Unexpected error when invalidating data")

	store := &mocksStore.StoreInterface{}
	store.On("Invalidate", options).Return(expectedErr)

	codec := New(store)

	// When
	err := codec.Invalidate(options)

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
	store := &mocksStore.StoreInterface{}
	store.On("Clear").Return(nil)

	codec := New(store)

	// When
	err := codec.Clear()

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
	expectedErr := errors.New("Unexpected error when clearing cache")

	store := &mocksStore.StoreInterface{}
	store.On("Clear").Return(expectedErr)

	codec := New(store)

	// When
	err := codec.Clear()

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
	store := &mocksStore.StoreInterface{}

	codec := New(store)

	// When - Then
	assert.Equal(t, store, codec.GetStore())
}

func TestGetStats(t *testing.T) {
	// Given
	store := &mocksStore.StoreInterface{}

	codec := New(store)

	// When - Then
	expectedStats := &Stats{}
	assert.Equal(t, expectedStats, codec.GetStats())
}
