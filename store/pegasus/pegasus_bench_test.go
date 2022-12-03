package pegasus

import (
	"context"
	"fmt"
	"math"
	"testing"

	lib_store "github.com/eko/gocache/lib/v4/store"
)

// run go test -bench='BenchmarkPegasusStore*' -benchtime=1s -count=1 -run=none
func BenchmarkPegasusStore_Set(b *testing.B) {
	ctx := context.Background()

	p, _ := NewPegasus(ctx, testPegasusOptions())
	defer p.Close()

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				key := fmt.Sprintf("test-%d", n)
				value := []byte(fmt.Sprintf("value-%d", n))

				p.Set(ctx, key, value, lib_store.WithTags([]string{fmt.Sprintf("tag-%d", n)}))
			}
		})
	}
}

func BenchmarkPegasusStore_Get(b *testing.B) {
	ctx := context.Background()

	p, _ := NewPegasus(ctx, testPegasusOptions())
	defer p.Close()

	key := "test"
	value := []byte("value")

	p.Set(ctx, key, value, nil)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				_, _ = p.Get(ctx, key)
			}
		})
	}
}
