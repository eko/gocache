package store

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"

	redis "github.com/go-redis/redis/v8"
)

// should be configured to connect to real Redis Cluster
func BenchmarkRedisClusterSet(b *testing.B) {
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

				store.Set(context.TODO(), key, value, &Options{
					Tags: []string{fmt.Sprintf("tag-%d", n)},
				})
			}
		})
	}
}

func BenchmarkRedisClusterGet(b *testing.B) {
	addr := strings.Split("redis:6379", ",")
	store := NewRedisCluster(redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: addr,
	}), nil)

	key := "test"
	value := []byte("value")

	store.Set(context.TODO(), key, value, nil)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				_, _ = store.Get(context.TODO(), key)
			}
		})
	}
}
