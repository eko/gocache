package store

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"math"
	"testing"
	"time"
)

func BenchmarkGoCacheSet(b *testing.B) {
	client := cache.New(10*time.Second, 30*time.Second)

	store := NewGoCache(client, nil)

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

func BenchmarkGoCacheGet(b *testing.B) {
	client := cache.New(10*time.Second, 30*time.Second)

	store := NewGoCache(client, nil)

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
