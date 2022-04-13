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
	"github.com/XiaoMi/pegasus-go-client/session"
)

// TTL inherits op.Request.
type TTL struct {
	HashKey []byte
	SortKey []byte

	req *base.Blob
}

// Validate arguments.
func (r *TTL) Validate() error {
	if err := validateHashKey(r.HashKey); err != nil {
		return err
	}
	if err := validateSortKey(r.SortKey); err != nil {
		return err
	}
	r.req = encodeHashKeySortKey(r.HashKey, r.SortKey)
	return nil
}

// Run operation.
// Returns -2 if entry doesn't exist.
func (r *TTL) Run(ctx context.Context, gpid *base.Gpid, rs *session.ReplicaSession) (interface{}, error) {
	resp, err := rs.TTL(ctx, gpid, r.req)
	err = wrapRPCFailure(resp, err)
	if err == base.NotFound {
		// Success for non-existed entry.
		return -2, nil
	}
	if err != nil {
		return -2, err
	}
	return int(resp.GetTTLSeconds()), nil
}
