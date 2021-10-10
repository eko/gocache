package marshaler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eko/gocache/v2/store"
	mocksCache "github.com/eko/gocache/v2/test/mocks/cache"
	"github.com/eko/gocache/v2/test/proto/proto3_proto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
)

type testProtoCacheValue struct {
	Hello string
}

func TestNewProtoMarshaler(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache := mocksCache.NewMockCacheInterface(ctrl)

	// When
	marshaler := NewProtoMarshaler(cache)

	// Then
	assert.IsType(t, new(ProtoMarshaler), marshaler)
	assert.Equal(t, cache, marshaler.cache)
}

func TestNewProtoMarshalerWithMarshalOptions(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache := mocksCache.NewMockCacheInterface(ctrl)

	// When
	marshaler := NewProtoMarshaler(cache, WithMarshalerOption(protojson.MarshalOptions{
		AllowPartial: true,
	}))

	// Then
	assert.IsType(t, new(ProtoMarshaler), marshaler)
	assert.Equal(t, protojson.MarshalOptions{
		AllowPartial: true,
	}, marshaler.marshalOpts)
	assert.Equal(t, cache, marshaler.cache)
}

func TestNewProtoMarshalerWithUnmarshalOptions(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache := mocksCache.NewMockCacheInterface(ctrl)

	// When
	marshaler := NewProtoMarshaler(cache, WithUnmarshalerOption(protojson.UnmarshalOptions{
		AllowPartial: true,
	}))

	// Then
	assert.IsType(t, new(ProtoMarshaler), marshaler)
	assert.Equal(t, protojson.UnmarshalOptions{
		AllowPartial: true,
	}, marshaler.unmarshalOpts)
	assert.Equal(t, cache, marshaler.cache)
}

func TestProtoGetWhenStoreReturnsSliceOfBytes(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &proto3_proto.Message{
		Submessage: &proto3_proto.Message{
			StringMap: map[string]string{
				"hello": "world",
			},
		},
		StringMap: map[string]string{
			"hello": "world",
		},
	}

	cacheValueBytes, err := protojson.MarshalOptions{}.Marshal(cacheValue)
	if err != nil {
		assert.Error(t, err)
	}

	cache := mocksCache.NewMockCacheInterface(ctrl)
	cache.EXPECT().Get(ctx, "my-key").Return(cacheValueBytes, nil)

	marshaler := NewProtoMarshaler(cache)

	retValue := &proto3_proto.Message{}
	// When
	value, err := marshaler.Get(ctx, "my-key", retValue)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestProtoGetWhenStoreReturnsString(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &proto3_proto.Message{
		Submessage: &proto3_proto.Message{
			StringMap: map[string]string{
				"hello": "world",
			},
		},
		StringMap: map[string]string{
			"hello": "world",
		},
	}

	cacheValueBytes, err := protojson.MarshalOptions{}.Marshal(cacheValue)
	if err != nil {
		assert.Error(t, err)
	}

	cache := mocksCache.NewMockCacheInterface(ctrl)
	cache.EXPECT().Get(ctx, "my-key").Return(string(cacheValueBytes), nil)

	marshaler := NewProtoMarshaler(cache)

	retValue := &proto3_proto.Message{}

	// When
	value, err := marshaler.Get(ctx, "my-key", retValue)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestProtoGetWhenUnmarshalingError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache := mocksCache.NewMockCacheInterface(ctrl)
	cache.EXPECT().Get(ctx, "my-key").Return("unknown-string", nil)

	marshaler := NewProtoMarshaler(cache)

	retValue := &proto3_proto.Message{}

	// When
	value, err := marshaler.Get(ctx, "my-key", retValue)

	// Then
	assert.NotNil(t, err)
	assert.Nil(t, value)
}

func TestProtoGetWhenNotFoundInStore(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("Unable to find item in store")

	cache := mocksCache.NewMockCacheInterface(ctrl)
	cache.EXPECT().Get(ctx, "my-key").Return(nil, expectedErr)

	marshaler := NewProtoMarshaler(cache)

	retValue := &proto3_proto.Message{}

	// When
	value, err := marshaler.Get(ctx, "my-key", retValue)

	// Then
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, value)
}

func TestProtoSetWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &proto3_proto.MessageWithMap{
		ByteMapping: map[bool][]byte{
			true: []byte("hello world"),
		},
	}

	cacheValueBytes, err := protojson.MarshalOptions{}.Marshal(cacheValue)
	if err != nil {
		assert.Error(t, err)
	}

	options := &store.Options{
		Expiration: 5 * time.Second,
	}

	expectedErr := errors.New("An unexpected error occurred")

	cache := mocksCache.NewMockCacheInterface(ctrl)
	cache.EXPECT().Set(ctx, "my-key", cacheValueBytes, options).Return(expectedErr)

	marshaler := NewProtoMarshaler(cache)

	// When
	err = marshaler.Set(ctx, "my-key", cacheValue, options)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestProtoDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache := mocksCache.NewMockCacheInterface(ctrl)
	cache.EXPECT().Delete(ctx, "my-key").Return(nil)

	marshaler := NewProtoMarshaler(cache)

	// When
	err := marshaler.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)
}

func TestProtoDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("Unable to delete key")

	cache := mocksCache.NewMockCacheInterface(ctrl)
	cache.EXPECT().Delete(ctx, "my-key").Return(expectedErr)

	marshaler := NewProtoMarshaler(cache)

	// When
	err := marshaler.Delete(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestProtoInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	cache := mocksCache.NewMockCacheInterface(ctrl)
	cache.EXPECT().Invalidate(ctx, options).Return(nil)

	marshaler := NewProtoMarshaler(cache)

	// When
	err := marshaler.Invalidate(ctx, options)

	// Then
	assert.Nil(t, err)
}

func TestProtoInvalidatingWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	options := store.InvalidateOptions{
		Tags: []string{"tag1"},
	}

	expectedErr := errors.New("Unexpected error when invalidating data")

	cache := mocksCache.NewMockCacheInterface(ctrl)
	cache.EXPECT().Invalidate(ctx, options).Return(expectedErr)

	marshaler := NewProtoMarshaler(cache)

	// When
	err := marshaler.Invalidate(ctx, options)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestProtoClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache := mocksCache.NewMockCacheInterface(ctrl)
	cache.EXPECT().Clear(ctx).Return(nil)

	marshaler := NewProtoMarshaler(cache)

	// When
	err := marshaler.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestProtoClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("An unexpected error occurred")

	cache := mocksCache.NewMockCacheInterface(ctrl)
	cache.EXPECT().Clear(ctx).Return(expectedErr)

	marshaler := NewProtoMarshaler(cache)

	// When
	err := marshaler.Clear(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}
