// Copyright (c) 2017, Xiaomi, Inc.  All rights reserved.
// This source code is licensed under the Apache License Version 2.0, which
// can be found in the LICENSE file in the root directory of this source tree.

package base

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/pegasus-kv/thrift/lib/go/thrift"
)

type RPCAddress struct {
	address int64
}

func NewRPCAddress(ip net.IP, port int) *RPCAddress {
	return &RPCAddress{
		address: (int64(binary.BigEndian.Uint32(ip.To4())) << 32) + (int64(port) << 16) + 1,
	}
}

func (r *RPCAddress) Read(iprot thrift.TProtocol) error {
	address, err := iprot.ReadI64()
	if err != nil {
		return err
	}
	r.address = address
	return nil
}

func (r *RPCAddress) Write(oprot thrift.TProtocol) error {
	return oprot.WriteI64(r.address)
}

func (r *RPCAddress) String() string {
	if r == nil {
		return "<nil>"
	}
	return fmt.Sprintf("RPCAddress(%s)", r.GetAddress())
}

func (r *RPCAddress) getIp() net.IP {
	return net.IPv4(byte(0xff&(r.address>>56)), byte(0xff&(r.address>>48)), byte(0xff&(r.address>>40)), byte(0xff&(r.address>>32)))
}

func (r *RPCAddress) getPort() int {
	return int(0xffff & (r.address >> 16))
}

func (r *RPCAddress) GetAddress() string {
	return fmt.Sprintf("%s:%d", r.getIp(), r.getPort())
}

func (r *RPCAddress) GetRawAddress() int64 {
	return r.address
}
