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
