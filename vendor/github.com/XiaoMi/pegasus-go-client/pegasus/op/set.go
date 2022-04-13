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
	"time"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/rrdb"
	"github.com/XiaoMi/pegasus-go-client/session"
)

// Set inherits op.Request.
type Set struct {
	HashKey []byte
	SortKey []byte
	Value   []byte
	TTL     time.Duration

	req *rrdb.UpdateRequest
}

// Validate arguments.
func (r *Set) Validate() error {
	if err := validateHashKey(r.HashKey); err != nil {
		return err
	}
	if err := validateSortKey(r.SortKey); err != nil {
		return err
	}
	if err := validateValue(r.Value); err != nil {
		return err
	}

	key := encodeHashKeySortKey(r.HashKey, r.SortKey)
	val := &base.Blob{Data: r.Value}
	expireTsSec := int32(0)
	if r.TTL != 0 {
		expireTsSec = expireTsSeconds(r.TTL)
	}
	r.req = &rrdb.UpdateRequest{Key: key, Value: val, ExpireTsSeconds: expireTsSec}
	return nil
}

// Run operation.
func (r *Set) Run(ctx context.Context, gpid *base.Gpid, rs *session.ReplicaSession) (interface{}, error) {
	resp, err := rs.Put(ctx, gpid, r.req)
	if err := wrapRPCFailure(resp, err); err != nil {
		return 0, err
	}
	return nil, nil
}
