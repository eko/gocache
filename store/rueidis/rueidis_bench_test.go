package rueidis

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/redis/rueidis"
)

func BenchmarkRueidisSet(b *testing.B) {
	ctx := context.Background()

	ruedisClient, _ := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{"localhost:26379"},
		Sentinel: rueidis.SentinelOption{
			MasterSet: "mymaster",
		},
		SelectDB: 0,
	})

	store := NewRueidis(ruedisClient, lib_store.WithExpiration(time.Hour*4))

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				key := fmt.Sprintf("test-%d", n)
				value := fmt.Sprintf("value-%d", n)

				store.Set(ctx, key, value, lib_store.WithTags([]string{fmt.Sprintf("tag-%d", n)}))
			}
		})
	}
}

func BenchmarkRueidisGet(b *testing.B) {
	ctx := context.Background()

	ruedisClient, _ := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{"localhost:26379"},
		Sentinel: rueidis.SentinelOption{
			MasterSet: "mymaster",
		},
		SelectDB: 0,
	})

	store := NewRueidis(ruedisClient, lib_store.WithExpiration(time.Hour*4))

	key := "test"
	value := "value"

	_ = store.Set(ctx, key, value)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				_, _ = store.Get(ctx, key)
			}
		})
	}
}
