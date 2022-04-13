// Copyright (c) 2017, Xiaomi, Inc.  All rights reserved.
// This source code is licensed under the Apache License Version 2.0, which
// can be found in the LICENSE file in the root directory of this source tree.

package rpc

import (
	"io"
)

// low-level rpc writer.
type WriteStream struct {
	writer io.Writer
}

// NewWriteStream always receives a *net.TcpConn as `writer`, except in
// testing it can accept a buffer as the fake writer.
func NewWriteStream(writer io.Writer) *WriteStream {
	return &WriteStream{
		writer: writer,
	}
}

// invoke an asynchronous write for message.
func (s *WriteStream) Write(msgBytes []byte) error {
	var err error
	var total = 0
	var written = 0

	toWrite := len(msgBytes)

	for total < toWrite && err == nil {
		written, err = s.writer.Write(msgBytes[total:])
		total += written
	}

	return err
}
