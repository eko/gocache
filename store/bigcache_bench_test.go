package store

import (
	"fmt"
	"math"
	"testing"
	time "time"

	"github.com/allegro/bigcache"
)

func BenchmarkBigcacheSet(b *testing.B) {
	client, _ := bigcache.NewBigCache(bigcache.DefaultConfig(5 * time.Minute))
	store := NewBigcache(client, nil)

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

func BenchmarkBigcacheGet(b *testing.B) {
	client, _ := bigcache.NewBigCache(bigcache.DefaultConfig(5 * time.Minute))
	store := NewBigcache(client, nil)

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
