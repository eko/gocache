package nats

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	natsclient "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// integrationBucket creates an isolated bucket on an explicitly supplied test server.
func integrationBucket(t testing.TB) jetstream.KeyValue {
	t.Helper()
	url := os.Getenv("NATS_TEST_URL")
	if url == "" {
		t.Skip("set NATS_TEST_URL to a NATS 2.11+ server with JetStream enabled")
	}
	nc, err := natsclient.Connect(url)
	require.NoError(t, err)
	t.Cleanup(nc.Close)
	js, err := jetstream.New(nc)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	name := fmt.Sprintf("gocache_test_%d", time.Now().UnixNano())
	kv, err := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: name, Storage: jetstream.MemoryStorage, LimitMarkerTTL: 10 * time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		assert.NoError(t, js.DeleteKeyValue(ctx, name))
	})
	return kv
}

func TestIntegrationExpirationAndOverrides(t *testing.T) {
	store := NewNats(integrationBucket(t), lib_store.WithExpiration(time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	written, err := store.SetIfNotExists(ctx, "lock", "owner")
	require.NoError(t, err)
	require.True(t, written)
	written, err = store.SetIfNotExists(ctx, "lock", "other")
	require.NoError(t, err)
	require.False(t, written)
	written, err = store.SetIfNotExists(ctx, "persistent", "value", lib_store.WithExpiration(0))
	require.NoError(t, err)
	require.True(t, written)
	written, err = store.SetIfNotExists(ctx, "longer", "value", lib_store.WithExpiration(30*time.Second))
	require.NoError(t, err)
	require.True(t, written)
	require.Eventually(t, func() bool {
		_, err := store.Get(ctx, "lock")
		return errors.Is(err, lib_store.NotFound{})
	}, 5*time.Second, 25*time.Millisecond)
	for _, key := range []string{"persistent", "longer"} {
		_, err := store.Get(ctx, key)
		require.NoError(t, err)
	}
	written, err = store.SetIfNotExists(ctx, "lock", "new-owner")
	require.NoError(t, err)
	require.True(t, written)
	require.NoError(t, store.Set(ctx, "default-expiration", "value"))
	require.NoError(t, store.Set(ctx, "chained", "value", lib_store.WithExpiration(time.Minute)))
}

func TestIntegrationConcurrentCreateAndClear(t *testing.T) {
	store := NewNats(integrationBucket(t))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var successes atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			written, err := store.SetIfNotExists(ctx, "contended", "value", lib_store.WithExpiration(30*time.Second))
			assert.NoError(t, err)
			if written {
				successes.Add(1)
			}
		}()
	}
	wg.Wait()
	require.Equal(t, int32(1), successes.Load())
	require.NoError(t, store.Set(ctx, "other", "value"))
	require.NoError(t, store.Clear(ctx))
	for _, key := range []string{"contended", "other"} {
		_, err := store.Get(ctx, key)
		assert.ErrorIs(t, err, lib_store.NotFound{})
	}
	require.NoError(t, store.Clear(ctx))
}
