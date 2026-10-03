package nats

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRead(t *testing.T) {
	for _, withTTL := range []bool{false, true} {
		for _, sourceErr := range []error{nil, jetstream.ErrKeyNotFound, jetstream.ErrKeyDeleted, errors.New("get failed")} {
			ctrl := gomock.NewController(t)
			client := NewMockNatsClientInterface(ctrl)
			entry := NewMockKeyValueEntry(ctrl)
			client.EXPECT().Get(gomock.Any(), "key").Return(entry, sourceErr)
			if sourceErr == nil {
				entry.EXPECT().Value().Return([]byte("value"))
				if withTTL {
					status := NewMockKeyValueStatus(ctrl)
					client.EXPECT().Status(gomock.Any()).Return(status, nil)
					status.EXPECT().Config().Return(jetstream.KeyValueConfig{})
				}
			}
			store := NewNats(client)
			var value any
			var err error
			if withTTL {
				var ttl time.Duration
				value, ttl, err = store.GetWithTTL(context.Background(), "key")
				assert.Zero(t, ttl)
			} else {
				value, err = store.Get(context.Background(), "key")
			}
			assert.ErrorIs(t, err, sourceErr)
			if sourceErr == nil {
				assert.Equal(t, []byte("value"), value)
			} else {
				assert.Nil(t, value)
			}
			if isNotFound(sourceErr) {
				assert.ErrorIs(t, err, lib_store.NotFound{})
			}
		}
	}
}

func TestGetWithTTLCachesConfigurationConcurrently(t *testing.T) {
	for _, ttl := range []time.Duration{0, 30 * time.Second, 5 * time.Minute} {
		ctrl := gomock.NewController(t)
		client := NewMockNatsClientInterface(ctrl)
		status := NewMockKeyValueStatus(ctrl)
		entry := NewMockKeyValueEntry(ctrl)
		client.EXPECT().Status(gomock.Any()).Return(status, nil).Times(1)
		status.EXPECT().Config().Return(jetstream.KeyValueConfig{TTL: ttl}).Times(1)
		client.EXPECT().Get(gomock.Any(), "key").Return(entry, nil).Times(10)
		entry.EXPECT().Value().Return([]byte("value")).Times(10)
		entry.EXPECT().Created().Return(time.Now().Add(-time.Minute)).AnyTimes()
		store := NewNats(client)
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				value, remaining, err := store.GetWithTTL(context.Background(), "key")
				assert.NoError(t, err)
				assert.Equal(t, []byte("value"), value)
				if ttl <= time.Minute {
					assert.Zero(t, remaining)
				} else {
					assert.InDelta(t, 240, remaining.Seconds(), 1)
				}
			}()
		}
		wg.Wait()
	}
}

func TestFailedConfigurationReadCanBeRetried(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := NewMockNatsClientInterface(ctrl)
	status := NewMockKeyValueStatus(ctrl)
	entry := NewMockKeyValueEntry(ctrl)
	wantErr := errors.New("status failed")
	client.EXPECT().Get(gomock.Any(), "key").Return(entry, nil).Times(2)
	gomock.InOrder(
		client.EXPECT().Status(gomock.Any()).Return(nil, wantErr),
		client.EXPECT().Status(gomock.Any()).Return(status, nil),
	)
	status.EXPECT().Config().Return(jetstream.KeyValueConfig{})
	entry.EXPECT().Value().Return([]byte("value"))
	store := NewNats(client)
	value, ttl, err := store.GetWithTTL(context.Background(), "key")
	assert.Nil(t, value)
	assert.Zero(t, ttl)
	assert.ErrorIs(t, err, wantErr)
	_, _, err = store.GetWithTTL(context.Background(), "key")
	require.NoError(t, err)
}

func TestSet(t *testing.T) {
	for _, value := range []any{"value", []byte("value")} {
		client := NewMockNatsClientInterface(gomock.NewController(t))
		client.EXPECT().Put(gomock.Any(), "key", []byte("value")).Return(uint64(1), nil)
		require.NoError(t, NewNats(client).Set(context.Background(), "key", value))
	}
}

func TestSetRejectsTagsBeforeWriting(t *testing.T) {
	for _, option := range []lib_store.Option{
		lib_store.WithTags([]string{"users"}), lib_store.WithTagsTTL(time.Minute),
	} {
		client := NewMockNatsClientInterface(gomock.NewController(t))
		assert.ErrorIs(t, NewNats(client).Set(context.Background(), "key", "value", option), lib_store.ErrNotSupported)
		assert.ErrorIs(t, NewNats(client, option).Set(context.Background(), "key", "value"), lib_store.ErrNotSupported)
	}
}

func TestSetIgnoresExpiration(t *testing.T) {
	for _, option := range []lib_store.Option{
		lib_store.WithExpiration(time.Minute), lib_store.WithExpiration(-time.Second),
	} {
		client := NewMockNatsClientInterface(gomock.NewController(t))
		client.EXPECT().Put(gomock.Any(), "key", []byte("value")).Return(uint64(1), nil).Times(2)
		assert.NoError(t, NewNats(client).Set(context.Background(), "key", "value", option))
		assert.NoError(t, NewNats(client, option).Set(context.Background(), "key", "value"))
	}
}

func TestSetOverridesDefaults(t *testing.T) {
	client := NewMockNatsClientInterface(gomock.NewController(t))
	client.EXPECT().Put(gomock.Any(), "key", []byte("value")).Return(uint64(1), nil)
	store := NewNats(client, lib_store.WithExpiration(time.Minute))
	require.NoError(t, store.Set(context.Background(), "key", "value", lib_store.WithExpiration(0)))
	assert.Equal(t, time.Minute, store.options.Expiration)
}

func TestWritesRejectUnsupportedValues(t *testing.T) {
	store := NewNats(NewMockNatsClientInterface(gomock.NewController(t)))
	assert.Error(t, store.Set(context.Background(), "key", 42))
	written, err := store.SetIfNotExists(context.Background(), "key", 42)
	assert.False(t, written)
	assert.Error(t, err)
}

func TestWriteErrors(t *testing.T) {
	for _, wantErr := range []error{nil, jetstream.ErrKeyExists, errors.New("write failed")} {
		client := NewMockNatsClientInterface(gomock.NewController(t))
		client.EXPECT().Create(gomock.Any(), "key", []byte("value")).Return(uint64(1), wantErr)
		written, err := NewNats(client).SetIfNotExists(context.Background(), "key", "value")
		assert.Equal(t, wantErr == nil, written)
		if errors.Is(wantErr, jetstream.ErrKeyExists) {
			assert.NoError(t, err)
		} else {
			assert.ErrorIs(t, err, wantErr)
		}
		client.EXPECT().Put(gomock.Any(), "key", []byte("value")).Return(uint64(1), wantErr)
		assert.ErrorIs(t, NewNats(client).Set(context.Background(), "key", "value"), wantErr)
	}
}

func TestSetIfNotExistsRejectsUnsupportedOptions(t *testing.T) {
	for _, option := range []lib_store.Option{lib_store.WithTags([]string{"tag"}), lib_store.WithTagsTTL(time.Minute), lib_store.WithExpiration(-time.Second)} {
		store := NewNats(NewMockNatsClientInterface(gomock.NewController(t)), option)
		written, err := store.SetIfNotExists(context.Background(), "key", "value")
		assert.False(t, written)
		assert.Error(t, err)
	}
}

func TestSetIfNotExistsRequiresLimitMarkerTTL(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := NewMockNatsClientInterface(ctrl)
	status := NewMockKeyValueStatus(ctrl)
	client.EXPECT().Status(gomock.Any()).Return(status, nil)
	status.EXPECT().Config().Return(jetstream.KeyValueConfig{})
	written, err := NewNats(client).SetIfNotExists(context.Background(), "key", "value", lib_store.WithExpiration(time.Minute))
	assert.False(t, written)
	assert.ErrorIs(t, err, lib_store.ErrNotSupported)
}

func TestDelete(t *testing.T) {
	for _, wantErr := range []error{nil, jetstream.ErrKeyNotFound, jetstream.ErrKeyDeleted, errors.New("delete failed")} {
		client := NewMockNatsClientInterface(gomock.NewController(t))
		client.EXPECT().Delete(gomock.Any(), "key").Return(wantErr)
		err := NewNats(client).Delete(context.Background(), "key")
		if isNotFound(wantErr) {
			assert.NoError(t, err)
		} else {
			assert.ErrorIs(t, err, wantErr)
		}
	}
}

func TestInvalidateIsNoOp(t *testing.T) {
	store := NewNats(NewMockNatsClientInterface(gomock.NewController(t)))
	assert.NoError(t, store.Invalidate(context.Background()))
	assert.NoError(t, store.Invalidate(context.Background(), lib_store.WithInvalidateTags([]string{"users"})))
	assert.Equal(t, NatsType, store.GetType())
}

func TestClearStopsListerOnSuccessAndFailure(t *testing.T) {
	for _, wantErr := range []error{nil, errors.New("delete failed")} {
		ctrl := gomock.NewController(t)
		client := NewMockNatsClientInterface(ctrl)
		lister := NewMockKeyLister(ctrl)
		keys := make(chan string, 2)
		keys <- "one"
		keys <- "two"
		close(keys)
		client.EXPECT().ListKeys(gomock.Any()).Return(lister, nil)
		lister.EXPECT().Keys().Return((<-chan string)(keys))
		lister.EXPECT().Stop().Return(nil)
		client.EXPECT().Delete(gomock.Any(), "one").Return(wantErr)
		if wantErr == nil {
			client.EXPECT().Delete(gomock.Any(), "two").Return(nil)
		}
		assert.ErrorIs(t, NewNats(client).Clear(context.Background()), wantErr)
	}
}

// blockingLister mirrors the nats.go keyLister producer: a goroutine doing a
// blocking send that only ends once the channel has been fully consumed.
type blockingLister struct {
	keys chan string
	done chan struct{}
}

func newBlockingLister(count int) *blockingLister {
	lister := &blockingLister{keys: make(chan string), done: make(chan struct{})}
	go func() {
		defer close(lister.done)
		defer close(lister.keys)
		for i := 0; i < count; i++ {
			lister.keys <- "key"
		}
	}()
	return lister
}

func (l *blockingLister) Keys() <-chan string { return l.keys }

func (l *blockingLister) Stop() error { return nil }

func TestClearDrainsListerOnFailure(t *testing.T) {
	client := NewMockNatsClientInterface(gomock.NewController(t))
	lister := newBlockingLister(10)
	wantErr := errors.New("delete failed")
	client.EXPECT().ListKeys(gomock.Any()).Return(lister, nil)
	client.EXPECT().Delete(gomock.Any(), "key").Return(wantErr)

	assert.ErrorIs(t, NewNats(client).Clear(context.Background()), wantErr)

	select {
	case <-lister.done:
	case <-time.After(time.Second):
		t.Fatal("lister goroutine is still blocked after Clear returned")
	}
}

func TestClearListError(t *testing.T) {
	wantErr := errors.New("list failed")
	client := NewMockNatsClientInterface(gomock.NewController(t))
	client.EXPECT().ListKeys(gomock.Any()).Return(nil, wantErr)
	assert.ErrorIs(t, NewNats(client).Clear(context.Background()), wantErr)
}
