// Copyright (c) 2017, Xiaomi, Inc.  All rights reserved.
// This source code is licensed under the Apache License Version 2.0, which
// can be found in the LICENSE file in the root directory of this source tree.

package session

import (
	"encoding/binary"
	"fmt"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
)

var thriftHeaderTypeStr = []byte{'T', 'H', 'F', 'T'}
var thriftHeaderBytesLen = 48

// thriftHeader stores the meta information of a particular RPC
type thriftHeader struct {
	headerVersion  uint32
	headerLength   uint32
	headerCrc32    uint32
	bodyLength     uint32
	bodyCrc32      uint32
	appId          int32
	partitionIndex int32
	clientTimeout  uint32
	threadHash     int32
	partitionHash  uint32
}

// Serialized this struct as the message header in pegasus messaging protocol.
// (See https://github.com/XiaoMi/pegasus/blob/master/docs/client-development.md)
func (t *thriftHeader) marshall(buf []byte) {
	if len(buf) != thriftHeaderBytesLen {
		panic(fmt.Sprintf("length of buf(%d) should be %d", len(buf), thriftHeaderBytesLen))
	}

	copy(buf[0:4], thriftHeaderTypeStr)
	binary.BigEndian.PutUint32(buf[4:8], t.headerVersion)
	binary.BigEndian.PutUint32(buf[8:12], t.headerLength)
	binary.BigEndian.PutUint32(buf[12:16], t.headerCrc32)
	binary.BigEndian.PutUint32(buf[16:20], t.bodyLength)
	binary.BigEndian.PutUint32(buf[20:24], t.bodyCrc32)
	binary.BigEndian.PutUint32(buf[24:28], uint32(t.appId))
	binary.BigEndian.PutUint32(buf[28:32], uint32(t.partitionIndex))
	binary.BigEndian.PutUint32(buf[32:36], t.clientTimeout)
	binary.BigEndian.PutUint32(buf[36:40], uint32(t.threadHash))
	binary.BigEndian.PutUint32(buf[40:48], t.partitionHash)
}

// Thread hash is a rDSN required header field. We copied the algorithm
// from java client.
func gpidToThreadHash(gpid *base.Gpid) int32 {
	return gpid.Appid*7919 + gpid.PartitionIndex
}
