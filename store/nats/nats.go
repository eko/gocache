package nats

import (
	"context"
	"errors"
	"fmt"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/nats-io/nats.go/jetstream"
)

// NatsType represents the storage type as a string value.
const NatsType = "nats"

// NatsClientInterface contains the JetStream KeyValue operations used by the
// store. A jetstream.KeyValue satisfies this interface directly.
type NatsClientInterface interface {
	Get(ctx context.Context, key string) (jetstream.KeyValueEntry, error)
	Put(ctx context.Context, key string, value []byte) (uint64, error)
	Create(ctx context.Context, key string, value []byte, opts ...jetstream.KVCreateOpt) (uint64, error)
	Delete(ctx context.Context, key string, opts ...jetstream.KVDeleteOpt) error
	Keys(ctx context.Context, opts ...jetstream.WatchOpt) ([]string, error)
	Status(ctx context.Context) (jetstream.KeyValueStatus, error)
}

var _ NatsClientInterface = (jetstream.KeyValue)(nil)

// NatsStore is a cache store backed by a NATS JetStream key-value bucket.
type NatsStore struct {
	client NatsClientInterface
}

var _ lib_store.StoreInterface = (*NatsStore)(nil)
var _ lib_store.SetIfNotExistsStore = (*NatsStore)(nil)

// NewNats creates a store backed by the given JetStream key-value bucket.
func NewNats(client NatsClientInterface) *NatsStore {
	return &NatsStore{client: client}
}

// Get returns data stored for a given key.
func (s *NatsStore) Get(ctx context.Context, key any) (any, error) {
	entry, err := s.client.Get(ctx, key.(string))
	if isNotFound(err) {
		return nil, lib_store.NotFoundWithCause(err)
	}
	if err != nil {
		return nil, err
	}

	return entry.Value(), nil
}

// GetWithTTL returns data stored for a given key and the remaining bucket TTL.
// NATS configures expiration on the bucket rather than on individual Put calls.
func (s *NatsStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	entry, err := s.client.Get(ctx, key.(string))
	if isNotFound(err) {
		return nil, 0, lib_store.NotFoundWithCause(err)
	}
	if err != nil {
		return nil, 0, err
	}

	status, err := s.client.Status(ctx)
	if err != nil {
		return nil, 0, err
	}

	ttl := status.TTL()
	if ttl > 0 {
		ttl -= time.Since(entry.Created())
		if ttl < 0 {
			ttl = 0
		}
	}

	return entry.Value(), ttl, nil
}

// Set stores a value for a given key.
func (s *NatsStore) Set(ctx context.Context, key any, value any, _ ...lib_store.Option) error {
	data, err := toBytes(value)
	if err != nil {
		return err
	}

	_, err = s.client.Put(ctx, key.(string), data)
	return err
}

// SetIfNotExists stores a value only if the key does not already exist.
func (s *NatsStore) SetIfNotExists(ctx context.Context, key any, value any, _ ...lib_store.Option) (bool, error) {
	data, err := toBytes(value)
	if err != nil {
		return false, err
	}

	_, err = s.client.Create(ctx, key.(string), data)
	if errors.Is(err, jetstream.ErrKeyExists) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// Delete removes data for a given key.
func (s *NatsStore) Delete(ctx context.Context, key any) error {
	err := s.client.Delete(ctx, key.(string))
	if isNotFound(err) {
		return nil
	}
	return err
}

// Invalidate reports that tag-based invalidation is not available in NATS KV.
func (s *NatsStore) Invalidate(_ context.Context, options ...lib_store.InvalidateOption) error {
	if len(lib_store.ApplyInvalidateOptions(options...).Tags) == 0 {
		return nil
	}
	return fmt.Errorf("NATS KV does not support tag invalidation: %w", lib_store.ErrNotSupported)
}

// Clear removes all current keys from the bucket.
func (s *NatsStore) Clear(ctx context.Context) error {
	keys, err := s.client.Keys(ctx)
	if errors.Is(err, jetstream.ErrNoKeysFound) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := s.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// GetType returns the store type.
func (s *NatsStore) GetType() string {
	return NatsType
}

func isNotFound(err error) bool {
	return errors.Is(err, jetstream.ErrKeyNotFound) || errors.Is(err, jetstream.ErrKeyDeleted)
}

func toBytes(value any) ([]byte, error) {
	switch value := value.(type) {
	case string:
		return []byte(value), nil
	case []byte:
		return value, nil
	default:
		return nil, fmt.Errorf("value type %T is not supported by NATS store", value)
	}
}
