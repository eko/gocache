package store

import (
	"fmt"
	"math"
	"testing"

	"github.com/yeqown/gocache/types"

	redis "github.com/go-redis/redis/v7"
)

func BenchmarkRedisSet(b *testing.B) {
	store := NewRedis(redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	}), nil)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				key := fmt.Sprintf("test-%d", n)
				value := []byte(fmt.Sprintf("value-%d", n))

				_ = store.Set(key, value, &types.StoreOptions{
					Tags: []string{fmt.Sprintf("tag-%d", n)},
				})
			}
		})
	}
}

func BenchmarkRedisGet(b *testing.B) {
	store := NewRedis(redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	}), nil)

	key := "test"
	value := []byte("value")

	_ = store.Set(key, value, nil)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				_, _ = store.Get(key)
			}
		})
	}
}
