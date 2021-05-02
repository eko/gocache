package store

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

func BenchmarkMemcacheSet(b *testing.B) {
	store := NewMemcache(
		memcache.New("127.0.0.1:11211"),
		&Options{
			Expiration: 100 * time.Second,
		},
	)

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

func BenchmarkMemcacheGet(b *testing.B) {
	store := NewMemcache(
		memcache.New("127.0.0.1:11211"),
		&Options{
			Expiration: 100 * time.Second,
		},
	)

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
