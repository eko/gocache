package hazelcast

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/types"
	"golang.org/x/sync/errgroup"
)

// HazelcastMapInterface represents a hazelcast/hazelcast-go-client map
type HazelcastMapInterface interface {
	Get(ctx context.Context, key any) (any, error)
	GetEntryView(ctx context.Context, key any) (*types.SimpleEntryView, error)
	SetWithTTL(ctx context.Context, key any, value any, ttl time.Duration) error
	Remove(ctx context.Context, key any) (any, error)
	Clear(ctx context.Context) error
}

type HazelcastMapInterfaceProvider func(ctx context.Context) (HazelcastMapInterface, error)

const (
	// HazelcastType represents the storage type as a string value
	HazelcastType = "hazelcast"
	// HazelcastTagPattern represents the tag pattern to be used as a key in specified storage
	HazelcastTagPattern = "gocache_tag_%s"

	TagKeyExpiry = 720 * time.Hour
)

// HazelcastStore is a store for Hazelcast
type HazelcastStore struct {
	mapProvider HazelcastMapInterfaceProvider
	options     *lib_store.Options
}

// NewHazelcast creates a new store to Hazelcast instance(s)
func NewHazelcast(hzClient *hazelcast.Client, mapName string, options ...lib_store.Option) *HazelcastStore {
	return &HazelcastStore{
		mapProvider: func(ctx context.Context) (HazelcastMapInterface, error) {
			return hzClient.GetMap(ctx, mapName)
		},
		options: lib_store.ApplyOptions(options...),
	}
}

// newHazelcast creates a new store with given HazelcastMapInterface for test purpose
func newHazelcast(hzMap HazelcastMapInterface, options ...lib_store.Option) *HazelcastStore {
	return &HazelcastStore{
		mapProvider: func(ctx context.Context) (HazelcastMapInterface, error) {
			return hzMap, nil
		},
		options: lib_store.ApplyOptions(options...),
	}
}

// Get returns data stored from a given key
func (s *HazelcastStore) Get(ctx context.Context, key any) (any, error) {
	hzMap, err := s.mapProvider(ctx)
	if err != nil {
		return nil, err
	}
	value, err := hzMap.Get(ctx, key)
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
	hzMap, err := s.mapProvider(ctx)
	if err != nil {
		return nil, 0, err
	}
	entryView, err := hzMap.GetEntryView(ctx, key)
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
	hzMap, err := s.mapProvider(ctx)
	if err != nil {
		return err
	}
	err = hzMap.SetWithTTL(ctx, key, value, opts.Expiration)
	if err != nil {
		return err
	}
	if tags := opts.Tags; len(tags) > 0 {
		s.setTags(ctx, hzMap, key, tags)
	}
	return nil
}

func (s *HazelcastStore) setTags(ctx context.Context, hzMap HazelcastMapInterface, key any, tags []string) {
	group, ctx := errgroup.WithContext(ctx)
	for _, tag := range tags {
		currentTag := tag
		group.Go(func() error {
			tagKey := fmt.Sprintf(HazelcastTagPattern, currentTag)
			tagValue, err := hzMap.Get(ctx, tagKey)
			if err != nil {
				return err
			}
			if tagValue == nil {
				return hzMap.SetWithTTL(ctx, tagKey, key.(string), TagKeyExpiry)
			}
			cacheKeys := strings.Split(tagValue.(string), ",")
			for _, cacheKey := range cacheKeys {
				if key == cacheKey {
					return nil
				}
			}
			cacheKeys = append(cacheKeys, key.(string))
			newTagValue := strings.Join(cacheKeys, ",")
			return hzMap.SetWithTTL(ctx, tagKey, newTagValue, TagKeyExpiry)
		})
	}
	group.Wait()
}

// Delete removes data from Hazelcast for given key identifier
func (s *HazelcastStore) Delete(ctx context.Context, key any) error {
	hzMap, err := s.mapProvider(ctx)
	if err != nil {
		return err
	}
	_, err = hzMap.Remove(ctx, key)
	return err
}

// Invalidate invalidates some cache data in Hazelcast for given options
func (s *HazelcastStore) Invalidate(ctx context.Context, options ...lib_store.InvalidateOption) error {
	opts := lib_store.ApplyInvalidateOptions(options...)
	if tags := opts.Tags; len(tags) > 0 {
		hzMap, err := s.mapProvider(ctx)
		if err != nil {
			return err
		}
		for _, tag := range tags {
			tagKey := fmt.Sprintf(HazelcastTagPattern, tag)
			tagValue, err := hzMap.Get(ctx, tagKey)
			if err != nil || tagValue == nil {
				continue
			}
			cacheKeys := strings.Split(tagValue.(string), ",")
			for _, cacheKey := range cacheKeys {
				hzMap.Remove(ctx, cacheKey)
			}
			hzMap.Remove(ctx, tagKey)
		}
	}
	return nil
}

// Clear resets all data in the store
func (s *HazelcastStore) Clear(ctx context.Context) error {
	hzMap, err := s.mapProvider(ctx)
	if err != nil {
		return err
	}
	return hzMap.Clear(ctx)
}

// GetType returns the store type
func (s *HazelcastStore) GetType() string {
	return HazelcastType
}
