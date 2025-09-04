package main

import (
	"context"
	"fmt"
	"time"

	ristretto "github.com/dgraph-io/ristretto/v2"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/metrics"
	"github.com/eko/gocache/lib/v4/store"
	store_ristretto "github.com/eko/gocache/store/ristretto/v4"
)

func main() {
	ctx := context.Background()

	ristrettoClient1, err := ristretto.NewCache(
		&ristretto.Config[string, []byte]{
			NumCounters: 1_000,
			MaxCost:     100_000_000,
			BufferItems: 64,
		},
	)
	if err != nil {
		panic(err)
	}

	ristrettoClient2, err := ristretto.NewCache(
		&ristretto.Config[string, []byte]{
			NumCounters: 1_000,
			MaxCost:     100_000_000,
			BufferItems: 64,
		},
	)
	if err != nil {
		panic(err)
	}

	// First: memory store
	memoryStore1 := store_ristretto.NewRistretto(
		ristrettoClient1,
		store.WithExpiration(1*time.Minute),
	)
	memoryCache1 := cache.New[[]byte](memoryStore1)

	// Second: memory store
	memoryStore2 := store_ristretto.NewRistretto(
		ristrettoClient2,
		store.WithExpiration(1*time.Minute),
	)
	memoryCache2 := cache.New[[]byte](memoryStore2)

	// Chain cache
	chainCache := cache.NewChain[[]byte](
		memoryCache1,
		memoryCache2,
	)

	// Wrap chain cache with metrics cache
	promMetrics := metrics.NewPrometheus(
		"manifest-api",
	)

	metricCache := cache.NewMetric[[]byte](
		promMetrics,
		chainCache,
	)

	// Loadable
	loadableCache := cache.NewLoadable[[]byte](
		func(ctx context.Context, key any) ([]byte, []store.Option, error) {
			return []byte(`ok-1`), nil, nil
		},
		metricCache,
	)

	loadableCache.Set(ctx, "my-key-1", []byte(`value-1`))
	time.Sleep(100 * time.Millisecond)

	// Remove from first cache, will be fetch from second cache and
	// set back to first cache.
	memoryCache1.Delete(ctx, "my-key-1")
	time.Sleep(100 * time.Millisecond)

	// Ensure value has been cached in both.
	value1, _ := memoryCache1.Get(ctx, "my-key-1")
	time.Sleep(100 * time.Millisecond)
	fmt.Println("value1:", string(value1))

	value2, _ := memoryCache2.Get(ctx, "my-key-1")
	time.Sleep(100 * time.Millisecond)
	fmt.Println("value2:", string(value2))

	// Retrieve final value.
	value, err := loadableCache.Get(ctx, "my-key-1")
	if err != nil {
		panic(err)
	}
	time.Sleep(100 * time.Millisecond)
	fmt.Println("final:", string(value))

	// Retrieve from every cache independently.
	value12, _ := memoryCache1.Get(ctx, "my-key-1")
	time.Sleep(100 * time.Millisecond)
	fmt.Println("value1:", string(value12))

	value22, _ := memoryCache2.Get(ctx, "my-key-1")
	time.Sleep(100 * time.Millisecond)
	fmt.Println("value2:", string(value22))
}
