package ristretto

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/dgraph-io/ristretto/v2"
	lib_store "github.com/eko/gocache/lib/v4/store"
)

func BenchmarkRistrettoSet(b *testing.B) {
	ctx := context.Background()

	client, err := ristretto.NewCache(&ristretto.Config[string, []byte]{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	store := NewRistretto(client)

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

func BenchmarkRistrettoSetWithSynchronousSet(b *testing.B) {
	ctx := context.Background()

	client, err := ristretto.NewCache(&ristretto.Config[string, []byte]{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	store := NewRistretto(client, lib_store.WithSynchronousSet())

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

func BenchmarkRistrettoGet(b *testing.B) {
	ctx := context.Background()

	client, err := ristretto.NewCache(&ristretto.Config[string, []byte]{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	store := NewRistretto(client)

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
