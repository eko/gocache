package valkey

import (
	"context"
	"fmt"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeycompat"
)

const (
	// ValkeyType represents the storage type as a string value
	ValkeyType = "valkey"
	// ValkeyTagPattern represents the tag pattern to be used as a key in specified storage
	ValkeyTagPattern = "gocache_tag_%s"

	defaultClientSideCacheExpiration = 10 * time.Second
)

// ValkeyStore is a store for Valkey
type ValkeyStore struct {
	client  valkey.Client
	options *lib_store.Options
}

// NewValkey creates a new store to Valkey instance(s)
func NewValkey(client valkey.Client, options ...lib_store.Option) *ValkeyStore {
	// defaults client side cache expiration to 10s
	appliedOptions := lib_store.ApplyOptions(options...)

	if appliedOptions.ClientSideCacheExpiration == 0 {
		appliedOptions.ClientSideCacheExpiration = defaultClientSideCacheExpiration
	}

	return &ValkeyStore{
		client:  client,
		options: appliedOptions,
	}
}

// Get returns data stored from a given key
func (s *ValkeyStore) Get(ctx context.Context, key any) (any, error) {
	cmd := s.client.B().Get().Key(key.(string)).Cache()
	res := s.client.DoCache(ctx, cmd, s.options.ClientSideCacheExpiration)
	str, err := res.ToString()
	if valkey.IsValkeyNil(err) {
		err = lib_store.NotFoundWithCause(err)
	}
	return str, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *ValkeyStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	cmd := s.client.B().Get().Key(key.(string)).Cache()
	res := s.client.DoCache(ctx, cmd, s.options.ClientSideCacheExpiration)
	str, err := res.ToString()
	if valkey.IsValkeyNil(err) {
		err = lib_store.NotFoundWithCause(err)
	}
	return str, time.Duration(res.CacheTTL()) * time.Second, err
}

// Set defines data in Valkey for given key identifier
func (s *ValkeyStore) Set(ctx context.Context, key any, value any, options ...lib_store.Option) error {
	opts := lib_store.ApplyOptionsWithDefault(s.options, options...)
	ttl := int64(opts.Expiration.Seconds())
	var cmd valkey.Completed
	switch value.(type) {
	case string:
		cmd = s.client.B().Set().Key(key.(string)).Value(value.(string)).ExSeconds(ttl).Build()

	case []byte:
		cmd = s.client.B().Set().Key(key.(string)).Value(valkey.BinaryString(value.([]byte))).ExSeconds(ttl).Build()
	}
	err := s.client.Do(ctx, cmd).Error()
	if err != nil {
		return err
	}
	if tags := opts.Tags; len(tags) > 0 {
		s.setTags(ctx, key, tags)
	}

	return nil
}

func (s *ValkeyStore) setTags(ctx context.Context, key any, tags []string) {
	ttl := 720 * time.Hour
	for _, tag := range tags {
		tagKey := fmt.Sprintf(ValkeyTagPattern, tag)
		s.client.DoMulti(ctx,
			s.client.B().Sadd().Key(tagKey).Member(key.(string)).Build(),
			s.client.B().Expire().Key(tagKey).Seconds(int64(ttl.Seconds())).Build(),
		)
	}
}

// Delete removes data from Valkey for given key identifier
func (s *ValkeyStore) Delete(ctx context.Context, key any) error {
	return s.client.Do(ctx, s.client.B().Del().Key(key.(string)).Build()).Error()
}

// Invalidate invalidates some cache data in Valkey for given options
func (s *ValkeyStore) Invalidate(ctx context.Context, options ...lib_store.InvalidateOption) error {
	opts := lib_store.ApplyInvalidateOptions(options...)

	if tags := opts.Tags; len(tags) > 0 {
		for _, tag := range tags {
			tagKey := fmt.Sprintf(ValkeyTagPattern, tag)

			cacheKeys, err := s.client.Do(ctx, s.client.B().Smembers().Key(tagKey).Build()).AsStrSlice()
			if err != nil {
				continue
			}

			for _, cacheKey := range cacheKeys {
				s.Delete(ctx, cacheKey)
			}

			s.Delete(ctx, tagKey)
		}
	}

	return nil
}

// GetType returns the store type
func (s *ValkeyStore) GetType() string {
	return ValkeyType
}

// Clear resets all data in the store
func (s *ValkeyStore) Clear(ctx context.Context) error {
	return valkeycompat.NewAdapter(s.client).FlushAll(ctx).Err()
}
