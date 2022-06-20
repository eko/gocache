/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package op

import (
	"context"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/XiaoMi/pegasus-go-client/session"
)

// MultiGet inherits op.Request.
type MultiGet struct {
	HashKey      []byte
	SortKeys     [][]byte
	StartSortkey []byte
	StopSortkey  []byte

	Req *rrdb.MultiGetRequest
}

// MultiGetResult
type MultiGetResult struct {
	KVs        []*rrdb.KeyValue
	AllFetched bool
}

// Validate arguments.
func (r *MultiGet) Validate() error {
	if err := validateHashKey(r.HashKey); err != nil {
		return err
	}
	if len(r.SortKeys) != 0 {
		// sortKeys are nil-able, nil means fetching all entries.
		if err := validateSortKeys(r.SortKeys); err != nil {
			return err
		}
		r.Req.SorkKeys = make([]*base.Blob, len(r.SortKeys))
		for i, sortKey := range r.SortKeys {
			r.Req.SorkKeys[i] = &base.Blob{Data: sortKey}
		}
	}
	r.Req.HashKey = &base.Blob{Data: r.HashKey}
	if r.StartSortkey == nil {
		r.Req.StartSortkey = &base.Blob{}
	} else {
		r.Req.StartSortkey = &base.Blob{Data: r.StartSortkey}
	}
	if r.StopSortkey == nil {
		r.Req.StopSortkey = &base.Blob{}
	} else {
		r.Req.StopSortkey = &base.Blob{Data: r.StopSortkey}
	}
	return nil
}

// Run operation.
func (r *MultiGet) Run(ctx context.Context, gpid *base.Gpid, rs *session.ReplicaSession) (interface{}, error) {
	resp, err := rs.MultiGet(ctx, gpid, r.Req)
	err = wrapRPCFailure(resp, err)
	allFetched := true
	if err == base.Incomplete {
		// partial data is fetched
		allFetched = false
		err = nil
	}
	if err != nil {
		return nil, err
	}
	return &MultiGetResult{KVs: resp.Kvs, AllFetched: allFetched}, nil
}
