package nats

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkNatsSet(b *testing.B) {
	store := NewNats(integrationBucket(b))
	ctx := context.Background()
	value := []byte("value")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := store.Set(ctx, "key", value); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNatsGet(b *testing.B) {
	store := NewNats(integrationBucket(b))
	ctx := context.Background()
	require.NoError(b, store.Set(ctx, "key", "value"))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := store.Get(ctx, "key"); err != nil {
			b.Fatal(err)
		}
	}
}
