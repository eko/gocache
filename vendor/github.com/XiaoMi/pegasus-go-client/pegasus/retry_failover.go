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

package pegasus

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type tableRPCOp func() (confUpdated bool, result interface{}, err error)

// retryFailOver retries the operation when it encounters replica fail-over, until context reaches deadline.
func retryFailOver(ctx context.Context, op tableRPCOp) (interface{}, error) {
	bf := backoff.NewExponentialBackOff()
	bf.InitialInterval = time.Second
	bf.Multiplier = 2
	for {
		confUpdated, res, err := op()
		backoffCh := time.After(bf.NextBackOff())
		if confUpdated { // must fail
			select {
			case <-backoffCh:
				continue
			case <-ctx.Done():
				err = ctx.Err()
				break
			}
		}
		return res, err
	}
}
