package store

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/coocood/freecache"
)

func BenchmarkFreecacheSet(b *testing.B) {
	ctx := context.Background()

	c := freecache.NewCache(1000)
	freecacheStore := NewFreecache(c, WithExpiration(10*time.Second))

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				key := fmt.Sprintf("test-%d", n)
				value := []byte(fmt.Sprintf("value-%d", n))

				_ = freecacheStore.Set(ctx, key, value)
			}
		})
	}
}

func BenchmarkFreecacheGet(b *testing.B) {
	ctx := context.Background()

	c := freecache.NewCache(1000)
	freecacheStore := NewFreecache(c, WithExpiration(10*time.Second))
	key := "test"
	value := []byte("value")

	err := freecacheStore.Set(ctx, key, value)
	if err != nil {
		b.Error(err)
	}

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				_, _ = freecacheStore.Get(ctx, key)
			}
		})
	}
}
