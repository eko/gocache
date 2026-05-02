package hazelcast

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/hazelcast/hazelcast-go-client/types"
	"golang.org/x/sync/errgroup"
)

// HazelcastMapInterface represents a hazelcast/hazelcast-go-client map
type HazelcastMapInterface interface {
	Get(ctx context.Context, key any) (any, error)
	GetEntryView(ctx context.Context, key any) (*types.SimpleEntryView, error)
	SetWithTTL(ctx context.Context, key any, value any, ttl time.Duration) error
	SetTTL(ctx context.Context, key any, ttl time.Duration) error
	PutIfAbsentWithTTL(ctx context.Context, key any, value any, ttl time.Duration) (any, error)
	ReplaceIfSame(ctx context.Context, key any, oldValue any, newValue any) (bool, error)
	Remove(ctx context.Context, key any) (any, error)
	Clear(ctx context.Context) error
}

const (
	// HazelcastType represents the storage type as a string value
	HazelcastType = "hazelcast"
	// HazelcastTagPattern represents the tag pattern to be used as a key in specified storage
	HazelcastTagPattern = "gocache_tag_%s"
)

// HazelcastStore is a store for Hazelcast
type HazelcastStore struct {
	hzMap   HazelcastMapInterface
	options *lib_store.Options
}

// NewHazelcast creates a new store to Hazelcast instance(s)
func NewHazelcast(hzMap HazelcastMapInterface, options ...lib_store.Option) *HazelcastStore {
	return &HazelcastStore{
		hzMap:   hzMap,
		options: lib_store.ApplyOptions(options...),
	}
}

// Get returns data stored from a given key
func (s *HazelcastStore) Get(ctx context.Context, key any) (any, error) {
	value, err := s.hzMap.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, lib_store.NotFoundWithCause(errors.New("unable to retrieve data from hazelcast"))
	}
	return value, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *HazelcastStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	entryView, err := s.hzMap.GetEntryView(ctx, key)
	if err != nil {
		return nil, 0, err
	}
	if entryView == nil {
		return nil, 0, lib_store.NotFoundWithCause(errors.New("unable to retrieve data from hazelcast"))
	}
	return entryView.Value, time.Duration(entryView.TTL) * time.Millisecond, err
}

// Set defines data in Hazelcast for given key identifier
func (s *HazelcastStore) Set(ctx context.Context, key any, value any, options ...lib_store.Option) error {
	opts := lib_store.ApplyOptionsWithDefault(s.options, options...)
	err := s.hzMap.SetWithTTL(ctx, key, value, opts.Expiration)
	if err != nil {
		return err
	}
	if tags := opts.Tags; len(tags) > 0 {
		s.setTags(ctx, s.hzMap, key, tags, opts.TagsTTL)
	}
	return nil
}

func (s *HazelcastStore) setTags(ctx context.Context, hzMap HazelcastMapInterface, key any, tags []string, ttl time.Duration) {
	group, ctx := errgroup.WithContext(ctx)
	for _, tag := range tags {
		currentTag := tag
		group.Go(func() error {
			tagKey := fmt.Sprintf(HazelcastTagPattern, currentTag)

			var err error
			for i := 0; i < 3; i++ {
				if err = s.addKeyToTagValue(ctx, hzMap, tagKey, key, ttl); err == nil {
					return nil
				}
				// loop to retry any failure (including race conditions)
			}
			return err
		})
	}
	group.Wait()
}

func (s *HazelcastStore) addKeyToTagValue(ctx context.Context, hzMap HazelcastMapInterface, tagKey string, key any, ttl time.Duration) error {
	tagValue, err := hzMap.Get(ctx, tagKey)
	if err != nil {
		return err
	}

	if tagValue == nil {
		// first writer: try to insert atomically
		prev, err := hzMap.PutIfAbsentWithTTL(ctx, tagKey, key.(string), ttl)
		if err != nil {
			return err
		}
		if prev == nil {
			// PutIfAbsent returns the existing value or nil if absent;
			// nil here means our insert succeeded.
			return nil
		}
		// somebody else inserted; fall through to update path with the new value
		tagValue = prev
	}

	oldStr := tagValue.(string)
	cacheKeys := strings.Split(oldStr, ",")
	if slices.Contains(cacheKeys, key.(string)) {
		return hzMap.SetTTL(ctx, tagKey, ttl)
	}
	newStr := strings.Join(append(cacheKeys, key.(string)), ",")

	ok, err := hzMap.ReplaceIfSame(ctx, tagKey, oldStr, newStr)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("hazelcast tag key contended")
	}

	return hzMap.SetTTL(ctx, tagKey, ttl)
}

// Delete removes data from Hazelcast for given key identifier
func (s *HazelcastStore) Delete(ctx context.Context, key any) error {
	_, err := s.hzMap.Remove(ctx, key)
	return err
}

// Invalidate invalidates some cache data in Hazelcast for given options
func (s *HazelcastStore) Invalidate(ctx context.Context, options ...lib_store.InvalidateOption) error {
	opts := lib_store.ApplyInvalidateOptions(options...)
	if tags := opts.Tags; len(tags) > 0 {
		for _, tag := range tags {
			tagKey := fmt.Sprintf(HazelcastTagPattern, tag)
			tagValue, err := s.hzMap.Get(ctx, tagKey)
			if err != nil || tagValue == nil {
				continue
			}
			cacheKeys := strings.Split(tagValue.(string), ",")
			for _, cacheKey := range cacheKeys {
				s.hzMap.Remove(ctx, cacheKey)
			}
			s.hzMap.Remove(ctx, tagKey)
		}
	}
	return nil
}

// Clear resets all data in the store
func (s *HazelcastStore) Clear(ctx context.Context) error {
	return s.hzMap.Clear(ctx)
}

// GetType returns the store type
func (s *HazelcastStore) GetType() string {
	return HazelcastType
}
