// Copyright (c) 2017, Xiaomi, Inc.  All rights reserved.
// This source code is licensed under the Apache License Version 2.0, which
// can be found in the LICENSE file in the root directory of this source tree.

package pegasus

import (
	"fmt"
)

// PError is the return error type of all interfaces of pegasus client.
type PError struct {
	// Err is the error that occurred during the operation.
	Err error

	// The failed operation
	Op OpType
}

// OpType is the type of operation that led to PError.
type OpType int

// Operation types
const (
	OpQueryConfig OpType = iota
	OpGet
	OpSet
	OpDel
	OpMultiDel
	OpMultiGet
	OpMultiGetRange
	OpClose
	OpMultiSet
	OpTTL
	OpExist
	OpGetScanner
	OpGetUnorderedScanners
	OpNext
	OpScannerClose
	OpCheckAndSet
	OpSortKeyCount
	OpIncr
	OpBatchGet
)

var opTypeToStringMap = map[OpType]string{
	OpQueryConfig:          "table configuration query",
	OpGet:                  "GET",
	OpSet:                  "SET",
	OpDel:                  "DEL",
	OpMultiGet:             "MULTI_GET",
	OpMultiGetRange:        "MULTI_GET_RANGE",
	OpMultiDel:             "MULTI_DEL",
	OpClose:                "Close",
	OpMultiSet:             "MULTI_SET",
	OpTTL:                  "TTL",
	OpExist:                "EXIST",
	OpGetScanner:           "GET_SCANNER",
	OpGetUnorderedScanners: "GET_UNORDERED_SCANNERS",
	OpNext:                 "SCAN_NEXT",
	OpScannerClose:         "SCANNER_CLOSE",
	OpCheckAndSet:          "CHECK_AND_SET",
	OpSortKeyCount:         "SORTKEY_COUNT",
	OpIncr:                 "INCR",
	OpBatchGet:             "BATCH_GET",
}

func (op OpType) String() string {
	return opTypeToStringMap[op]
}

func (e *PError) Error() string {
	return fmt.Sprintf("pegasus %s failed: %s", e.Op, e.Err.Error())
}
