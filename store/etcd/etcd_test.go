package etcd

import (
	"context"
	"testing"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// run go test -run='TestEtcd*' -race -cover -coverprofile=coverage.txt -covermode=atomic -v ./...
func testGetEtcdClicent() (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints: []string{
			"192.168.0.127:12379",
			"192.168.0.127:22379",
			"192.168.0.127:32379",
		},
	})
}
func TestEtcdSet(t *testing.T) {
	ctx := context.Background()
	etcdClient, err := testGetEtcdClicent()
	if err != nil {
		t.Fatal(err)
	}

	cacheKey := "my-key"
	cacheValue := "my-cache-value"
	store := NewEtcd(etcdClient, lib_store.WithExpiration(6*time.Second))
	store.OnPut(func(evt *clientv3.Event) {
		t.Log("evt:", evt.Type, string(evt.Kv.Key))
	})

	// When
	err = store.Set(ctx, cacheKey, cacheValue, lib_store.WithExpiration(5*time.Second))

	// Then
	assert.Nil(t, err)
}

func TestEtcdDelete(t *testing.T) {
	ctx := context.Background()
	etcdClient, err := testGetEtcdClicent()
	if err != nil {
		t.Fatal(err)
	}

	cacheKey := "my-key"
	store := NewEtcd(etcdClient, lib_store.WithExpiration(6*time.Second))
	store.OnPut(func(evt *clientv3.Event) {
		t.Log("evt:", evt.Type, string(evt.Kv.Key))
	})

	// When
	err = store.Delete(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
}
