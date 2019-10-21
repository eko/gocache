package store

import (
	"fmt"
	"math"
	"testing"

	"github.com/dgraph-io/ristretto"
)

func BenchmarkRistrettoSet(b *testing.B) {
	client, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	store := NewRistretto(client, nil)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				key := fmt.Sprintf("test-%d", n)
				value := []byte(fmt.Sprintf("value-%d", n))

				store.Set(key, value, &Options{
					Tags: []string{fmt.Sprintf("tag-%d", n)},
				})
			}
		})
	}
}

func BenchmarkRistrettoGet(b *testing.B) {
	client, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	store := NewRistretto(client, nil)

	key := "test"
	value := []byte("value")

	store.Set(key, value, nil)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				_, _ = store.Get(key)
			}
		})
	}
}
