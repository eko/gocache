package store

import (
	"errors"
	"fmt"
	"testing"
	"time"

	mocksStore "github.com/eko/gocache/test/mocks/store/clients"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewFreecache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocksStore.NewMockFreecacheClientInterface(ctrl)
	options := &Options{
		Expiration: 6 * time.Second,
	}

	// When
	store := NewFreecache(client, options)

	// Then
	assert.IsType(t, new(FreecacheStore), store)
	assert.Equal(t, client, store.client)
	assert.Equal(t, options, store.options)
}

func TestFreecacheGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocksStore.NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Get([]byte("key1")).Return([]byte("val1"), nil)
	client.EXPECT().Get([]byte("key2")).Return([]byte("val2"), nil)

	s := NewFreecache(client, nil)

	value, err := s.Get("key1")
	assert.Nil(t, err)
	assert.Equal(t, []byte("val1"), value)

	value, err = s.Get("key2")
	assert.Nil(t, err)
	assert.Equal(t, []byte("val2"), value)
}

func TestFreecacheGetNotFound(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocksStore.NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Get([]byte("key1")).Return(nil, errors.New("value not found in Freecache store"))

	s := NewFreecache(client, nil)

	value, err := s.Get("key1")
	assert.EqualError(t, err, "value not found in Freecache store")
	assert.Nil(t, value)
}

func TestFreecacheSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")
	options := &Options{
		Expiration: 6 * time.Second,
	}

	client := mocksStore.NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Set([]byte(cacheKey), cacheValue, 6).Return(nil)

	s := NewFreecache(client, options)
	err := s.Set(cacheKey, cacheValue, options)
	assert.Nil(t, err)
}

func TestFreecacheSetInvalidValue(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"
	options := &Options{
		Expiration: 6 * time.Second,
	}
	expectedErr := errors.New("value type not supported by Freecache store")

	client := mocksStore.NewMockFreecacheClientInterface(ctrl)

	s := NewFreecache(client, options)
	err := s.Set(cacheKey, cacheValue, options)
	assert.Equal(t, expectedErr, err)
}

func TestFreecacheSetInvalidSize(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")
	options := &Options{
		Expiration: 6 * time.Second,
	}
	expectedErr := fmt.Errorf("size of key: %v, value: %v, err: %v", cacheKey, cacheValue, errors.New(""))
	client := mocksStore.NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Set([]byte(cacheKey), cacheValue, 6).Return(expectedErr)

	s := NewFreecache(client, options)
	err := s.Set(cacheKey, cacheValue, options)
	assert.NotNil(t, err)

}

func TestFreecacheSetInvalidKey(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := 1
	cacheValue := []byte("my-cache-value")
	options := &Options{
		Expiration: 6 * time.Second,
	}

	expectedErr := errors.New("key type not supported by Freecache store")

	client := mocksStore.NewMockFreecacheClientInterface(ctrl)

	s := NewFreecache(client, options)
	err := s.Set(cacheKey, cacheValue, options)
	assert.Equal(t, expectedErr, err)
}

func TestFreecacheDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "key"

	client := mocksStore.NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Del(gomock.Any()).Return(true)

	s := NewFreecache(client, nil)
	err := s.Delete(cacheKey)
	assert.Nil(t, err)
}

func TestFreecacheDeleteFailed(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := "key"
	expectedErr := fmt.Errorf("failed to delete key %v", cacheKey)
	client := mocksStore.NewMockFreecacheClientInterface(ctrl)
	client.EXPECT().Del(gomock.Any()).Return(false)

	s := NewFreecache(client, nil)
	err := s.Delete(cacheKey)
	assert.Equal(t, expectedErr, err)
}

func TestFreecacheDeleteInvalidKey(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cacheKey := 1
	expectedErr := errors.New("key type not supported by Freecache store")
	client := mocksStore.NewMockFreecacheClientInterface(ctrl)

	s := NewFreecache(client, nil)
	err := s.Delete(cacheKey)
	assert.Equal(t, expectedErr, err)
}
