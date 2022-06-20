// Copyright (c) 2017, Xiaomi, Inc.  All rights reserved.
// This source code is licensed under the Apache License Version 2.0, which
// can be found in the LICENSE file in the root directory of this source tree.

package pegasus

import "github.com/XiaoMi/pegasus-go-client/idl/rrdb"

// CheckType defines the types of value checking in a CAS.
type CheckType int

// The value checking types
const (
	CheckTypeNoCheck = CheckType(rrdb.CasCheckType_CT_NO_CHECK)

	// existence
	CheckTypeValueNotExist        = CheckType(rrdb.CasCheckType_CT_VALUE_NOT_EXIST)          // value is not exist
	CheckTypeValueNotExistOrEmpty = CheckType(rrdb.CasCheckType_CT_VALUE_NOT_EXIST_OR_EMPTY) // value is not exist or value is empty
	CheckTypeValueExist           = CheckType(rrdb.CasCheckType_CT_VALUE_EXIST)              // value is exist
	CheckTypeValueNotEmpty        = CheckType(rrdb.CasCheckType_CT_VALUE_NOT_EMPTY)          // value is exist and not empty

	// match
	CheckTypeMatchAnywhere = CheckType(rrdb.CasCheckType_CT_VALUE_MATCH_ANYWHERE) // operand matches anywhere in value
	CheckTypeMatchPrefix   = CheckType(rrdb.CasCheckType_CT_VALUE_MATCH_PREFIX)   // operand matches prefix in value
	CheckTypeMatchPostfix  = CheckType(rrdb.CasCheckType_CT_VALUE_MATCH_POSTFIX)  // operand matches postfix in value

	// bytes compare
	CheckTypeBytesLess           = CheckType(rrdb.CasCheckType_CT_VALUE_BYTES_LESS)             // bytes compare: value < operand
	CheckTypeBytesLessOrEqual    = CheckType(rrdb.CasCheckType_CT_VALUE_BYTES_LESS_OR_EQUAL)    // bytes compare: value <= operand
	CheckTypeBytesEqual          = CheckType(rrdb.CasCheckType_CT_VALUE_BYTES_EQUAL)            // bytes compare: value == operand
	CheckTypeBytesGreaterOrEqual = CheckType(rrdb.CasCheckType_CT_VALUE_BYTES_GREATER_OR_EQUAL) // bytes compare: value >= operand
	CheckTypeBytesGreater        = CheckType(rrdb.CasCheckType_CT_VALUE_BYTES_GREATER)          // bytes compare: value > operand

	// int compare: first transfer bytes to int64; then compare by int value
	CheckTypeIntLess           = CheckType(rrdb.CasCheckType_CT_VALUE_INT_LESS)             // int compare: value < operand
	CheckTypeIntLessOrEqual    = CheckType(rrdb.CasCheckType_CT_VALUE_INT_LESS_OR_EQUAL)    // int compare: value <= operand
	CheckTypeIntEqual          = CheckType(rrdb.CasCheckType_CT_VALUE_INT_EQUAL)            // int compare: value == operand
	CheckTypeIntGreaterOrEqual = CheckType(rrdb.CasCheckType_CT_VALUE_INT_GREATER_OR_EQUAL) // int compare: value >= operand
	CheckTypeIntGreater        = CheckType(rrdb.CasCheckType_CT_VALUE_BYTES_GREATER)        // int compare: value > operand
)

// CheckAndSetResult is the result of a CAS.
type CheckAndSetResult struct {
	// true if set value succeed.
	SetSucceed bool

	// the actual value if set value failed; null means the actual value is not exist.
	CheckValue []byte

	// if the check value is exist; can be used only when checkValueReturned is true.
	CheckValueExist bool

	// return the check value if exist; can be used only when checkValueExist is true.
	CheckValueReturned bool
}

// CheckAndSetOptions is the options of a CAS.
type CheckAndSetOptions struct {
	SetValueTTLSeconds int  // time to live in seconds of the set value, 0 means no ttl.
	ReturnCheckValue   bool // if return the check value in results.
}
