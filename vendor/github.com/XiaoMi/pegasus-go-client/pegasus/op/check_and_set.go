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

// CheckAndSet inherits op.Request.
type CheckAndSet struct {
	Req *rrdb.CheckAndSetRequest
}

// CheckAndSetResult is the result of a CAS.
type CheckAndSetResult struct {
	SetSucceed         bool
	CheckValue         []byte
	CheckValueExist    bool
	CheckValueReturned bool
}

// Validate arguments.
func (r *CheckAndSet) Validate() error {
	if err := validateHashKey(r.Req.HashKey.Data); err != nil {
		return err
	}
	return nil
}

// Run operation.
func (r *CheckAndSet) Run(ctx context.Context, gpid *base.Gpid, rs *session.ReplicaSession) (interface{}, error) {
	resp, err := rs.CheckAndSet(ctx, gpid, r.Req)
	err = wrapRPCFailure(resp, err)
	if err == base.TryAgain {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	result := &CheckAndSetResult{
		SetSucceed:         resp.Error == 0,
		CheckValueReturned: resp.CheckValueReturned,
		CheckValueExist:    resp.CheckValueReturned && resp.CheckValueExist,
	}
	if resp.CheckValueReturned && resp.CheckValueExist && resp.CheckValue != nil && resp.CheckValue.Data != nil && len(resp.CheckValue.Data) != 0 {
		result.CheckValue = resp.CheckValue.Data
	}
	return result, nil
}
