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
	"fmt"
	"time"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/XiaoMi/pegasus-go-client/session"
)

// MultiSet inherits op.Request.
type MultiSet struct {
	HashKey  []byte
	SortKeys [][]byte
	Values   [][]byte
	TTL      time.Duration

	req *rrdb.MultiPutRequest
}

// Validate arguments.
func (r *MultiSet) Validate() error {
	if err := validateHashKey(r.HashKey); err != nil {
		return err
	}
	if err := validateSortKeys(r.SortKeys); err != nil {
		return err
	}
	if err := validateValues(r.Values); err != nil {
		return err
	}
	if len(r.SortKeys) != len(r.Values) {
		return fmt.Errorf("InvalidParameter: unmatched key-value pairs: len(sortKeys)=%d len(values)=%d",
			len(r.SortKeys), len(r.Values))
	}

	r.req = rrdb.NewMultiPutRequest()
	r.req.HashKey = &base.Blob{Data: r.HashKey}
	r.req.Kvs = make([]*rrdb.KeyValue, len(r.SortKeys))
	for i := 0; i < len(r.SortKeys); i++ {
		r.req.Kvs[i] = &rrdb.KeyValue{
			Key:   &base.Blob{Data: r.SortKeys[i]},
			Value: &base.Blob{Data: r.Values[i]},
		}
	}
	r.req.ExpireTsSeconds = 0
	if r.TTL != 0 {
		r.req.ExpireTsSeconds = expireTsSeconds(r.TTL)
	}
	return nil
}

// Run operation.
func (r *MultiSet) Run(ctx context.Context, gpid *base.Gpid, rs *session.ReplicaSession) (interface{}, error) {
	resp, err := rs.MultiSet(ctx, gpid, r.req)
	if err := wrapRPCFailure(resp, err); err != nil {
		return nil, err
	}
	return nil, nil
}
