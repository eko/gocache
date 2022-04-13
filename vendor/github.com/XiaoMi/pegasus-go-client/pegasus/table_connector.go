// Copyright (c) 2017, Xiaomi, Inc.  All rights reserved.
// This source code is licensed under the Apache License Version 2.0, which
// can be found in the LICENSE file in the root directory of this source tree.

package pegasus

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/replication"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/XiaoMi/pegasus-go-client/pegalog"
	"github.com/XiaoMi/pegasus-go-client/pegasus/op"
	"github.com/XiaoMi/pegasus-go-client/session"
	"gopkg.in/tomb.v2"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
)

// KeyValue is the returned type of MultiGet and MultiGetRange.
type KeyValue struct {
	SortKey, Value []byte
}

// CompositeKey is a composition of HashKey and SortKey.
type CompositeKey struct {
	HashKey, SortKey []byte
}

// MultiGetOptions is the options for MultiGet and MultiGetRange, defaults to DefaultMultiGetOptions.
type MultiGetOptions struct {
	StartInclusive bool
	StopInclusive  bool
	SortKeyFilter  Filter

	// MaxFetchCount and MaxFetchSize limit the size of returned result.

	// Max count of k-v pairs to be fetched. MaxFetchCount <= 0 means no limit.
	MaxFetchCount int

	// Max size of k-v pairs to be fetched. MaxFetchSize <= 0 means no limit.
	MaxFetchSize int

	// Query order
	Reverse bool

	// Whether to retrieve keys only, without value.
	// Enabling this option will reduce the network load, improve the RPC latency.
	NoValue bool
}

// DefaultMultiGetOptions defines the defaults of MultiGetOptions.
var DefaultMultiGetOptions = &MultiGetOptions{
	StartInclusive: true,
	StopInclusive:  false,
	SortKeyFilter: Filter{
		Type:    FilterTypeNoFilter,
		Pattern: nil,
	},
	MaxFetchCount: 100,
	MaxFetchSize:  100000,
	NoValue:       false,
}

// TableConnector is used to communicate with single Pegasus table.
type TableConnector interface {
	// Get retrieves the entry for `hashKey` + `sortKey`.
	// Returns nil if no entry matches.
	// `hashKey` : CAN'T be nil or empty.
	// `sortKey` : CAN'T be nil but CAN be empty.
	Get(ctx context.Context, hashKey []byte, sortKey []byte) ([]byte, error)

	// Set the entry for `hashKey` + `sortKey` to `value`.
	// If Set is called or `ttl` == 0, no data expiration is specified.
	// `hashKey` : CAN'T be nil or empty.
	// `sortKey` / `value` : CAN'T be nil but CAN be empty.
	Set(ctx context.Context, hashKey []byte, sortKey []byte, value []byte) error
	SetTTL(ctx context.Context, hashKey []byte, sortKey []byte, value []byte, ttl time.Duration) error

	// Delete the entry for `hashKey` + `sortKey`.
	// `hashKey` : CAN'T be nil or empty.
	// `sortKey` : CAN'T be nil but CAN be empty.
	Del(ctx context.Context, hashKey []byte, sortKey []byte) error

	// MultiGet/MultiGetOpt retrieves the multiple entries for `hashKey` + `sortKeys[i]` atomically in one operation.
	// MultiGet is identical to MultiGetOpt except that the former uses DefaultMultiGetOptions as `options`.
	//
	// If `sortKeys` are given empty or nil, all entries under `hashKey` will be retrieved.
	// `hashKey` : CAN'T be nil or empty.
	// `sortKeys[i]` : CAN'T be nil but CAN be empty.
	//
	// The returned key-value pairs are sorted by sort key in ascending order.
	// Returns nil if no entries match.
	// Returns true if all data is fetched, false if only partial data is fetched.
	//
	MultiGet(ctx context.Context, hashKey []byte, sortKeys [][]byte) ([]*KeyValue, bool, error)
	MultiGetOpt(ctx context.Context, hashKey []byte, sortKeys [][]byte, options *MultiGetOptions) ([]*KeyValue, bool, error)

	// MultiGetRange retrieves the multiple entries under `hashKey`, between range (`startSortKey`, `stopSortKey`),
	// atomically in one operation.
	//
	// startSortKey: nil or len(startSortKey) == 0 means start from begin.
	// stopSortKey: nil or len(stopSortKey) == 0 means stop to end.
	// `hashKey` : CAN'T be nil.
	//
	// The returned key-value pairs are sorted by sort keys in ascending order.
	// Returns nil if no entries match.
	// Returns true if all data is fetched, false if only partial data is fetched.
	//
	MultiGetRange(ctx context.Context, hashKey []byte, startSortKey []byte, stopSortKey []byte) ([]*KeyValue, bool, error)
	MultiGetRangeOpt(ctx context.Context, hashKey []byte, startSortKey []byte, stopSortKey []byte, options *MultiGetOptions) ([]*KeyValue, bool, error)

	// MultiSet sets the multiple entries for `hashKey` + `sortKeys[i]` atomically in one operation.
	// `hashKey` / `sortKeys` / `values` : CAN'T be nil or empty.
	// `sortKeys[i]` / `values[i]` : CAN'T be nil but CAN be empty.
	MultiSet(ctx context.Context, hashKey []byte, sortKeys [][]byte, values [][]byte) error
	MultiSetOpt(ctx context.Context, hashKey []byte, sortKeys [][]byte, values [][]byte, ttl time.Duration) error

	// MultiDel deletes the multiple entries under `hashKey` all atomically in one operation.
	// `hashKey` / `sortKeys` : CAN'T be nil or empty.
	// `sortKeys[i]` : CAN'T be nil but CAN be empty.
	MultiDel(ctx context.Context, hashKey []byte, sortKeys [][]byte) error

	// Returns ttl(time-to-live) in seconds: -1 if ttl is not set; -2 if entry doesn't exist.
	// `hashKey` : CAN'T be nil or empty.
	// `sortKey` : CAN'T be nil but CAN be empty.
	TTL(ctx context.Context, hashKey []byte, sortKey []byte) (int, error)

	// Check value existence for the entry for `hashKey` + `sortKey`.
	// `hashKey`: CAN'T be nil or empty.
	Exist(ctx context.Context, hashKey []byte, sortKey []byte) (bool, error)

	// Get Scanner for {startSortKey, stopSortKey} within hashKey.
	// startSortKey: nil or len(startSortKey) == 0 means start from begin.
	// stopSortKey: nil or len(stopSortKey) == 0 means stop to end.
	// `hashKey`: CAN'T be nil or empty.
	GetScanner(ctx context.Context, hashKey []byte, startSortKey []byte, stopSortKey []byte, options *ScannerOptions) (Scanner, error)

	// Get Scanners for all data in pegasus, the count of scanners will
	// be no more than maxSplitCount
	GetUnorderedScanners(ctx context.Context, maxSplitCount int, options *ScannerOptions) ([]Scanner, error)

	// Atomically check and set value by key from the cluster. The value will be set if and only if check passed.
	// The sort key for checking and setting can be the same or different.
	//
	// `checkSortKey`: The sort key for checking.
	// `setSortKey`: The sort key for setting.
	// `checkOperand`:
	CheckAndSet(ctx context.Context, hashKey []byte, checkSortKey []byte, checkType CheckType,
		checkOperand []byte, setSortKey []byte, setValue []byte, options *CheckAndSetOptions) (*CheckAndSetResult, error)

	// Returns the count of sortkeys under hashkey.
	// `hashKey`: CAN'T be nil or empty.
	SortKeyCount(ctx context.Context, hashKey []byte) (int64, error)

	// Atomically increment value by key from the cluster.
	// Returns the new value.
	// `hashKey` / `sortKeys` : CAN'T be nil or empty
	Incr(ctx context.Context, hashKey []byte, sortKey []byte, increment int64) (int64, error)

	// Gets values from a batch of CompositeKeys. Internally it distributes each key
	// into a Get call and wait until all returned.
	//
	// `keys`: CAN'T be nil or empty, `hashkey` in `keys` can't be nil or empty either.
	// The returned values are in sequence order of each key, aka `keys[i] => values[i]`.
	// If keys[i] is not found, or the Get failed, values[i] is set nil.
	//
	// Returns a non-nil `err` once there's a failed Get call. It doesn't mean all calls failed.
	//
	// NOTE: this operation is not guaranteed to be atomic
	BatchGet(ctx context.Context, keys []CompositeKey) (values [][]byte, err error)

	Close() error
}

type pegasusTableConnector struct {
	meta    *session.MetaManager
	replica *session.ReplicaManager

	logger pegalog.Logger

	tableName string
	appID     int32
	parts     []*replicaNode
	mu        sync.RWMutex

	confUpdateCh chan bool
	tom          tomb.Tomb
}

type replicaNode struct {
	session *session.ReplicaSession
	pconf   *replication.PartitionConfiguration
}

// ConnectTable queries for the configuration of the given table, and set up connection to
// the replicas which the table locates on.
func ConnectTable(ctx context.Context, tableName string, meta *session.MetaManager, replica *session.ReplicaManager) (TableConnector, error) {
	p := &pegasusTableConnector{
		tableName:    tableName,
		meta:         meta,
		replica:      replica,
		confUpdateCh: make(chan bool, 1),
		logger:       pegalog.GetLogger(),
	}

	// if the session became unresponsive, TableConnector auto-triggers
	// a update of the routing table.
	p.replica.SetUnresponsiveHandler(func(n session.NodeSession) {
		p.tryConfUpdate(errors.New("session unresponsive for long"), n)
	})

	if err := p.updateConf(ctx); err != nil {
		return nil, err
	}

	p.tom.Go(p.loopForAutoUpdate)
	return p, nil
}

// Update configuration of this table.
func (p *pegasusTableConnector) updateConf(ctx context.Context) error {
	resp, err := p.meta.QueryConfig(ctx, p.tableName)
	if err == nil {
		err = p.handleQueryConfigResp(resp)
	}
	if err != nil {
		return fmt.Errorf("failed to connect table(%s): %s", p.tableName, err)
	}
	return nil
}

func (p *pegasusTableConnector) handleQueryConfigResp(resp *replication.QueryCfgResponse) error {
	if resp.Err.Errno != base.ERR_OK.String() {
		return errors.New(resp.Err.Errno)
	}
	if resp.PartitionCount == 0 || len(resp.Partitions) != int(resp.PartitionCount) {
		return fmt.Errorf("invalid table configuration: response [%v]", resp)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.appID = resp.AppID

	if len(resp.Partitions) > len(p.parts) {
		// during partition split or first configuration update of client.
		for _, part := range p.parts {
			part.session.Close()
		}
		p.parts = make([]*replicaNode, len(resp.Partitions))
	}

	// TODO(wutao1): make sure PartitionIndex are continuous
	for _, pconf := range resp.Partitions {
		if pconf == nil || pconf.Primary == nil || pconf.Primary.GetRawAddress() == 0 {
			return fmt.Errorf("unable to resolve routing table [appid: %d]: [%v]", p.appID, pconf)
		}
		r := &replicaNode{
			pconf:   pconf,
			session: p.replica.GetReplica(pconf.Primary.GetAddress()),
		}
		p.parts[pconf.Pid.PartitionIndex] = r
	}
	return nil
}

func validateHashKey(hashKey []byte) error {
	if hashKey == nil {
		return fmt.Errorf("InvalidParameter: hashkey must not be nil")
	}
	if len(hashKey) == 0 {
		return fmt.Errorf("InvalidParameter: hashkey must not be empty")
	}
	if len(hashKey) > math.MaxUint16 {
		return fmt.Errorf("InvalidParameter: length of hashkey (%d) must be less than %d", len(hashKey), math.MaxUint16)
	}
	return nil
}

func validateCompositeKeys(keys []CompositeKey) error {
	if keys == nil {
		return fmt.Errorf("InvalidParameter: CompositeKeys must not be nil")
	}
	if len(keys) == 0 {
		return fmt.Errorf("InvalidParameter: CompositeKeys must not be empty")
	}
	return nil
}

// WrapError wraps up the internal errors for ensuring that all types of errors
// returned by public interfaces are pegasus.PError.
func WrapError(err error, op OpType) error {
	if err != nil {
		if pe, ok := err.(*PError); ok {
			pe.Op = op
			return pe
		}
		return &PError{
			Err: err,
			Op:  op,
		}
	}
	return nil
}

func (p *pegasusTableConnector) wrapPartitionError(err error, gpid *base.Gpid, replica *session.ReplicaSession, opType OpType) error {
	err = WrapError(err, opType)
	if err == nil {
		return nil
	}
	perr := err.(*PError)
	if perr.Err != nil {
		perr.Err = fmt.Errorf("%s [%s, %s, table=%s]", perr.Err, gpid, replica, p.tableName)
	} else {
		perr.Err = fmt.Errorf("[%s, %s, table=%s]", gpid, replica, p.tableName)
	}
	return perr
}

func (p *pegasusTableConnector) Get(ctx context.Context, hashKey []byte, sortKey []byte) ([]byte, error) {
	res, err := p.runPartitionOp(ctx, hashKey, &op.Get{HashKey: hashKey, SortKey: sortKey}, OpGet)
	if err != nil {
		return nil, err
	}
	if res == nil { // indicates the record is not found
		return nil, nil
	}
	return res.([]byte), err
}

func (p *pegasusTableConnector) SetTTL(ctx context.Context, hashKey []byte, sortKey []byte, value []byte, ttl time.Duration) error {
	req := &op.Set{HashKey: hashKey, SortKey: sortKey, Value: value, TTL: ttl}
	_, err := p.runPartitionOp(ctx, hashKey, req, OpSet)
	return err
}

func (p *pegasusTableConnector) Set(ctx context.Context, hashKey []byte, sortKey []byte, value []byte) error {
	return p.SetTTL(ctx, hashKey, sortKey, value, 0)
}

func (p *pegasusTableConnector) Del(ctx context.Context, hashKey []byte, sortKey []byte) error {
	req := &op.Del{HashKey: hashKey, SortKey: sortKey}
	_, err := p.runPartitionOp(ctx, hashKey, req, OpDel)
	return err
}

func setRequestByOption(options *MultiGetOptions, request *rrdb.MultiGetRequest) {
	request.MaxKvCount = int32(options.MaxFetchCount)
	request.MaxKvSize = int32(options.MaxFetchSize)
	request.StartInclusive = options.StartInclusive
	request.StopInclusive = options.StopInclusive
	request.SortKeyFilterType = rrdb.FilterType(options.SortKeyFilter.Type)
	request.SortKeyFilterPattern = &base.Blob{Data: options.SortKeyFilter.Pattern}
	request.Reverse = options.Reverse
	request.NoValue = options.NoValue
}

func (p *pegasusTableConnector) MultiGetOpt(ctx context.Context, hashKey []byte, sortKeys [][]byte, options *MultiGetOptions) ([]*KeyValue, bool, error) {
	req := &op.MultiGet{HashKey: hashKey, SortKeys: sortKeys, Req: rrdb.NewMultiGetRequest()}
	setRequestByOption(options, req.Req)
	res, err := p.runPartitionOp(ctx, hashKey, req, OpMultiGet)
	if err != nil {
		return nil, false, err
	}
	return extractMultiGetResult(res.(*op.MultiGetResult))
}

func (p *pegasusTableConnector) MultiGet(ctx context.Context, hashKey []byte, sortKeys [][]byte) ([]*KeyValue, bool, error) {
	return p.MultiGetOpt(ctx, hashKey, sortKeys, DefaultMultiGetOptions)
}

func (p *pegasusTableConnector) MultiGetRangeOpt(ctx context.Context, hashKey []byte, startSortKey []byte, stopSortKey []byte, options *MultiGetOptions) ([]*KeyValue, bool, error) {
	req := &op.MultiGet{HashKey: hashKey, StartSortkey: startSortKey, StopSortkey: stopSortKey, Req: rrdb.NewMultiGetRequest()}
	setRequestByOption(options, req.Req)
	res, err := p.runPartitionOp(ctx, hashKey, req, OpMultiGetRange)
	if err != nil {
		return nil, false, err
	}
	return extractMultiGetResult(res.(*op.MultiGetResult))
}

func extractMultiGetResult(res *op.MultiGetResult) ([]*KeyValue, bool, error) {
	if len(res.KVs) == 0 {
		return nil, res.AllFetched, nil
	}
	kvs := make([]*KeyValue, len(res.KVs))
	for i, blobKv := range res.KVs {
		kvs[i] = &KeyValue{
			SortKey: blobKv.Key.Data,
			Value:   blobKv.Value.Data,
		}
	}
	return kvs, res.AllFetched, nil
}

func (p *pegasusTableConnector) MultiGetRange(ctx context.Context, hashKey []byte, startSortKey []byte, stopSortKey []byte) ([]*KeyValue, bool, error) {
	return p.MultiGetRangeOpt(ctx, hashKey, startSortKey, stopSortKey, DefaultMultiGetOptions)
}

func (p *pegasusTableConnector) MultiSet(ctx context.Context, hashKey []byte, sortKeys [][]byte, values [][]byte) error {
	return p.MultiSetOpt(ctx, hashKey, sortKeys, values, 0)
}

func (p *pegasusTableConnector) MultiSetOpt(ctx context.Context, hashKey []byte, sortKeys [][]byte, values [][]byte, ttl time.Duration) error {
	req := &op.MultiSet{HashKey: hashKey, SortKeys: sortKeys, Values: values, TTL: ttl}
	_, err := p.runPartitionOp(ctx, hashKey, req, OpMultiSet)
	return err
}

func (p *pegasusTableConnector) MultiDel(ctx context.Context, hashKey []byte, sortKeys [][]byte) error {
	_, err := p.runPartitionOp(ctx, hashKey, &op.MultiDel{HashKey: hashKey, SortKeys: sortKeys}, OpMultiDel)
	return err
}

// -2 means entry not found.
func (p *pegasusTableConnector) TTL(ctx context.Context, hashKey []byte, sortKey []byte) (int, error) {
	res, err := p.runPartitionOp(ctx, hashKey, &op.TTL{HashKey: hashKey, SortKey: sortKey}, OpTTL)
	return res.(int), err
}

func (p *pegasusTableConnector) Exist(ctx context.Context, hashKey []byte, sortKey []byte) (bool, error) {
	ttl, err := p.TTL(ctx, hashKey, sortKey)
	if err == nil {
		if ttl == -2 {
			return false, nil
		}
		return true, nil
	}
	return false, WrapError(err, OpExist)
}

func (p *pegasusTableConnector) GetScanner(ctx context.Context, hashKey []byte, startSortKey []byte, stopSortKey []byte,
	options *ScannerOptions) (Scanner, error) {
	scanner, err := func() (Scanner, error) {
		if err := validateHashKey(hashKey); err != nil {
			return nil, err
		}

		start := encodeHashKeySortKey(hashKey, startSortKey)
		var stop *base.Blob
		if len(stopSortKey) == 0 {
			stop = encodeHashKeySortKey(hashKey, []byte{0xFF, 0xFF}) // []byte{0xFF, 0xFF} means the max sortKey value
			options.StopInclusive = false
		} else {
			stop = encodeHashKeySortKey(hashKey, stopSortKey)
		}

		if options.SortKeyFilter.Type == FilterTypeMatchPrefix {
			prefixStartBlob := encodeHashKeySortKey(hashKey, options.SortKeyFilter.Pattern)

			// if the prefixStartKey generated by pattern is greater than the startKey, start from the prefixStartKey
			if bytes.Compare(prefixStartBlob.Data, start.Data) > 0 {
				start = prefixStartBlob
				options.StartInclusive = true
			}

			prefixStop := encodeNextBytesByKeys(hashKey, options.SortKeyFilter.Pattern)

			// if the prefixStopKey generated by pattern is less than the stopKey, end to the prefixStopKey
			if bytes.Compare(prefixStop.Data, stop.Data) <= 0 {
				stop = prefixStop
				options.StopInclusive = false
			}
		}

		cmp := bytes.Compare(start.Data, stop.Data)
		if cmp < 0 || (cmp == 0 && options.StartInclusive && options.StopInclusive) {
			gpid, err := p.getGpid(start.Data)
			if err != nil && gpid != nil {
				return nil, err
			}
			return newPegasusScanner(p, gpid, options, start, stop), nil
		}
		return nil, fmt.Errorf("the scanning interval MUST NOT BE EMPTY")
	}()
	return scanner, WrapError(err, OpGetScanner)
}

func (p *pegasusTableConnector) GetUnorderedScanners(ctx context.Context, maxSplitCount int,
	options *ScannerOptions) ([]Scanner, error) {
	scanners, err := func() ([]Scanner, error) {
		if maxSplitCount <= 0 {
			return nil, fmt.Errorf("invalid maxSplitCount: %d", maxSplitCount)
		}
		allGpid := p.getAllGpid()
		total := len(allGpid)

		var split int // the actual split count
		if total < maxSplitCount {
			split = total
		} else {
			split = maxSplitCount
		}
		scanners := make([]Scanner, split)

		// k: the smallest multiple of split which is greater than or equal to total
		k := 1
		for ; k*split < total; k++ {
		}
		left := total - k*(split-1)

		sliceLen := 0
		id := 0
		for i := 0; i < split; i++ {
			if i == 0 {
				sliceLen = left
			} else {
				sliceLen = k
			}
			gpidSlice := make([]*base.Gpid, sliceLen)
			for j := 0; j < sliceLen; j++ {
				gpidSlice[j] = allGpid[id]
				id++
			}
			scanners[i] = newPegasusScannerForUnorderedScanners(p, gpidSlice, options)
		}
		return scanners, nil
	}()
	return scanners, WrapError(err, OpGetUnorderedScanners)
}

func (p *pegasusTableConnector) CheckAndSet(ctx context.Context, hashKey []byte, checkSortKey []byte, checkType CheckType,
	checkOperand []byte, setSortKey []byte, setValue []byte, options *CheckAndSetOptions) (*CheckAndSetResult, error) {

	if options == nil {
		options = &CheckAndSetOptions{}
	}
	request := rrdb.NewCheckAndSetRequest()
	request.CheckType = rrdb.CasCheckType(checkType)
	request.CheckOperand = &base.Blob{Data: checkOperand}
	request.CheckSortKey = &base.Blob{Data: checkSortKey}
	request.HashKey = &base.Blob{Data: hashKey}
	request.SetExpireTsSeconds = 0
	if options.SetValueTTLSeconds != 0 {
		request.SetExpireTsSeconds = expireTsSeconds(time.Second * time.Duration(options.SetValueTTLSeconds))
	}
	request.SetSortKey = &base.Blob{Data: setSortKey}
	request.SetValue = &base.Blob{Data: setValue}
	request.ReturnCheckValue = options.ReturnCheckValue
	if !bytes.Equal(checkSortKey, setSortKey) {
		request.SetDiffSortKey = true
	} else {
		request.SetDiffSortKey = false
	}

	req := &op.CheckAndSet{Req: request}
	res, err := p.runPartitionOp(ctx, hashKey, req, OpCheckAndSet)
	if err != nil {
		return nil, err
	}
	casRes := res.(*op.CheckAndSetResult)
	return &CheckAndSetResult{
		SetSucceed:         casRes.SetSucceed,
		CheckValue:         casRes.CheckValue,
		CheckValueExist:    casRes.CheckValueExist,
		CheckValueReturned: casRes.CheckValueReturned,
	}, nil
}

func (p *pegasusTableConnector) SortKeyCount(ctx context.Context, hashKey []byte) (int64, error) {
	res, err := p.runPartitionOp(ctx, hashKey, &op.SortKeyCount{HashKey: hashKey}, OpSortKeyCount)
	if err != nil {
		return 0, err
	}
	return res.(int64), nil
}

func (p *pegasusTableConnector) Incr(ctx context.Context, hashKey []byte, sortKey []byte, increment int64) (int64, error) {
	req := &op.Incr{HashKey: hashKey, SortKey: sortKey, Increment: increment}
	res, err := p.runPartitionOp(ctx, hashKey, req, OpIncr)
	if err != nil {
		return 0, err
	}
	return res.(int64), nil
}

func (p *pegasusTableConnector) runPartitionOp(ctx context.Context, hashKey []byte, req op.Request, optype OpType) (interface{}, error) {
	// validate arguments
	if err := req.Validate(); err != nil {
		return 0, WrapError(err, optype)
	}
	gpid, part := p.getPartition(hashKey)
	res, err := retryFailOver(ctx, func() (confUpdated bool, result interface{}, err error) {
		result, err = req.Run(ctx, gpid, part)
		confUpdated, err = p.handleReplicaError(err, part)
		return
	})
	return res, p.wrapPartitionError(err, gpid, part, optype)
}

func (p *pegasusTableConnector) BatchGet(ctx context.Context, keys []CompositeKey) (values [][]byte, err error) {
	v, err := func() ([][]byte, error) {
		if err := validateCompositeKeys(keys); err != nil {
			return nil, err
		}

		values = make([][]byte, len(keys))
		funcs := make([]func() error, 0, len(keys))
		for i := 0; i < len(keys); i++ {
			idx := i
			funcs = append(funcs, func() (subErr error) {
				key := keys[idx]
				values[idx], subErr = p.Get(ctx, key.HashKey, key.SortKey)
				if subErr != nil {
					values[idx] = nil
					return subErr
				}
				return nil
			})
		}
		return values, kerrors.AggregateGoroutines(funcs...)
	}()
	return v, WrapError(err, OpBatchGet)
}

func getPartitionIndex(hashKey []byte, partitionCount int) int32 {
	return int32(crc64Hash(hashKey) % uint64(partitionCount))
}

func (p *pegasusTableConnector) getPartition(hashKey []byte) (*base.Gpid, *session.ReplicaSession) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	gpid := &base.Gpid{
		Appid:          p.appID,
		PartitionIndex: getPartitionIndex(hashKey, len(p.parts)),
	}
	part := p.parts[gpid.PartitionIndex].session

	return gpid, part
}

func (p *pegasusTableConnector) getPartitionByGpid(gpid *base.Gpid) *session.ReplicaSession {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.parts[gpid.PartitionIndex].session
}

func (p *pegasusTableConnector) Close() error {
	p.tom.Kill(errors.New("table closed"))
	return nil
}

func (p *pegasusTableConnector) handleReplicaError(err error, replica *session.ReplicaSession) (bool, error) {
	if err != nil {
		confUpdate := false

		switch err {
		case base.ERR_OK:
			// should not happen
			return false, nil

		case base.ERR_TIMEOUT:
		case context.DeadlineExceeded:
		case context.Canceled:
			// timeout will not trigger a configuration update

		case base.ERR_NOT_ENOUGH_MEMBER:
		case base.ERR_CAPACITY_EXCEEDED:

		case base.ERR_BUSY:
			// throttled by server, skip confUpdate

		default:
			confUpdate = true
		}

		switch err {
		case base.ERR_BUSY:
			err = errors.New(err.Error() + " Rate of requests exceeds the throughput limit")
		case base.ERR_INVALID_STATE:
			err = errors.New(err.Error() + " The target replica is not primary")
		case base.ERR_OBJECT_NOT_FOUND:
			err = errors.New(err.Error() + " The replica server doesn't serve this partition")
		}

		if confUpdate {
			// we need to check if there's newer configuration.
			p.tryConfUpdate(err, replica)
		}

		return confUpdate, err
	}
	return false, nil
}

// tryConfUpdate makes an attempt to update table configuration by querying meta server.
func (p *pegasusTableConnector) tryConfUpdate(err error, replica session.NodeSession) {
	select {
	case p.confUpdateCh <- true:
		p.logger.Printf("trigger configuration update of table [%s] due to RPC failure [%s] to %s", p.tableName, err, replica)
	default:
	}
}

func (p *pegasusTableConnector) loopForAutoUpdate() error {
	for {
		select {
		case <-p.confUpdateCh:
			p.selfUpdate()
		case <-p.tom.Dying():
			return nil
		}

		// sleep a while
		select {
		case <-time.After(time.Second):
		case <-p.tom.Dying():
			return nil
		}
	}
}

func (p *pegasusTableConnector) selfUpdate() bool {
	// ignore the returned error
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := p.updateConf(ctx); err != nil {
		p.logger.Printf("self update failed [table: %s]: %s", p.tableName, err.Error())
	}

	// flush confUpdateCh
	select {
	case <-p.confUpdateCh:
	default:
	}

	return true
}

func (p *pegasusTableConnector) getGpid(key []byte) (*base.Gpid, error) {
	if key == nil || len(key) < 2 {
		return nil, fmt.Errorf("unable to getGpid by key: %s", key)
	}

	hashKeyLen := 0xFFFF & binary.BigEndian.Uint16(key[:2])
	if hashKeyLen != 0xFFFF && int(2+hashKeyLen) <= len(key) {
		gpid := &base.Gpid{Appid: p.appID}
		if hashKeyLen == 0 {
			gpid.PartitionIndex = int32(crc64Hash(key[2:]) % uint64(len(p.parts)))
		} else {
			gpid.PartitionIndex = int32(crc64Hash(key[2:hashKeyLen+2]) % uint64(len(p.parts)))
		}
		return gpid, nil

	}
	return nil, fmt.Errorf("unable to getGpid, hashKey length invalid")
}

func (p *pegasusTableConnector) getAllGpid() []*base.Gpid {
	p.mu.RLock()
	defer p.mu.RUnlock()
	count := len(p.parts)
	ret := make([]*base.Gpid, count)
	for i := 0; i < count; i++ {
		ret[i] = &base.Gpid{Appid: p.appID, PartitionIndex: int32(i)}
	}
	return ret
}
