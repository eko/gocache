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

// Incr inherits op.Request.
type Incr struct {
	HashKey   []byte
	SortKey   []byte
	Increment int64

	req *rrdb.IncrRequest
}

// Validate arguments.
func (r *Incr) Validate() error {
	if err := validateHashKey(r.HashKey); err != nil {
		return err
	}
	if err := validateSortKey(r.SortKey); err != nil {
		return err
	}

	r.req = rrdb.NewIncrRequest()
	r.req.Key = encodeHashKeySortKey(r.HashKey, r.SortKey)
	r.req.Increment = r.Increment
	return nil
}

// Run operation.
func (r *Incr) Run(ctx context.Context, gpid *base.Gpid, rs *session.ReplicaSession) (interface{}, error) {
	resp, err := rs.Incr(ctx, gpid, r.req)
	if err := wrapRPCFailure(resp, err); err != nil {
		return 0, err
	}
	return resp.NewValue_, nil
}
