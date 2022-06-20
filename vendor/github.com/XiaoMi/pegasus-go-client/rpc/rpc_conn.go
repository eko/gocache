// Copyright (c) 2017, Xiaomi, Inc.  All rights reserved.
// This source code is licensed under the Apache License Version 2.0, which
// can be found in the LICENSE file in the root directory of this source tree.

package rpc

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/XiaoMi/pegasus-go-client/pegalog"
)

// TODO(wutao1): make these parameters configurable
const (
	ConnDialTimeout  = time.Second * 3
	ConnReadTimeout  = 30 * time.Second
	ConnWriteTimeout = 10 * time.Second
)

type ConnState int

const (
	// The state that a connection starts from.
	ConnStateInit ConnState = iota

	ConnStateConnecting

	ConnStateReady

	// The state that indicates some error occurred in the previous operations.
	ConnStateTransientFailure

	// The state that RpcConn will turn into after Close() is called.
	ConnStateClosed
)

func (s ConnState) String() string {
	switch s {
	case ConnStateInit:
		return "ConnStateInit"
	case ConnStateConnecting:
		return "ConnStateConnecting"
	case ConnStateReady:
		return "ConnStateReady"
	case ConnStateTransientFailure:
		return "ConnStateTransientFailure"
	case ConnStateClosed:
		return "ConnStateClosed"
	default:
		panic("no such state")
	}
}

var ErrConnectionNotReady = errors.New("connection is not ready")

// RpcConn maintains a network connection to a particular endpoint.
type RpcConn struct {
	Endpoint string

	wstream *WriteStream
	rstream *ReadStream
	conn    net.Conn

	writeTimeout time.Duration
	readTimeout  time.Duration

	cstate ConnState
	mu     sync.RWMutex

	logger pegalog.Logger
}

// thread-safe
func (rc *RpcConn) GetState() ConnState {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.cstate
}

// thread-safe
func (rc *RpcConn) setState(state ConnState) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cstate = state
}

// This function is thread-safe.
func (rc *RpcConn) TryConnect() (err error) {
	err = func() error {
		// set state to ConnStateConnecting to
		// make sure there's only 1 goroutine dialing simultaneously.
		rc.mu.Lock()
		defer rc.mu.Unlock()
		if rc.cstate != ConnStateReady && rc.cstate != ConnStateConnecting {
			rc.cstate = ConnStateConnecting
			rc.mu.Unlock()

			// unlock for blocking call
			d := &net.Dialer{
				Timeout: ConnDialTimeout,
			}
			conn, err := d.Dial("tcp", rc.Endpoint)

			rc.mu.Lock()
			rc.conn = conn
			if err != nil {
				return err
			}
			tcpConn, _ := rc.conn.(*net.TCPConn)
			tcpConn.SetNoDelay(true)
			rc.setReady(rc.conn, rc.conn)
		}
		return err
	}()

	if err != nil {
		rc.setState(ConnStateTransientFailure)
	}
	return err
}

// This function is thread-safe.
func (rc *RpcConn) Close() (err error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.cstate = ConnStateClosed
	if rc.conn != nil {
		err = rc.conn.Close()
	}

	return
}

func (rc *RpcConn) Write(msgBytes []byte) (err error) {
	err = func() error {
		if rc.GetState() != ConnStateReady {
			return ErrConnectionNotReady
		}

		tcpConn, ok := rc.conn.(*net.TCPConn)
		if ok {
			tcpConn.SetWriteDeadline(time.Now().Add(rc.writeTimeout))
		}

		return rc.wstream.Write(msgBytes)
	}()

	if err != nil {
		rc.setState(ConnStateTransientFailure)
	}
	return err
}

// Read is not intended to be cancellable using context by outside user.
// The only approach to cancel the operation is to close the connection.
// If the current socket is not well established for reading, the operation will
// fail and return error immediately.
// This function is not-thread-safe, because the underlying TCP IO buffer
// is not-thread-safe. Package users should call Read in a single goroutine.
func (rc *RpcConn) Read(size int) (bytes []byte, err error) {
	bytes, err = func() ([]byte, error) {
		if rc.GetState() != ConnStateReady {
			return nil, ErrConnectionNotReady
		}

		tcpConn, ok := rc.conn.(*net.TCPConn)
		if ok {
			tcpConn.SetReadDeadline(time.Now().Add(rc.readTimeout))
		}

		bytes, err = rc.rstream.Next(size)
		return bytes, err
	}()

	if err != nil && !IsNetworkTimeoutErr(err) {
		rc.setState(ConnStateTransientFailure)
	}
	return bytes, err
}

// Returns an idle connection.
func NewRpcConn(addr string) *RpcConn {
	return &RpcConn{
		Endpoint:     addr,
		logger:       pegalog.GetLogger(),
		cstate:       ConnStateInit,
		readTimeout:  ConnReadTimeout,
		writeTimeout: ConnWriteTimeout,
	}
}

// Not thread-safe
func (rc *RpcConn) SetWriteTimeout(timeout time.Duration) {
	rc.writeTimeout = timeout
}

// Not thread-safe
func (rc *RpcConn) SetReadTimeout(timeout time.Duration) {
	rc.readTimeout = timeout
}

func (rc *RpcConn) setReady(reader io.Reader, writer io.Writer) {
	rc.cstate = ConnStateReady
	rc.rstream = NewReadStream(reader)
	rc.wstream = NewWriteStream(writer)
}

// Create a fake client with specified reader and writer.
func NewFakeRpcConn(reader io.Reader, writer io.Writer) *RpcConn {
	conn := NewRpcConn("")
	conn.setReady(reader, writer)
	return conn
}
