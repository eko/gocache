package nats

import (
	"context"
	"errors"
	"testing"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeNatsClient struct {
	getEntry  jetstream.KeyValueEntry
	getErr    error
	putKey    string
	putValue  []byte
	putErr    error
	createKey string
	createVal []byte
	createErr error
	deleteErr map[string]error
	deleted   []string
	keys      []string
	keysErr   error
	status    jetstream.KeyValueStatus
	statusErr error
}

func (f *fakeNatsClient) Get(context.Context, string) (jetstream.KeyValueEntry, error) {
	return f.getEntry, f.getErr
}

func (f *fakeNatsClient) Put(_ context.Context, key string, value []byte) (uint64, error) {
	f.putKey, f.putValue = key, value
	return 1, f.putErr
}

func (f *fakeNatsClient) Create(_ context.Context, key string, value []byte, _ ...jetstream.KVCreateOpt) (uint64, error) {
	f.createKey, f.createVal = key, value
	return 1, f.createErr
}

func (f *fakeNatsClient) Delete(_ context.Context, key string, _ ...jetstream.KVDeleteOpt) error {
	f.deleted = append(f.deleted, key)
	return f.deleteErr[key]
}

func (f *fakeNatsClient) Keys(context.Context, ...jetstream.WatchOpt) ([]string, error) {
	return f.keys, f.keysErr
}

func (f *fakeNatsClient) Status(context.Context) (jetstream.KeyValueStatus, error) {
	return f.status, f.statusErr
}

type fakeEntry struct {
	value   []byte
	created time.Time
}

func (e fakeEntry) Bucket() string                  { return "cache" }
func (e fakeEntry) Key() string                     { return "key" }
func (e fakeEntry) Value() []byte                   { return e.value }
func (e fakeEntry) Revision() uint64                { return 1 }
func (e fakeEntry) Created() time.Time              { return e.created }
func (e fakeEntry) Delta() uint64                   { return 0 }
func (e fakeEntry) Operation() jetstream.KeyValueOp { return jetstream.KeyValuePut }

type fakeStatus struct{ ttl time.Duration }

func (s fakeStatus) Bucket() string                   { return "cache" }
func (s fakeStatus) Values() uint64                   { return 1 }
func (s fakeStatus) History() int64                   { return 1 }
func (s fakeStatus) TTL() time.Duration               { return s.ttl }
func (s fakeStatus) BackingStore() string             { return "JetStream" }
func (s fakeStatus) Bytes() uint64                    { return 1 }
func (s fakeStatus) IsCompressed() bool               { return false }
func (s fakeStatus) LimitMarkerTTL() time.Duration    { return 0 }
func (s fakeStatus) Metadata() map[string]string      { return nil }
func (s fakeStatus) Config() jetstream.KeyValueConfig { return jetstream.KeyValueConfig{} }

func TestGet(t *testing.T) {
	client := &fakeNatsClient{getEntry: fakeEntry{value: []byte("value")}}
	value, err := NewNats(client).Get(context.Background(), "key")
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), value)
}

func TestGetMapsMissingAndDeletedKeysToNotFound(t *testing.T) {
	for _, sourceErr := range []error{jetstream.ErrKeyNotFound, jetstream.ErrKeyDeleted} {
		client := &fakeNatsClient{getErr: sourceErr}
		value, err := NewNats(client).Get(context.Background(), "missing")
		assert.Nil(t, value)
		assert.ErrorIs(t, err, lib_store.NotFound{})
		assert.ErrorIs(t, err, sourceErr)
	}
}

func TestGetReturnsClientError(t *testing.T) {
	wantErr := errors.New("get failed")
	value, err := NewNats(&fakeNatsClient{getErr: wantErr}).Get(context.Background(), "key")
	assert.Nil(t, value)
	assert.ErrorIs(t, err, wantErr)
}

func TestGetWithTTLReturnsRemainingBucketTTL(t *testing.T) {
	client := &fakeNatsClient{
		getEntry: fakeEntry{value: []byte("value"), created: time.Now().Add(-time.Minute)},
		status:   fakeStatus{ttl: 5 * time.Minute},
	}
	value, ttl, err := NewNats(client).GetWithTTL(context.Background(), "key")
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), value)
	assert.InDelta(t, (4 * time.Minute).Seconds(), ttl.Seconds(), 1)
}

func TestGetWithTTLPropagatesStatusError(t *testing.T) {
	wantErr := errors.New("status failed")
	client := &fakeNatsClient{
		getEntry:  fakeEntry{value: []byte("value")},
		statusErr: wantErr,
	}
	value, ttl, err := NewNats(client).GetWithTTL(context.Background(), "key")
	assert.Nil(t, value)
	assert.Zero(t, ttl)
	assert.ErrorIs(t, err, wantErr)
}

func TestSetAcceptsStringsAndBytes(t *testing.T) {
	for _, value := range []any{"value", []byte("value")} {
		client := &fakeNatsClient{}
		err := NewNats(client).Set(context.Background(), "key", value)
		require.NoError(t, err)
		assert.Equal(t, "key", client.putKey)
		assert.Equal(t, []byte("value"), client.putValue)
	}
}

func TestSetRejectsUnsupportedValue(t *testing.T) {
	err := NewNats(&fakeNatsClient{}).Set(context.Background(), "key", 42)
	assert.EqualError(t, err, "value type int is not supported by NATS store")
}

func TestSetIfNotExists(t *testing.T) {
	client := &fakeNatsClient{}
	written, err := NewNats(client).SetIfNotExists(context.Background(), "key", "value")
	require.NoError(t, err)
	assert.True(t, written)
	assert.Equal(t, "key", client.createKey)
	assert.Equal(t, []byte("value"), client.createVal)
}

func TestSetIfNotExistsReportsExistingKey(t *testing.T) {
	client := &fakeNatsClient{createErr: jetstream.ErrKeyExists}
	written, err := NewNats(client).SetIfNotExists(context.Background(), "key", "value")
	require.NoError(t, err)
	assert.False(t, written)
}

func TestDeleteIsIdempotentForMissingKeys(t *testing.T) {
	client := &fakeNatsClient{deleteErr: map[string]error{"missing": jetstream.ErrKeyNotFound}}
	assert.NoError(t, NewNats(client).Delete(context.Background(), "missing"))
}

func TestInvalidateRejectsTags(t *testing.T) {
	store := NewNats(&fakeNatsClient{})
	assert.NoError(t, store.Invalidate(context.Background()))
	err := store.Invalidate(context.Background(), lib_store.WithInvalidateTags([]string{"users"}))
	assert.ErrorIs(t, err, lib_store.ErrNotSupported)
}

func TestClearDeletesEveryKey(t *testing.T) {
	client := &fakeNatsClient{keys: []string{"one", "two"}}
	require.NoError(t, NewNats(client).Clear(context.Background()))
	assert.Equal(t, []string{"one", "two"}, client.deleted)
}

func TestClearHandlesEmptyBucket(t *testing.T) {
	client := &fakeNatsClient{keysErr: jetstream.ErrNoKeysFound}
	assert.NoError(t, NewNats(client).Clear(context.Background()))
}

func TestClearStopsOnDeleteError(t *testing.T) {
	wantErr := errors.New("delete failed")
	client := &fakeNatsClient{
		keys:      []string{"one", "two"},
		deleteErr: map[string]error{"one": wantErr},
	}
	err := NewNats(client).Clear(context.Background())
	assert.ErrorIs(t, err, wantErr)
	assert.Equal(t, []string{"one"}, client.deleted)
}

func TestGetType(t *testing.T) {
	assert.Equal(t, NatsType, NewNats(&fakeNatsClient{}).GetType())
}
