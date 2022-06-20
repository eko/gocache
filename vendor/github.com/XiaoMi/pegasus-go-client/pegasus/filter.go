// Copyright (c) 2017, Xiaomi, Inc.  All rights reserved.
// This source code is licensed under the Apache License Version 2.0, which
// can be found in the LICENSE file in the root directory of this source tree.

package pegasus

import "github.com/XiaoMi/pegasus-go-client/idl/rrdb"

// FilterType defines the type of key filtering.
type FilterType int

// Filter types
const (
	FilterTypeNoFilter      = FilterType(rrdb.FilterType_FT_NO_FILTER)
	FilterTypeMatchAnywhere = FilterType(rrdb.FilterType_FT_MATCH_ANYWHERE)
	FilterTypeMatchPrefix   = FilterType(rrdb.FilterType_FT_MATCH_PREFIX)
	FilterTypeMatchPostfix  = FilterType(rrdb.FilterType_FT_MATCH_POSTFIX)
)

// Filter is used to filter based on the key.
type Filter struct {
	Type    FilterType
	Pattern []byte
}
