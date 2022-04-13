// Copyright (c) 2017, Xiaomi, Inc.  All rights reserved.
// This source code is licensed under the Apache License Version 2.0, which
// can be found in the LICENSE file in the root directory of this source tree.

package session

import (
	"context"
	"sync"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/replication"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/XiaoMi/pegasus-go-client/pegalog"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
)

// metaSession represents the network session between client and meta server.
type metaSession struct {
	NodeSession

	logger pegalog.Logger
}

func (ms *metaSession) call(ctx context.Context, args RpcRequestArgs, rpcName string) (RpcResponseResult, error) {
	return ms.CallWithGpid(ctx, &base.Gpid{Appid: 0, PartitionIndex: 0}, args, rpcName)
}

func (ms *metaSession) queryConfig(ctx context.Context, tableName string) (*replication.QueryCfgResponse, error) {
	ms.logger.Printf("querying configuration of table(%s) from %s", tableName, ms)

	arg := rrdb.NewMetaQueryCfgArgs()
	arg.Query = replication.NewQueryCfgRequest()
	arg.Query.AppName = tableName
	arg.Query.PartitionIndices = []int32{}

	result, err := ms.call(ctx, arg, "RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX")
	if err != nil {
		ms.logger.Printf("failed to query configuration from %s: %s", ms, err)
		return nil, err
	}

	ret, _ := result.(*rrdb.MetaQueryCfgResult)
	return ret.GetSuccess(), nil
}

// MetaManager manages the list of metas, but only the leader will it request to.
// If the one is not the actual leader, it will retry with another.
type MetaManager struct {
	logger pegalog.Logger

	metaIPAddrs   []string
	metas         []*metaSession
	currentLeader int // current leader of meta servers

	// protect access of currentLeader
	mu sync.RWMutex
}

//
func NewMetaManager(addrs []string, creator NodeSessionCreator) *MetaManager {
	metas := make([]*metaSession, len(addrs))
	metaIPAddrs := make([]string, len(addrs))
	for i, addr := range addrs {
		metas[i] = &metaSession{
			NodeSession: creator(addr, NodeTypeMeta),
			logger:      pegalog.GetLogger(),
		}
		metaIPAddrs[i] = addr
	}

	mm := &MetaManager{
		currentLeader: 0,
		metas:         metas,
		metaIPAddrs:   metaIPAddrs,
		logger:        pegalog.GetLogger(),
	}
	return mm
}

func (m *MetaManager) call(ctx context.Context, callFunc metaCallFunc) (metaResponse, error) {
	lead := m.getCurrentLeader()
	call := newMetaCall(lead, m.metas, callFunc)
	resp, err := call.Run(ctx)
	if err == nil {
		m.setCurrentLeader(int(call.newLead))
	}
	return resp, err
}

// QueryConfig queries table configuration from the leader of meta servers. If the leader was changed,
// it retries for other servers until it finds the true leader, unless no leader exists.
// Thread-Safe
func (m *MetaManager) QueryConfig(ctx context.Context, tableName string) (*replication.QueryCfgResponse, error) {
	m.logger.Printf("querying configuration of table(%s) [metaList=%s]", tableName, m.metaIPAddrs)
	resp, err := m.call(ctx, func(rpcCtx context.Context, ms *metaSession) (metaResponse, error) {
		return ms.queryConfig(rpcCtx, tableName)
	})
	if err == nil {
		queryCfgResp := resp.(*replication.QueryCfgResponse)
		return queryCfgResp, nil
	}
	return nil, err
}

func (m *MetaManager) getCurrentLeader() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.currentLeader
}

func (m *MetaManager) setCurrentLeader(lead int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentLeader = lead
}

// Close the sessions.
func (m *MetaManager) Close() error {
	funcs := make([]func() error, len(m.metas))
	for i := 0; i < len(m.metas); i++ {
		idx := i
		funcs[idx] = func() error {
			return m.metas[idx].Close()
		}
	}
	return kerrors.AggregateGoroutines(funcs...)
}
