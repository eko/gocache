package valkey

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/valkey-io/valkey-go"
)

func BenchmarkValkeySet(b *testing.B) {
	ctx := context.Background()

	valkeyClient, _ := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{"localhost:26379"},
		Sentinel: valkey.SentinelOption{
			MasterSet: "mymaster",
		},
		SelectDB: 0,
	})

	store := NewValkey(valkeyClient, lib_store.WithExpiration(time.Hour*4))

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for range b.N * n {
				key := fmt.Sprintf("test-%d", n)
				value := fmt.Sprintf("value-%d", n)

				store.Set(ctx, key, value, lib_store.WithTags([]string{fmt.Sprintf("tag-%d", n)}))
			}
		})
	}
}

func BenchmarkValkeyGet(b *testing.B) {
	ctx := context.Background()

	valkeyClient, _ := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{"localhost:26379"},
		Sentinel: valkey.SentinelOption{
			MasterSet: "mymaster",
		},
		SelectDB: 0,
	})

	store := NewValkey(valkeyClient, lib_store.WithExpiration(time.Hour*4))

	key := "test"
	value := "value"

	_ = store.Set(ctx, key, value)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for range b.N * n {
				_, _ = store.Get(ctx, key)
			}
		})
	}
}
