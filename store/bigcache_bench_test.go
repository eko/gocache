package store

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
)

func BenchmarkBigcacheSet(b *testing.B) {
	ctx := context.Background()

	client, _ := bigcache.NewBigCache(bigcache.DefaultConfig(5 * time.Minute))
	store := NewBigcache(client, nil)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				key := fmt.Sprintf("test-%d", n)
				value := []byte(fmt.Sprintf("value-%d", n))

				store.Set(ctx, key, value, &Options{
					Tags: []string{fmt.Sprintf("tag-%d", n)},
				})
			}
		})
	}
}

func BenchmarkBigcacheGet(b *testing.B) {
	ctx := context.Background()

	client, _ := bigcache.NewBigCache(bigcache.DefaultConfig(5 * time.Minute))
	store := NewBigcache(client, nil)

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
