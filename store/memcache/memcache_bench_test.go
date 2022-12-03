package memcache

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	lib_store "github.com/eko/gocache/v4/lib/store"
)

func BenchmarkMemcacheSet(b *testing.B) {
	ctx := context.Background()

	store := NewMemcache(
		memcache.New("127.0.0.1:11211"),
		lib_store.WithExpiration(100*time.Second),
	)

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

func BenchmarkMemcacheGet(b *testing.B) {
	ctx := context.Background()

	store := NewMemcache(
		memcache.New("127.0.0.1:11211"),
		lib_store.WithExpiration(100*time.Second),
	)

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
