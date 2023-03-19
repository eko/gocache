package rediscluster

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"

	lib_store "github.com/eko/gocache/lib/v4/store"
	redis "github.com/redis/go-redis/v9"
)

// should be configured to connect to real Redis Cluster
func BenchmarkRedisClusterSet(b *testing.B) {
	ctx := context.Background()

	addr := strings.Split("redis:6379", ",")
	store := NewRedisCluster(redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: addr,
	}), nil)

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

func BenchmarkRedisClusterGet(b *testing.B) {
	ctx := context.Background()

	addr := strings.Split("redis:6379", ",")
	store := NewRedisCluster(redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: addr,
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
