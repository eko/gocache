// Copyright (c) 2017, Xiaomi, Inc.  All rights reserved.
// This source code is licensed under the Apache License Version 2.0, which
// can be found in the LICENSE file in the root directory of this source tree.

package base

import (
	"fmt"

	"github.com/pegasus-kv/thrift/lib/go/thrift"
)

type Blob struct {
	Data []byte
}

func (b *Blob) Read(iprot thrift.TProtocol) error {
	data, err := iprot.ReadBinary()
	if err != nil {
		return err
	}
	b.Data = data
	return nil
}

func (b *Blob) Write(oprot thrift.TProtocol) error {
	return oprot.WriteBinary(b.Data)
}

func (b *Blob) String() string {
	if b == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Blob(%+v)", *b)
}

func NewBlob() *Blob {
	return &Blob{}
}
