package base

import (
	"fmt"

	"github.com/pegasus-kv/thrift/lib/go/thrift"
)

/// Primitive for Pegasus thrift framework.
type ErrorCode struct {
	Errno string
}

// How to generate the map from string to error codes?
// First:
//  - go get github.com/alvaroloes/enumer
// Second:
//  - cd idl/base
//  - enumer -type=DsnErrCode -output=dsn_err_string.go

//go:generate enumer -type=DsnErrCode -output=err_type_string.go
type DsnErrCode int32

const (
	ERR_OK DsnErrCode = iota
	ERR_UNKNOWN
	ERR_REPLICATION_FAILURE
	ERR_APP_EXIST
	ERR_APP_NOT_EXIST
	ERR_APP_DROPPED
	ERR_BUSY_CREATING
	ERR_BUSY_DROPPING
	ERR_EXPIRED
	ERR_LOCK_ALREADY_EXIST
	ERR_HOLD_BY_OTHERS
	ERR_RECURSIVE_LOCK
	ERR_NO_OWNER
	ERR_NODE_ALREADY_EXIST
	ERR_INCONSISTENT_STATE
	ERR_ARRAY_INDEX_OUT_OF_RANGE
	ERR_SERVICE_NOT_FOUND
	ERR_SERVICE_ALREADY_RUNNING
	ERR_IO_PENDING
	ERR_TIMEOUT
	ERR_SERVICE_NOT_ACTIVE
	ERR_BUSY
	ERR_NETWORK_INIT_FAILED
	ERR_FORWARD_TO_OTHERS
	ERR_OBJECT_NOT_FOUND
	ERR_HANDLER_NOT_FOUND
	ERR_LEARN_FILE_FAILED
	ERR_GET_LEARN_STATE_FAILED
	ERR_INVALID_VERSION
	ERR_INVALID_PARAMETERS
	ERR_CAPACITY_EXCEEDED
	ERR_INVALID_STATE
	ERR_INACTIVE_STATE
	ERR_NOT_ENOUGH_MEMBER
	ERR_FILE_OPERATION_FAILED
	ERR_HANDLE_EOF
	ERR_WRONG_CHECKSUM
	ERR_INVALID_DATA
	ERR_INVALID_HANDLE
	ERR_INCOMPLETE_DATA
	ERR_VERSION_OUTDATED
	ERR_PATH_NOT_FOUND
	ERR_PATH_ALREADY_EXIST
	ERR_ADDRESS_ALREADY_USED
	ERR_STATE_FREEZED
	ERR_LOCAL_APP_FAILURE
	ERR_BIND_IOCP_FAILED
	ERR_NETWORK_START_FAILED
	ERR_NOT_IMPLEMENTED
	ERR_CHECKPOINT_FAILED
	ERR_WRONG_TIMING
	ERR_NO_NEED_OPERATE
	ERR_CORRUPTION
	ERR_TRY_AGAIN
	ERR_CLUSTER_NOT_FOUND
	ERR_CLUSTER_ALREADY_EXIST
	ERR_SERVICE_ALREADY_EXIST
	ERR_INJECTED
	ERR_NETWORK_FAILURE
	ERR_UNDER_RECOVERY
	ERR_OPERATION_DISABLED
	ERR_ZOOKEEPER_OPERATION
)

func (e DsnErrCode) Error() string {
	return fmt.Sprintf("[%s]", e.String())
}

func (ec *ErrorCode) Read(iprot thrift.TProtocol) (err error) {
	ec.Errno, err = iprot.ReadString()
	return
}

func (ec *ErrorCode) Write(oprot thrift.TProtocol) error {
	return oprot.WriteString(ec.Errno)
}

func (ec *ErrorCode) String() string {
	if ec == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ErrorCode(%+v)", *ec)
}

//go:generate enumer -type=RocksDBErrCode -output=rocskdb_err_string.go
type RocksDBErrCode int32

const (
	Ok RocksDBErrCode = iota
	NotFound
	Corruption
	NotSupported
	InvalidArgument
	IOError
	MergeInProgress
	Incomplete
	ShutdownInProgress
	TimedOut
	Aborted
	Busy
	Expired
	TryAgain
)

func NewRocksDBErrFromInt(e int32) error {
	err := RocksDBErrCode(e)
	if err == Ok {
		return nil
	}
	return err
}

func (e RocksDBErrCode) Error() string {
	return fmt.Sprintf("ROCSKDB_ERR(%s)", e.String())
}
