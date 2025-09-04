package etcd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/spf13/cast"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var errNotFound = errors.New("ErrNotFound")

const (
	// EtcdType represents the storage type as a string value
	EtcdType = "etcd"
	// EtcdPrefix
	EtcdPrefix = "/gocache_etcd"
	// EtcdTagPattern represents the tag pattern to be used as a key in specified storage
	EtcdTagPattern = "gocache_tag_%s"
)

// EventHandler
type EventHandler func(*clientv3.Event)

// EtcdStore is a store for etcd
type EtcdStore struct {
	ctx         context.Context
	ctxCancelFn context.CancelFunc
	client      *clientv3.Client
	options     *lib_store.Options
	onPut       atomic.Value
	onDelete    atomic.Value
}

// NewEtcd creates a new store to etcd instance(s)
func NewEtcd(client *clientv3.Client, options ...lib_store.Option) *EtcdStore {
	ctx, cancel := context.WithCancel(client.Ctx())
	s := &EtcdStore{
		ctx:         ctx,
		ctxCancelFn: cancel,
		client:      client,
		options:     lib_store.ApplyOptions(options...),
	}
	go s.processWatch()
	return s
}

// Get returns data stored from a given key
func (s *EtcdStore) Get(ctx context.Context, key any) (any, error) {
	resp, err := s.client.Get(ctx, s.keys(key.(string)))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, lib_store.NotFoundWithCause(fmt.Errorf("key not found"))
	}
	return string(resp.Kvs[0].Value), nil
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *EtcdStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	resp, err := s.client.Get(ctx, s.keys(key.(string)))
	if err != nil {
		return nil, 0, err
	}
	if len(resp.Kvs) == 0 {
		return nil, 0, lib_store.NotFoundWithCause(fmt.Errorf("key not found"))
	}

	leaseID := resp.Kvs[0].Lease
	var ttl time.Duration
	if leaseID != 0 {
		grantResp, err := s.client.Grant(ctx, 1) // Grant a dummy lease to get the remaining TTL
		if err != nil {
			return nil, 0, err
		}
		ttlResp, err := s.client.TimeToLive(ctx, grantResp.ID)
		if err != nil {
			return nil, 0, err
		}
		ttl = time.Duration(ttlResp.TTL) * time.Second
	}

	return cast.ToString(resp.Kvs[0].Value), ttl, nil
}

// Set defines data in etcd for given key identifier
func (s *EtcdStore) Set(ctx context.Context, key any, value any, options ...lib_store.Option) error {
	opts := lib_store.ApplyOptionsWithDefault(s.options, options...)
	var valueStr string
	switch v := value.(type) {
	case string:
		valueStr = v
	case []byte:
		valueStr = cast.ToString(v)
	default:
		return errors.New("unsupported value type: must be string or []byte")
	}

	var leaseID clientv3.LeaseID
	if opts.Expiration > 0 {
		grantResp, err := s.client.Grant(ctx, int64(opts.Expiration.Seconds()))
		if err != nil {
			return err
		}
		leaseID = grantResp.ID
	}

	_, err := s.client.Put(ctx, s.keys(key.(string)), valueStr, clientv3.WithLease(leaseID))
	if err != nil {
		return err
	}

	if tags := opts.Tags; len(tags) > 0 {
		if ttl := opts.TagsTTL; ttl == 0 {
			s.setTags(ctx, key, tags)
		} else {
			s.setTagsWithTTL(ctx, key, tags, ttl)
		}
	}

	return nil
}

func (s *EtcdStore) setTags(ctx context.Context, key any, tags []string) {
	ttl := time.Hour * 720
	for _, tag := range tags {
		tagKey := s.keys(fmt.Sprintf(EtcdTagPattern, tag))
		members, err := s.getMembers(ctx, tagKey)
		if err != nil && !errors.Is(err, errNotFound) {
			continue
		}

		membersNew := make([]string, 0, len(members)+1)
		for _, k := range members {
			if k == key.(string) || k == "" {
				continue
			}
			membersNew = append(membersNew, k)
		}
		membersNew = append(membersNew, key.(string))
		_, err = s.client.Put(ctx, tagKey, fmt.Sprintf("%v", membersNew))
		if err != nil {
			continue
		}

		grantResp, err := s.client.Grant(ctx, int64(ttl.Seconds()))
		if err != nil {
			continue
		}
		s.client.Put(ctx, tagKey, fmt.Sprintf("%v", membersNew), clientv3.WithLease(grantResp.ID))
	}
}

func (s *EtcdStore) setTagsWithTTL(ctx context.Context, key any, tags []string, ttl time.Duration) {
	for _, tag := range tags {
		tagKey := s.keys(fmt.Sprintf(EtcdTagPattern, tag))
		members, err := s.getMembers(ctx, tagKey)
		if err != nil && !errors.Is(err, errNotFound) {
			continue
		}
		membersNew := make([]string, 0, len(members)+1)
		for _, k := range members {
			if k == key.(string) || k == "" {
				continue
			}
			membersNew = append(membersNew, k)
		}
		membersNew = append(membersNew, key.(string))
		_, err = s.client.Put(ctx, tagKey, fmt.Sprintf("%v", membersNew))
		if err != nil {
			continue
		}

		grantResp, err := s.client.Grant(ctx, int64(ttl.Seconds()))
		if err != nil {
			continue
		}
		s.client.Put(ctx, tagKey, fmt.Sprintf("%v", membersNew), clientv3.WithLease(grantResp.ID))
	}
}

func (s *EtcdStore) getMembers(ctx context.Context, key string) ([]string, error) {
	resp, err := s.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, errNotFound
	}

	value := string(resp.Kvs[0].Value)
	if len(value) < 2 || value[0] != '[' || value[len(value)-1] != ']' {
		return nil, fmt.Errorf("invalid format: expected a string in the form [item1,item2,...]")
	}

	cleanedInput := value[1 : len(value)-1]
	members := strings.Split(cleanedInput, ",")
	for i, member := range members {
		members[i] = strings.TrimSpace(member)
	}
	return members, nil
}

// Delete removes data from etcd for given key identifier
func (s *EtcdStore) Delete(ctx context.Context, key any) error {
	_, err := s.client.Delete(ctx, s.keys(key.(string)))
	return err
}

// Invalidate invalidates some cache data in etcd for given options
func (s *EtcdStore) Invalidate(ctx context.Context, options ...lib_store.InvalidateOption) error {
	opts := lib_store.ApplyInvalidateOptions(options...)

	for _, tag := range opts.Tags {
		tagKey := fmt.Sprintf(EtcdTagPattern, tag)
		cacheKeys, err := s.getMembers(ctx, s.keys(tagKey))
		if err != nil && !errors.Is(err, errNotFound) {
			continue
		}

		for _, cacheKey := range cacheKeys {
			s.Delete(ctx, cacheKey)
		}
		s.Delete(ctx, tagKey)
	}
	return nil
}

// GetType returns the store type
func (s *EtcdStore) GetType() string {
	return EtcdType
}

func (s *EtcdStore) keys(key ...string) string {
	newKeys := make([]string, 0, len(key)+1)
	newKeys = append(newKeys, EtcdPrefix)
	for _, k := range key {
		newKeys = append(newKeys, strings.Trim(k, "/"))
	}
	return strings.Join(newKeys, "/")
}

// Clear resets all data in the store
func (s *EtcdStore) Clear(ctx context.Context) error {
	_, err := s.client.Delete(ctx, EtcdPrefix+"/", clientv3.WithPrefix())
	return err
}

func (s *EtcdStore) Close() {
	if s.ctxCancelFn == nil {
		return
	}
	s.ctxCancelFn()
}

func (s *EtcdStore) OnPut(fn EventHandler) {
	s.onPut.Store(fn)
}

func (s *EtcdStore) OnDelete(fn EventHandler) {
	s.onDelete.Store(fn)
}

func (s *EtcdStore) processWatch() {
	wctx, wcf := context.WithCancel(s.ctx)
	watcher := clientv3.NewWatcher(s.client)

	defer func() {
		watcher.Close()
		wcf()
	}()

	wch := watcher.Watch(wctx, EtcdPrefix, clientv3.WithPrefix(), clientv3.WithRev(0))
	for {
		select {
		case wr := <-wch:
			if wr.Canceled {
				return
			}

			for _, event := range wr.Events {
				switch event.Type {
				case clientv3.EventTypePut:
					if fn, ok := s.onPut.Load().(EventHandler); ok && fn != nil {
						fn(event)
					}

				case clientv3.EventTypeDelete:
					if fn, ok := s.onDelete.Load().(EventHandler); ok && fn != nil {
						fn(event)
					}
				}
			}

		case <-wctx.Done():
			return

		case <-s.ctx.Done():
			return
		}
	}
}
