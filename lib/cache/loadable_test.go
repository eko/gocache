package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	mockcache "github.com/eko/gocache/lib/v4/internal/mocks/cache"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewLoadable(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := mockcache.NewMockSetterCacheInterface[any](ctrl)

	loadFunc := func(_ context.Context, key any) (any, []store.Option, error) {
		return "test data loaded", []store.Option{}, nil
	}

	// When
	cache := NewLoadable[any](loadFunc, cache1)
	defer cache.Close()

	// Then
	assert.IsType(t, new(LoadableCache[any]), cache)

	assert.IsType(t, new(LoadFunction[any]), &cache.loadFunc)
	assert.Equal(t, cache1, cache.cache)
}

func TestLoadableGetWhenAlreadyInCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	cache1 := mockcache.NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return(cacheValue, nil)

	loadFunc := func(_ context.Context, key any) (any, []store.Option, error) {
		return nil, []store.Option{}, errors.New("should not be called")
	}

	cache := NewLoadable[any](loadFunc, cache1)
	defer cache.Close()

	// When
	value, err := cache.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestLoadableGetWhenNotAvailableInLoadFunc(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// Cache
	cache1 := mockcache.NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return(nil, errors.New("unable to find in cache 1"))

	loadFunc := func(_ context.Context, key any) (any, []store.Option, error) {
		return nil, []store.Option{}, errors.New("an error has occurred while loading data from custom source")
	}

	cache := NewLoadable[any](loadFunc, cache1)
	defer cache.Close()

	// When
	value, err := cache.Get(ctx, "my-key")

	// Then
	assert.Nil(t, value)
	assert.Equal(t, errors.New("an error has occurred while loading data from custom source"), err)
}

func TestLoadableGetWhenAvailableInLoadFunc(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	// Cache 1
	cache1 := mockcache.NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return(nil, errors.New("unable to find in cache 1"))
	cache1.EXPECT().Set(ctx, "my-key", cacheValue).AnyTimes().Return(nil)

	var loadCallCount int32
	pauseLoadFn := make(chan struct{})

	loadFunc := func(_ context.Context, key any) (any, []store.Option, error) {
		atomic.AddInt32(&loadCallCount, 1)
		<-pauseLoadFn
		time.Sleep(time.Millisecond * 10)
		return cacheValue, []store.Option{}, nil
	}

	cache := NewLoadable[any](loadFunc, cache1)
	defer cache.Close()

	const numRequests = 3
	var started sync.WaitGroup
	started.Add(numRequests)
	var finished sync.WaitGroup
	finished.Add(numRequests)
	for i := 0; i < numRequests; i++ {
		go func() {
			defer finished.Done()
			started.Done()
			// When
			value, err := cache.Get(ctx, "my-key")

			// Wait for data to be processed
			for len(cache.setChannel) > 0 {
				time.Sleep(1 * time.Millisecond)
			}

			// Then
			assert.Nil(t, err)
			assert.Equal(t, cacheValue, value)
		}()
	}

	started.Wait()
	close(pauseLoadFn)
	finished.Wait()

	assert.Equal(t, int32(1), loadCallCount)
}

func TestLoadableDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := mockcache.NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(nil)

	loadFunc := func(_ context.Context, key any) (any, []store.Option, error) {
		return "a value", []store.Option{}, nil
	}

	cache := NewLoadable[any](loadFunc, cache1)
	defer cache.Close()

	// When
	err := cache.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)
}

func TestLoadableDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unable to delete key")

	cache1 := mockcache.NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(expectedErr)

	loadFunc := func(_ context.Context, key any) (any, []store.Option, error) {
		return "a value", []store.Option{}, nil
	}

	cache := NewLoadable[any](loadFunc, cache1)
	defer cache.Close()

	// When
	err := cache.Delete(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestLoadableInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := mockcache.NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Invalidate(ctx).Return(nil)

	loadFunc := func(_ context.Context, key any) (any, []store.Option, error) {
		return "a value", []store.Option{}, nil
	}

	cache := NewLoadable[any](loadFunc, cache1)
	defer cache.Close()

	// When
	err := cache.Invalidate(ctx)

	// Then
	assert.Nil(t, err)
}

func TestLoadableInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unexpected error when invalidating data")

	cache1 := mockcache.NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Invalidate(ctx).Return(expectedErr)

	loadFunc := func(_ context.Context, key any) (any, []store.Option, error) {
		return "a value", []store.Option{}, nil
	}

	cache := NewLoadable[any](loadFunc, cache1)
	defer cache.Close()

	// When
	err := cache.Invalidate(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestLoadableClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := mockcache.NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Clear(ctx).Return(nil)

	loadFunc := func(_ context.Context, key any) (any, []store.Option, error) {
		return "a value", []store.Option{}, nil
	}

	cache := NewLoadable[any](loadFunc, cache1)
	defer cache.Close()

	// When
	err := cache.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestLoadableClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unexpected error when invalidating data")

	cache1 := mockcache.NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Clear(ctx).Return(expectedErr)

	loadFunc := func(_ context.Context, key any) (any, []store.Option, error) {
		return "a value", []store.Option{}, nil
	}

	cache := NewLoadable[any](loadFunc, cache1)
	defer cache.Close()

	// When
	err := cache.Clear(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestLoadableGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := mockcache.NewMockSetterCacheInterface[any](ctrl)

	loadFunc := func(_ context.Context, key any) (any, []store.Option, error) {
		return "test data loaded", []store.Option{}, nil
	}

	cache := NewLoadable[any](loadFunc, cache1)
	defer cache.Close()

	// When - Then
	assert.Equal(t, LoadableType, cache.GetType())
}

func TestLoadableGetTwice(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := mockcache.NewMockSetterCacheInterface[any](ctrl)

	var counter atomic.Uint64
	loadFunc := func(_ context.Context, key any) (any, []store.Option, error) {
		return counter.Add(1), []store.Option{}, nil
	}

	cache := NewLoadable[any](loadFunc, cache1)
	defer cache.Close()

	key := 1
	cache1.EXPECT().Get(context.Background(), key).Return(nil, store.NotFound{}).AnyTimes()
	cache1.EXPECT().Set(context.Background(), key, uint64(1)).Times(1)
	v1, err1 := cache.Get(context.Background(), key)
	v2, err2 := cache.Get(context.Background(), key) // setter may not be called now because it's done by another goroutine
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, uint64(1), v1)
	assert.Equal(t, uint64(1), v2)
	assert.Equal(t, uint64(1), counter.Load())
	_ = cache.Close() // wait for setter
}

func TestLoadableClose(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := mockcache.NewMockSetterCacheInterface[any](ctrl)

	loadFunc := func(_ context.Context, key any) (any, []store.Option, error) {
		return "test data loaded", []store.Option{}, nil
	}

	cache := NewLoadable[any](loadFunc, cache1)

	// When - Close joins the setter goroutine, so it only returns once released
	closed := make(chan error, 1)
	go func() {
		closed <- cache.Close()
	}()

	// Then
	select {
	case err := <-closed:
		assert.Nil(t, err)
	case <-time.After(time.Second):
		t.Fatal("Close did not release the setter goroutine")
	}
}

func TestLoadableCloseWhenCalledMultipleTimes(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := mockcache.NewMockSetterCacheInterface[any](ctrl)

	loadFunc := func(_ context.Context, key any) (any, []store.Option, error) {
		return "test data loaded", []store.Option{}, nil
	}

	cache := NewLoadable[any](loadFunc, cache1)

	// When - Then
	assert.Nil(t, cache.Close())
	assert.Nil(t, cache.Close())
}

func TestLoadableCloseWhileGetting(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := mockcache.NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Get(ctx, gomock.Any()).AnyTimes().Return(nil, store.NotFound{})
	cache1.EXPECT().Set(ctx, gomock.Any(), "test data loaded").AnyTimes().Return(nil)

	loadFunc := func(_ context.Context, key any) (any, []store.Option, error) {
		return "test data loaded", []store.Option{}, nil
	}

	cache := NewLoadable[any](loadFunc, cache1)

	const numRequests = 100

	var finished sync.WaitGroup
	finished.Add(numRequests)

	// When - loading concurrently with Close must not send on a closed channel
	for i := 0; i < numRequests; i++ {
		go func(i int) {
			defer finished.Done()

			value, err := cache.Get(ctx, fmt.Sprintf("my-key-%d", i))

			// Then
			assert.Nil(t, err)
			assert.Equal(t, "test data loaded", value)
		}(i)
	}

	assert.Nil(t, cache.Close())

	finished.Wait()
}
