package go_cache

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	lib_store "github.com/eko/gocache/v4/lib/store"
	cache "github.com/patrickmn/go-cache"
)

func BenchmarkGoCacheSet(b *testing.B) {
	ctx := context.Background()

	client := cache.New(10*time.Second, 30*time.Second)

	store := NewGoCache(client, nil)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				key := fmt.Sprintf("test-%d", n)
				value := []byte(fmt.Sprintf("value-%d", n))

				store.Set(ctx, key, value, lib_store.WithTags([]string{fmt.Sprintf("tag-%d", n)}))
			}
		})
	}
}

func BenchmarkGoCacheGet(b *testing.B) {
	ctx := context.Background()

	client := cache.New(10*time.Second, 30*time.Second)

	store := NewGoCache(client, nil)

	key := "test"
	value := []byte("value")

	store.Set(ctx, key, value, nil)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				_, _ = store.Get(ctx, key)
			}
		})
	}
}
