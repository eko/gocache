// Copyright (c) 2017, Xiaomi, Inc.  All rights reserved.
// This source code is licensed under the Apache License Version 2.0, which
// can be found in the LICENSE file in the root directory of this source tree.

package rpc

import (
	"bufio"
	"io"
)

const (
// In our experiment(go-ycsb, 100w insertions, 100 goroutines, 100 bytes record size),
// rpc performance can significantly be improved by increasing read buffer.
// As we continue to double the buffer size from 256KB to 512KB, the throughput
// as well as average latency stop gaining improvement.
// See Issue#4 for more detail.
//
// read buffer 64kb
// INSERT - Count: 192010, Avg(us): 3482, Min(us): 386, Max(us): 42951, 95th(us): 8000, 99th(us): 14000
// INSERT - Count: 387387, Avg(us): 3447, Min(us): 356, Max(us): 45644, 95th(us): 8000, 99th(us): 14000
// INSERT - Count: 584503, Avg(us): 3412, Min(us): 356, Max(us): 45644, 95th(us): 7000, 99th(us): 13000
// INSERT - Count: 774928, Avg(us): 3438, Min(us): 356, Max(us): 45644, 95th(us): 7000, 99th(us): 13000
// INSERT - Count: 965434, Avg(us): 3451, Min(us): 338, Max(us): 77322, 95th(us): 7000, 99th(us): 13000
// INSERT - Count: 1000000, Avg(us): 3443, Min(us): 338, Max(us): 77322, 95th(us): 7000, 99th(us): 13000
// Run finished, takes 51.837521852s
//
// read buffer 128kb
// INSERT - Count: 225254, Avg(us): 3139, Min(us): 357, Max(us): 36666, 95th(us): 7000, 99th(us): 14000
// INSERT - Count: 458059, Avg(us): 3110, Min(us): 357, Max(us): 42223, 95th(us): 7000, 99th(us): 14000
// INSERT - Count: 683384, Avg(us): 3135, Min(us): 340, Max(us): 42223, 95th(us): 7000, 99th(us): 14000
// INSERT - Count: 915600, Avg(us): 3157, Min(us): 322, Max(us): 57728, 95th(us): 7000, 99th(us): 15000
// INSERT - Count: 999999, Avg(us): 3140, Min(us): 322, Max(us): 57728, 95th(us): 7000, 99th(us): 15000
// Run finished, takes 43.703584059s
//
// read buffer 256kb
// INSERT - Count: 366927, Avg(us): 2511, Min(us): 347, Max(us): 50030, 95th(us): 7000, 99th(us): 15000
// INSERT - Count: 701266, Avg(us): 2649, Min(us): 344, Max(us): 73976, 95th(us): 8000, 99th(us): 17000
// INSERT - Count: 1000000, Avg(us): 2615, Min(us): 340, Max(us): 73976, 95th(us): 8000, 99th(us): 17000
// Run finished, takes 28.381599693s
//
// read buffer 512kb
// INSERT - Count: 366486, Avg(us): 2596, Min(us): 332, Max(us): 83957, 95th(us): 8000, 99th(us): 17000
// INSERT - Count: 725917, Avg(us): 2624, Min(us): 320, Max(us): 83957, 95th(us): 8000, 99th(us): 18000
// INSERT - Count: 999999, Avg(us): 2634, Min(us): 320, Max(us): 95898, 95th(us): 8000, 99th(us): 18000
// Run finished, takes 27.91239882s

// readStreamBufferSize = 1024 * 256
)

// low-level rpc reader.
type ReadStream struct {
	bufReader *bufio.Reader
}

func (r *ReadStream) Next(toRead int) ([]byte, error) {
	buf := make([]byte, toRead)
	var total = 0

	readSz, err := r.bufReader.Read(buf)
	total += readSz
	for total < toRead && err == nil {
		readSz, err = r.bufReader.Read(buf[total:])
		total += readSz
	}

	if err != nil {
		return nil, err
	}
	return buf, nil
}

func NewReadStream(reader io.Reader) *ReadStream {
	// By default readStreamBufferSize is not used in order to save memory usage,
	// since for pegasus2, user may create a large number of replicaSession
	// (100 TableConnectors eg.).
	// TODO(wutao1): provide function to create read stream with readStreamBufferSize
	return &ReadStream{
		bufReader: bufio.NewReader(reader),
	}
}
