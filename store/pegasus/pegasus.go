package pegasus

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/XiaoMi/pegasus-go-client/admin"
	"github.com/XiaoMi/pegasus-go-client/pegasus"
	"github.com/spf13/cast"

	lib_store "github.com/eko/gocache/lib/v4/store"
)

const (
	// PegasusType represents the storage type as a string value
	PegasusType = "pegasus"
	// PegasusTagPattern represents the tag pattern to be used as a key in specified storage
	PegasusTagPattern = "gocache_tag_%s"
	// Pegasus ttl(time-to-live) in seconds: -1 if ttl is not set; -2 if entry doesn't exist
	PegasusNOTTL   = -1
	PegasusNOENTRY = -2

	DefaultTable             = "gocache_pegasus"
	DefaultTablePartitionNum = 4
	DefaultScanNum           = 100
)

// empty represent empty sort key, more info reference: https://github.com/XiaoMi/pegasus-go-client/blob/f3b6b08bc4c227982bb5b73106329435fda97a38/pegasus/table_connector.go#L83
var empty = []byte("-")

// OptionsPegasus is options of Pegasus
type OptionsPegasus struct {
	*lib_store.Options
	MetaServers []string

	TableName         string
	TablePartitionNum int
	TableScanNum      int
}

// PegasusStore is a store for Pegasus
type PegasusStore struct {
	client  pegasus.Client
	options *OptionsPegasus
}

// NewPegasus creates a new store to pegasus instance(s)
func NewPegasus(ctx context.Context, options *OptionsPegasus) (*PegasusStore, error) {
	if options == nil {
		options = &OptionsPegasus{}
	}

	if err := createTable(ctx, options); err != nil {
		return nil, err
	}

	client := pegasus.NewClient(pegasus.Config{
		MetaServers: options.MetaServers,
	})
	table, err := client.OpenTable(ctx, options.TableName)
	if err != nil {
		return nil, err
	}
	defer table.Close()

	return &PegasusStore{
		client:  client,
		options: options,
	}, nil
}

// validateOptions validate pegasus options
func validateOptions(options *OptionsPegasus) error {
	if len(options.MetaServers) == 0 {
		return errors.New("pegasus meta servers must fill")
	}
	if len(options.TableName) == 0 {
		options.TableName = DefaultTable
	}
	if options.TablePartitionNum < 1 {
		options.TablePartitionNum = DefaultTablePartitionNum
	}
	if options.TableScanNum < 1 {
		options.TableScanNum = DefaultScanNum
	}

	return nil
}

// createTable for create table by options
func createTable(ctx context.Context, options *OptionsPegasus) error {
	if err := validateOptions(options); err != nil {
		return err
	}

	tableClient := admin.NewClient(admin.Config{MetaServers: options.MetaServers})
	tableList, err := tableClient.ListTables(ctx)
	if err != nil {
		return err
	}

	for i := range tableList {
		if tableList[i].Name == options.TableName {
			return nil
		}
	}

	// if not found then create table of options
	return tableClient.CreateTable(ctx, options.TableName, options.TablePartitionNum)
}

// dropTable for drop table
func dropTable(ctx context.Context, options *OptionsPegasus) error {
	if err := validateOptions(options); err != nil {
		return err
	}

	tableClient := admin.NewClient(admin.Config{MetaServers: options.MetaServers})
	return tableClient.DropTable(ctx, options.TableName)
}

// Close when exit store
func (p *PegasusStore) Close() error {
	return p.client.Close()
}

// Get returns data stored from a given key
func (p *PegasusStore) Get(ctx context.Context, key any) (any, error) {
	table, err := p.client.OpenTable(ctx, p.options.TableName)
	if err != nil {
		return nil, err
	}
	defer table.Close()

	value, err := table.Get(ctx, []byte(cast.ToString(key)), empty)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, &lib_store.NotFound{}
	}
	return value, nil
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (p *PegasusStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	table, err := p.client.OpenTable(ctx, p.options.TableName)
	if err != nil {
		return nil, 0, err
	}
	defer table.Close()

	value, err := table.Get(ctx, []byte(cast.ToString(key)), empty)
	if err != nil {
		return nil, 0, err
	}
	if value == nil {
		return nil, 0, &lib_store.NotFound{}
	}

	ttl, err := table.TTL(ctx, []byte(cast.ToString(key)), empty)
	if err != nil {
		return nil, 0, err
	}

	return value, time.Duration(ttl) * time.Second, nil
}

// Set defines data in Pegasus for given key identifier
func (p *PegasusStore) Set(ctx context.Context, key, value any, options ...lib_store.Option) error {
	opts := lib_store.ApplyOptions(options...)

	table, err := p.client.OpenTable(ctx, p.options.TableName)
	if err != nil {
		return err
	}
	defer table.Close()

	err = table.SetTTL(ctx, []byte(cast.ToString(key)), empty, []byte(cast.ToString(value)), opts.Expiration)
	if err != nil {
		return err
	}

	if tags := opts.Tags; len(tags) > 0 {
		if err = p.setTags(ctx, key, tags); err != nil {
			return err
		}
	}
	return nil
}

func (p *PegasusStore) setTags(ctx context.Context, key any, tags []string) error {
	for _, tag := range tags {
		tagKey := fmt.Sprintf(PegasusTagPattern, tag)
		cacheKeys := []string{}

		if result, err := p.Get(ctx, tagKey); err == nil {
			if bytes, ok := result.([]byte); ok {
				cacheKeys = strings.Split(string(bytes), ",")
			}
		}

		alreadyInserted := false
		for _, cacheKey := range cacheKeys {
			if cacheKey == key.(string) {
				alreadyInserted = true
				break
			}
		}

		if !alreadyInserted {
			cacheKeys = append(cacheKeys, key.(string))
		}

		if err := p.Set(ctx, tagKey, []byte(strings.Join(cacheKeys, ",")), lib_store.WithExpiration(720*time.Hour)); err != nil {
			return err
		}
	}

	return nil
}

// Delete removes data from Pegasus for given key identifier
func (p *PegasusStore) Delete(ctx context.Context, key any) error {
	table, err := p.client.OpenTable(ctx, p.options.TableName)
	if err != nil {
		return err
	}
	defer table.Close()

	return table.Del(ctx, []byte(cast.ToString(key)), empty)
}

// Invalidate invalidates some cache data in Pegasus for given options
func (p *PegasusStore) Invalidate(ctx context.Context, options ...lib_store.InvalidateOption) error {
	opts := lib_store.ApplyInvalidateOptions(options...)
	if tags := opts.Tags; len(tags) > 0 {
		for _, tag := range tags {
			tagKey := fmt.Sprintf(PegasusTagPattern, tag)
			result, err := p.Get(ctx, tagKey)
			if err != nil {
				return nil
			}

			cacheKeys := []string{}
			if bytes, ok := result.([]byte); ok {
				cacheKeys = strings.Split(string(bytes), ",")
			}

			for _, cacheKey := range cacheKeys {
				if err := p.Delete(ctx, cacheKey); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Clear resets all data in the store
func (p *PegasusStore) Clear(ctx context.Context) error {
	table, err := p.client.OpenTable(ctx, p.options.TableName)
	if err != nil {
		return err
	}
	defer table.Close()

	// init full scan
	scanners, err := table.GetUnorderedScanners(ctx, p.options.TablePartitionNum, &pegasus.ScannerOptions{
		BatchSize: p.options.TableScanNum,
		// Values can be optimized out during scanning to reduce the workload.
		NoValue: true,
	})
	if err != nil {
		return err
	}

	// full scan and delete
	for _, scanner := range scanners {
		// Iterates sequentially.
		for {
			completed, hashKey, _, _, err := scanner.Next(ctx)
			if err != nil {
				return err
			}
			if completed {
				break
			}
			err = p.Delete(ctx, hashKey)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetType returns the store type
func (p *PegasusStore) GetType() string {
	return PegasusType
}
