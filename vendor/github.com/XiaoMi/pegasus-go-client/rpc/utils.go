// Copyright (c) 2017, Xiaomi, Inc.  All rights reserved.
// This source code is licensed under the Apache License Version 2.0, which
// can be found in the LICENSE file in the root directory of this source tree.

package rpc

import (
	"io"
	"net"
)

// IsNetworkTimeoutErr returns whether the given error is a timeout error.
func IsNetworkTimeoutErr(err error) bool {
	// if it's a network timeout error
	opErr, ok := err.(*net.OpError)
	if ok {
		return opErr.Timeout()
	}

	return false
}

// IsNetworkClosed returns whether the session is shutdown by the peer.
func IsNetworkClosed(err error) bool {
	opErr, ok := err.(*net.OpError)
	if ok {
		return opErr.Err == io.EOF
	}

	return err == io.EOF
}
