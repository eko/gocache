package store

import (
	"context"
	"fmt"
	"github.com/rueian/rueidis"
	"math"
	"testing"
	"time"
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

	store := NewRueidis(ruedisClient, nil, WithExpiration(time.Hour*4))

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				key := fmt.Sprintf("test-%d", n)
				value := []byte(fmt.Sprintf("value-%d", n))

				store.Set(ctx, key, value, WithTags([]string{fmt.Sprintf("tag-%d", n)}))
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

	store := NewRueidis(ruedisClient, nil, WithExpiration(time.Hour*4))

	key := "test"
	value := []byte("value")

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
