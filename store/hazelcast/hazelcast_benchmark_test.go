package hazelcast

import (
	"context"
	"fmt"
	"math"
	"testing"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/hazelcast/hazelcast-go-client"
)

func BenchmarkHazelcastSet(b *testing.B) {
	ctx := context.Background()

	hzClient, err := hazelcast.StartNewClient(ctx)
	if err != nil {
		b.Fatalf("Failed to start client: %v", err)
	}

	hzMap, err := hzClient.GetMap(ctx, "gocache")
	if err != nil {
		b.Fatalf("Failed to get map: %v", err)
	}

	store := NewHazelcast(hzMap)

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

func BenchmarkHazelcastGet(b *testing.B) {
	ctx := context.Background()

	hzClient, err := hazelcast.StartNewClient(ctx)
	if err != nil {
		b.Fatalf("Failed to start client: %v", err)
	}

	hzMap, err := hzClient.GetMap(ctx, "gocache")
	if err != nil {
		b.Fatalf("Failed to get map: %v", err)
	}

	store := NewHazelcast(hzMap)

	key := "test"
	value := []byte("value")

	store.Set(ctx, key, value)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				_, _ = store.Get(ctx, key)
			}
		})
	}
}
