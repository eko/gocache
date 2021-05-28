package store

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/go-redis/redis/v8"
)

func BenchmarkRedisSet(b *testing.B) {
	ctx := context.Background()

	store := NewRedis(redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	}), nil)

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

func BenchmarkRedisGet(b *testing.B) {
	ctx := context.Background()

	store := NewRedis(redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	}), nil)

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
