package store

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/coocood/freecache"
)

func BenchmarkFreecacheSet(b *testing.B) {
	c := freecache.NewCache(1000)
	opts := &Options{
		Expiration: 10 * time.Second,
	}
	freecacheStore := NewFreecache(c, opts)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				key := fmt.Sprintf("test-%d", n)
				value := []byte(fmt.Sprintf("value-%d", n))

				_ = freecacheStore.Set(key, value, nil)
			}
		})
	}
}

func BenchmarkFreecacheGet(b *testing.B) {
	c := freecache.NewCache(1000)
	opts := &Options{
		Expiration: 10 * time.Second,
	}
	freecacheStore := NewFreecache(c, opts)
	key := "test"
	value := []byte("value")

	err := freecacheStore.Set(key, value, nil)
	if err != nil {
		b.Error(err)
	}

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				_, _ = freecacheStore.Get(key)
			}
		})
	}
}
