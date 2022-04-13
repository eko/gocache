// Copyright (c) 2017, Xiaomi, Inc.  All rights reserved.
// This source code is licensed under the Apache License Version 2.0, which
// can be found in the LICENSE file in the root directory of this source tree.

package session

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/XiaoMi/pegasus-go-client/idl/admin"
	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/cmd"
	"github.com/XiaoMi/pegasus-go-client/idl/radmin"
	"github.com/XiaoMi/pegasus-go-client/idl/replication"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/XiaoMi/pegasus-go-client/pegalog"
	"github.com/XiaoMi/pegasus-go-client/rpc"
	"github.com/pegasus-kv/thrift/lib/go/thrift"
)

type PegasusCodec struct {
	logger pegalog.Logger
}

func NewPegasusCodec() *PegasusCodec {
	return &PegasusCodec{logger: pegalog.GetLogger()}
}

func (p *PegasusCodec) Marshal(v interface{}) ([]byte, error) {
	r, _ := v.(*PegasusRpcCall)

	header := &thriftHeader{
		headerLength:   uint32(thriftHeaderBytesLen),
		appId:          r.Gpid.Appid,
		partitionIndex: r.Gpid.PartitionIndex,
		threadHash:     gpidToThreadHash(r.Gpid),
		partitionHash:  0,
	}

	// skip the first ThriftHeaderBytesLen bytes
	buf := thrift.NewTMemoryBuffer()
	buf.Write(make([]byte, thriftHeaderBytesLen))

	// encode body into buffer
	oprot := thrift.NewTBinaryProtocolTransport(buf)

	var err error
	if err = oprot.WriteMessageBegin(r.Name, thrift.CALL, r.SeqId); err != nil {
		return nil, err
	}
	if err = r.Args.Write(oprot); err != nil {
		return nil, err
	}
	if err = oprot.WriteMessageEnd(); err != nil {
		return nil, err
	}

	// encode header into buffer
	header.bodyLength = uint32(buf.Len() - thriftHeaderBytesLen)
	header.marshall(buf.Bytes()[0:thriftHeaderBytesLen])

	return buf.Bytes(), nil
}

func (p *PegasusCodec) Unmarshal(data []byte, v interface{}) error {
	r, _ := v.(*PegasusRpcCall)

	iprot := thrift.NewTBinaryProtocolTransport(thrift.NewStreamTransportR(bytes.NewBuffer(data)))
	ec := &base.ErrorCode{}
	if err := ec.Read(iprot); err != nil {
		return err
	}

	name, _, seqId, err := iprot.ReadMessageBegin()
	if err != nil {
		return err
	}

	r.Name = name
	r.SeqId = seqId

	if ec.Errno != base.ERR_OK.String() {
		// convert string to base.DsnErrCode
		err, parseErr := base.DsnErrCodeString(ec.Errno)
		if parseErr != nil {
			p.logger.Print("failed to unmarshal the heading error code of rpc response: ", parseErr)
			return parseErr
		}

		r.Err = err
		return nil
	}

	nameToResultFunc, ok := nameToResultMap[name]
	if !ok {
		return fmt.Errorf("failed to find rpc name: %s", name)
	}
	r.Result = nameToResultFunc()

	// read response body
	if err = r.Result.Read(iprot); err != nil {
		return err
	}
	if err = iprot.ReadMessageEnd(); err != nil {
		return err
	}

	return nil
}

func (p *PegasusCodec) String() string {
	return "pegasus"
}

// RegisterRPCResultHandler registers an external RPC that's not including in
// pegasus-go-client.
//
// The following example registers an response handler for Pegasus's remote-command RPC.
// Usage:
//
// ```go
//   RegisterRpcResultHandler("RPC_CLI_CLI_CALL_ACK", func() RpcResponseResult {
//     return &RemoteCmdServiceCallCommandResult{Success: new(string)}
//   })
// ```
func RegisterRPCResultHandler(responseAck string, handler func() RpcResponseResult) {
	nameToResultMapLock.Lock()
	defer nameToResultMapLock.Unlock()
	_, found := nameToResultMap[responseAck]
	if found {
		panic(fmt.Sprintf("register an registered RPC result handler: %s", responseAck))
	} else {
		nameToResultMap[responseAck] = handler
	}
}

var nameToResultMapLock sync.Mutex
var nameToResultMap = map[string]func() RpcResponseResult{
	"RPC_CM_QUERY_PARTITION_CONFIG_BY_INDEX_ACK": func() RpcResponseResult {
		return &rrdb.MetaQueryCfgResult{
			Success: replication.NewQueryCfgResponse(),
		}
	},
	"RPC_CM_CREATE_APP_ACK": func() RpcResponseResult {
		return &admin.AdminClientCreateAppResult{
			Success: admin.NewCreateAppResponse(),
		}
	},
	"RPC_CM_DROP_APP_ACK": func() RpcResponseResult {
		return &admin.AdminClientDropAppResult{
			Success: admin.NewDropAppResponse(),
		}
	},
	"RPC_CM_RECALL_APP_ACK": func() RpcResponseResult {
		return &admin.AdminClientRecallAppResult{
			Success: admin.NewRecallAppResponse(),
		}
	},
	"RPC_CM_LIST_APPS_ACK": func() RpcResponseResult {
		return &admin.AdminClientListAppsResult{
			Success: admin.NewListAppsResponse(),
		}
	},
	"RPC_QUERY_APP_INFO_ACK": func() RpcResponseResult {
		return &admin.AdminClientQueryAppInfoResult{
			Success: admin.NewQueryAppInfoResponse(),
		}
	},
	"RPC_CM_UPDATE_APP_ENV_ACK": func() RpcResponseResult {
		return &admin.AdminClientUpdateAppEnvResult{
			Success: admin.NewUpdateAppEnvResponse(),
		}
	},
	"RPC_CM_QUERY_DUPLICATION_ACK": func() RpcResponseResult {
		return &admin.AdminClientQueryDuplicationResult{
			Success: admin.NewDuplicationQueryResponse(),
		}
	},
	"RPC_CM_MODIFY_DUPLICATION_ACK": func() RpcResponseResult {
		return &admin.AdminClientModifyDuplicationResult{
			Success: admin.NewDuplicationModifyResponse(),
		}
	},
	"RPC_CM_ADD_DUPLICATION_ACK": func() RpcResponseResult {
		return &admin.AdminClientAddDuplicationResult{
			Success: admin.NewDuplicationAddResponse(),
		}
	},
	"RPC_CM_QUERY_BACKUP_POLICY_ACK": func() RpcResponseResult {
		return &admin.AdminClientQueryBackupPolicyResult{
			Success: admin.NewQueryBackupPolicyResponse(),
		}
	},
	"RPC_CM_CLUSTER_INFO_ACK": func() RpcResponseResult {
		return &admin.AdminClientQueryClusterInfoResult{
			Success: admin.NewClusterInfoResponse(),
		}
	},
	"RPC_CM_CONTROL_META_ACK": func() RpcResponseResult {
		return &admin.AdminClientMetaControlResult{
			Success: admin.NewMetaControlResponse(),
		}
	},
	"RPC_CM_LIST_NODES_ACK": func() RpcResponseResult {
		return &admin.AdminClientListNodesResult{
			Success: admin.NewListNodesResponse(),
		}
	},
	"RPC_CM_PROPOSE_BALANCER_ACK": func() RpcResponseResult {
		return &admin.AdminClientBalanceResult{
			Success: admin.NewBalanceResponse(),
		}
	},
	"RPC_QUERY_DISK_INFO_ACK": func() RpcResponseResult {
		return &radmin.ReplicaClientQueryDiskInfoResult{
			Success: radmin.NewQueryDiskInfoResponse(),
		}
	},
	"RPC_REPLICA_DISK_MIGRATE_ACK": func() RpcResponseResult {
		return &radmin.ReplicaClientDiskMigrateResult{
			Success: radmin.NewReplicaDiskMigrateResponse(),
		}
	},
	"RPC_RRDB_RRDB_GET_ACK": func() RpcResponseResult {
		return &rrdb.RrdbGetResult{
			Success: rrdb.NewReadResponse(),
		}
	},
	"RPC_RRDB_RRDB_PUT_ACK": func() RpcResponseResult {
		return &rrdb.RrdbPutResult{
			Success: rrdb.NewUpdateResponse(),
		}
	},
	"RPC_RRDB_RRDB_REMOVE_ACK": func() RpcResponseResult {
		return &rrdb.RrdbRemoveResult{
			Success: rrdb.NewUpdateResponse(),
		}
	},
	"RPC_RRDB_RRDB_MULTI_GET_ACK": func() RpcResponseResult {
		return &rrdb.RrdbMultiGetResult{
			Success: rrdb.NewMultiGetResponse(),
		}
	},
	"RPC_RRDB_RRDB_MULTI_REMOVE_ACK": func() RpcResponseResult {
		return &rrdb.RrdbMultiRemoveResult{
			Success: rrdb.NewMultiRemoveResponse(),
		}
	},
	"RPC_RRDB_RRDB_MULTI_PUT_ACK": func() RpcResponseResult {
		return &rrdb.RrdbMultiPutResult{
			Success: rrdb.NewUpdateResponse(),
		}
	},
	"RPC_RRDB_RRDB_TTL_ACK": func() RpcResponseResult {
		return &rrdb.RrdbTTLResult{
			Success: rrdb.NewTTLResponse(),
		}
	},
	"RPC_RRDB_RRDB_GET_SCANNER_ACK": func() RpcResponseResult {
		return &rrdb.RrdbGetScannerResult{
			Success: rrdb.NewScanResponse(),
		}
	},
	"RPC_RRDB_RRDB_SCAN_ACK": func() RpcResponseResult {
		return &rrdb.RrdbScanResult{
			Success: rrdb.NewScanResponse(),
		}
	},
	"RPC_RRDB_RRDB_CHECK_AND_SET_ACK": func() RpcResponseResult {
		return &rrdb.RrdbCheckAndSetResult{
			Success: rrdb.NewCheckAndSetResponse(),
		}
	},
	"RPC_RRDB_RRDB_SORTKEY_COUNT_ACK": func() RpcResponseResult {
		return &rrdb.RrdbSortkeyCountResult{
			Success: rrdb.NewCountResponse(),
		}
	},
	"RPC_RRDB_RRDB_INCR_ACK": func() RpcResponseResult {
		return &rrdb.RrdbIncrResult{
			Success: rrdb.NewIncrResponse(),
		}
	},
	"RPC_CLI_CLI_CALL_ACK": func() RpcResponseResult {
		return &cmd.RemoteCmdServiceCallCommandResult{
			Success: new(string),
		}
	},
}

// MockCodec is only used for testing.
// By default it does nothing on marshalling and unmarshalling,
// thus it returns no error even if the input was ill-formed.
type MockCodec struct {
	mars   MarshalFunc
	unmars UnmarshalFunc
}

type UnmarshalFunc func(data []byte, v interface{}) error

type MarshalFunc func(v interface{}) ([]byte, error)

func (p *MockCodec) Marshal(v interface{}) ([]byte, error) {
	if p.mars != nil {
		return p.mars(v)
	}
	return nil, nil
}

func (p *MockCodec) Unmarshal(data []byte, v interface{}) error {
	if p.unmars != nil {
		return p.unmars(data, v)
	}
	return nil
}

func (p *MockCodec) String() string {
	return "mock"
}

func (p *MockCodec) MockMarshal(marshal MarshalFunc) {
	p.mars = marshal
}

func (p *MockCodec) MockUnMarshal(unmarshal UnmarshalFunc) {
	p.unmars = unmarshal
}

// a trait of the thrift-generated argument type (MetaQueryCfgArgs, RrdbPutArgs e.g.)
type RpcRequestArgs interface {
	String() string
	Write(oprot thrift.TProtocol) error
}

// a trait of the thrift-generated result type (MetaQueryCfgResult e.g.)
type RpcResponseResult interface {
	String() string
	Read(iprot thrift.TProtocol) error
}

type PegasusRpcCall struct {
	Args   RpcRequestArgs
	Result RpcResponseResult
	Name   string // the rpc's name
	SeqId  int32
	Gpid   *base.Gpid
	RawReq []byte // the marshalled request in bytes
	Err    error

	// hooks on each stage during rpc processing
	OnRpcCall time.Time
	OnRpcSend time.Time
	OnRpcRecv time.Time
}

func (call *PegasusRpcCall) Trace() string {
	return fmt.Sprintf("call->%dus->send->%dus->recv->%dus->now",
		call.OnRpcSend.Sub(call.OnRpcCall)/time.Microsecond,
		call.OnRpcRecv.Sub(call.OnRpcSend)/time.Microsecond,
		time.Since(call.OnRpcRecv)/time.Microsecond)
}

func (call *PegasusRpcCall) TilNow() time.Duration {
	return time.Since(call.OnRpcCall)
}

func MarshallPegasusRpc(codec rpc.Codec, seqId int32, gpid *base.Gpid, args RpcRequestArgs, name string) (*PegasusRpcCall, error) {
	rcall := &PegasusRpcCall{}
	rcall.Args = args
	rcall.Name = name
	rcall.SeqId = seqId
	rcall.Gpid = gpid

	var err error
	rcall.RawReq, err = codec.Marshal(rcall)
	if err != nil {
		return nil, err
	}
	return rcall, nil
}

func ReadRpcResponse(conn *rpc.RpcConn, codec rpc.Codec) (*PegasusRpcCall, error) {
	// read length field
	lenBuf, err := conn.Read(4)
	if err != nil && len(lenBuf) < 4 {
		return nil, err
	}
	resplen := binary.BigEndian.Uint32(lenBuf)
	if resplen < 4 {
		return nil, fmt.Errorf("response length(%d) smaller than 4 bytes", resplen)
	}
	resplen -= 4 // 4 bytes for length

	// read data field
	buf, err := conn.Read(int(resplen))
	if err != nil || len(buf) != int(resplen) {
		return nil, err
	}

	r := &PegasusRpcCall{}
	if err := codec.Unmarshal(buf, r); err != nil {
		return nil, err
	}

	return r, nil
}
