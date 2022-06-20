// Copyright (c) 2017, Xiaomi, Inc.  All rights reserved.
// This source code is licensed under the Apache License Version 2.0, which
// can be found in the LICENSE file in the root directory of this source tree.

package base

import (
	"fmt"

	"github.com/pegasus-kv/thrift/lib/go/thrift"
)

type Gpid struct {
	Appid, PartitionIndex int32
}

func (id *Gpid) Read(iprot thrift.TProtocol) error {
	v, err := iprot.ReadI64()
	if err != nil {
		return err
	}

	id.Appid = int32(v & int64(0x00000000ffffffff))
	id.PartitionIndex = int32(v >> 32)
	return nil
}

func (id *Gpid) Write(oprot thrift.TProtocol) error {
	v := int64(id.Appid) + int64(id.PartitionIndex)<<32
	return oprot.WriteI64(v)
}

func (id *Gpid) String() string {
	if id == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%+v", *id)
}
