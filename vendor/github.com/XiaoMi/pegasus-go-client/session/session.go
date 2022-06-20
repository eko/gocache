// Copyright (c) 2017, Xiaomi, Inc.  All rights reserved.
// This source code is licensed under the Apache License Version 2.0, which
// can be found in the LICENSE file in the root directory of this source tree.

package session

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/pegalog"
	"github.com/XiaoMi/pegasus-go-client/rpc"
	"gopkg.in/tomb.v2"
)

// NodeType represents the type of the NodeSession.
type NodeType string

const (
	// NodeTypeMeta indicates it's a session to MetaServer.
	NodeTypeMeta NodeType = "meta"

	// NodeTypeReplica indicates it's a session to ReplicaServer.
	NodeTypeReplica NodeType = "replica"

	kDialInterval = time.Second * 60

	// LatencyTracingThreshold means RPC's latency higher than the threshold (1000ms) will be traced
	LatencyTracingThreshold = time.Millisecond * 1000
)

// NodeSession represents the network session to a node
// (either a meta server or a replica server).
// It encapsulates the internal rpc processing, including
// network communication and message (de)serialization.
type NodeSession interface {
	String() string

	// Invoke an rpc call.
	CallWithGpid(ctx context.Context, gpid *base.Gpid, args RpcRequestArgs, name string) (result RpcResponseResult, err error)

	// Get connection state.
	ConnState() rpc.ConnState

	Close() error
}

// NodeSessionCreator creates an instance of NodeSession,
// receiving argument `string` as host address, `NodeType`
// as the type of the node.
type NodeSessionCreator func(string, NodeType) NodeSession

// An implementation of NodeSession.
type nodeSession struct {
	logger pegalog.Logger

	// atomic incremented counter that ensures each rpc
	// has a unique sequence id
	seqId int32

	addr  string
	ntype NodeType
	conn  *rpc.RpcConn

	tom *tomb.Tomb

	reqc        chan *requestListener
	pendingResp map[int32]*requestListener
	mu          sync.Mutex

	redialc      chan bool
	lastDialTime time.Time

	codec rpc.Codec

	unresponsiveHandler UnresponsiveHandler
	lastWriteTime       int64
}

// withUnresponsiveHandler enables the session to handle the event when a network connection becomes unresponsive.
func withUnresponsiveHandler(s NodeSession, handler UnresponsiveHandler) {
	ns, ok := s.(*nodeSession)
	if !ok {
		return
	}
	ns.unresponsiveHandler = handler
}

type requestListener struct {
	ch   chan bool
	call *PegasusRpcCall
}

func newNodeSessionAddr(addr string, ntype NodeType) *nodeSession {
	return &nodeSession{
		logger:      pegalog.GetLogger(),
		ntype:       ntype,
		seqId:       0,
		codec:       NewPegasusCodec(),
		pendingResp: make(map[int32]*requestListener),
		reqc:        make(chan *requestListener),
		addr:        addr,
		tom:         &tomb.Tomb{},

		//
		redialc: make(chan bool, 1),
	}
}

// NewNodeSession always returns a non-nil value even when the
// connection attempt failed.
// Each nodeSession corresponds to an RpcConn.
func NewNodeSession(addr string, ntype NodeType) NodeSession {
	return newNodeSession(addr, ntype)
}

func newNodeSession(addr string, ntype NodeType) *nodeSession {
	logger := pegalog.GetLogger()

	n := newNodeSessionAddr(addr, ntype)
	logger.Printf("create session with %s", n)

	n.conn = rpc.NewRpcConn(addr)

	n.tom.Go(n.loopForDialing)
	return n
}

// thread-safe
func (n *nodeSession) ConnState() rpc.ConnState {
	return n.conn.GetState()
}

func (n *nodeSession) String() string {
	return fmt.Sprintf("[%s(%s)]", n.addr, n.ntype)
}

// Loop in background and keep watching for redialc.
// Since loopForDialing is the only consumer of redialc, it guarantees
// only 1 goroutine dialing simultaneously.
// This goroutine will not be killed due to io failure, unless the session
// is manually closed.
func (n *nodeSession) loopForDialing() error { // no error returned actually
	for {
		select {
		case <-n.tom.Dying():
			return nil
		case <-n.redialc:
			if n.ConnState() != rpc.ConnStateReady {
				n.dial()
			}
		}
	}
}

func (n *nodeSession) tryDial() {
	select {
	case n.redialc <- true:
	default:
	}
}

// If the dialing ended successfully, it will start loopForRequest and
// loopForResponse which handle the data communications.
// If the last attempt failed, it will retry again.
func (n *nodeSession) dial() {
	if time.Now().Sub(n.lastDialTime) < kDialInterval {
		select {
		case <-time.After(kDialInterval):
		case <-n.tom.Dying():
			return
		}
	}

	select {
	case <-n.tom.Dying():
		// ended if session closed.
	default:
		n.logger.Print("dial to ", n)
		err := n.conn.TryConnect()
		n.lastDialTime = time.Now()

		if err != nil {
			n.logger.Printf("failed to dial %s: %s", n, err)
		} else {
			n.tom.Go(n.loopForRequest)
			n.tom.Go(n.loopForResponse)
		}
	}

	n.logger.Printf("stop dialing for %s, connection state: %s", n, n.ConnState())
}

func (n *nodeSession) notifyCallerAndDrop(req *requestListener) {
	select {
	// notify the caller
	case req.ch <- true:
		n.mu.Lock()
		delete(n.pendingResp, req.call.SeqId)
		n.mu.Unlock()
	default:
		panic("impossible for concurrent notifiers")
	}
}

// single-routine worker used for sending requests.
// Any error occurred will end up this goroutine as well as the connection.
func (n *nodeSession) loopForRequest() error { // no error returned actually
	for {
		select {
		case <-n.tom.Dying():
			return nil
		case req := <-n.reqc:
			n.mu.Lock()
			n.pendingResp[req.call.SeqId] = req
			n.mu.Unlock()

			atomic.StoreInt64(&n.lastWriteTime, time.Now().UnixNano())
			req.call.OnRpcSend = time.Now()
			if err := n.writeRequest(req.call); err != nil {
				n.logger.Printf("failed to send request to %s: %s", n, err)

				// notify the rpc caller.
				req.call.Err = err
				n.notifyCallerAndDrop(req)

				// don give up if there's still hope
				if !rpc.IsNetworkTimeoutErr(err) {
					return nil
				}
			}
		}
	}
}

// hasRecentUnresponsiveWrite returns if session is active in sending tcp request but gets no response.
func (n *nodeSession) hasRecentUnresponsiveWrite() bool {
	// 10s is usually the max limit that the server promises to respond.
	var unresponsiveThreshold = int64(math.Max(float64(rpc.ConnReadTimeout.Nanoseconds()/2), float64(time.Second.Nanoseconds()*10)))
	return time.Now().UnixNano()-atomic.LoadInt64(&n.lastWriteTime) < unresponsiveThreshold
}

// single-routine worker used for reading response.
// We register a map of sequence id -> recvItem for each coming request,
// so that when a response is received, we are able to notify its caller.
// Any un-retryable error occurred will end up this goroutine.
func (n *nodeSession) loopForResponse() error { // no error returned actually
	for {
		select {
		case <-n.tom.Dying():
			return nil
		default:
		}

		call, err := n.readResponse()
		if err != nil {
			if rpc.IsNetworkTimeoutErr(err) {
				// If a session encounters a read-timeout, and it's simultaneously writing (depends on lastWriteTime),
				// this sesion is considered as unresponsive.
				// When in this state, it's in very danger with potential network failure.
				if n.unresponsiveHandler != nil && n.hasRecentUnresponsiveWrite() {
					n.unresponsiveHandler(n)
				}
				continue // retry if no data to read
			}
			if rpc.IsNetworkClosed(err) { // EOF
				n.logger.Printf("session %s is closed by the peer", n)
				return nil
			}
			n.logger.Printf("failed to read response from %s: %s", n, err)
			return nil
		}
		call.OnRpcRecv = time.Now()

		n.mu.Lock()
		reqListener, ok := n.pendingResp[call.SeqId]
		n.mu.Unlock()

		if !ok {
			n.logger.Printf("ignore stale response (seqId: %d) from %s: %s",
				call.SeqId, n, call.Result)
			continue
		}

		reqListener.call.Err = call.Err
		reqListener.call.Result = call.Result
		reqListener.call.OnRpcRecv = call.OnRpcRecv
		n.notifyCallerAndDrop(reqListener)
	}
}

func (n *nodeSession) waitUntilSessionReady(ctx context.Context) error {
	if n.ConnState() != rpc.ConnStateReady {
		dialStart := time.Now()

		n.tryDial()

		var ready bool
		ticker := time.NewTicker(1 * time.Millisecond) // polling 1ms each time to minimize the connection time.
		defer ticker.Stop()
		for {
			breakLoop := false
			select {
			case <-ctx.Done(): // exceeds the user timeout, or this context is cancelled, or the session transiently failed.
				breakLoop = true
			case <-ticker.C:
				if n.ConnState() == rpc.ConnStateReady {
					ready = true
					breakLoop = true
				}
			}
			if breakLoop {
				break
			}
		}

		if !ready {
			return fmt.Errorf("session %s is unable to connect (used %dms), the context error: %s", n, time.Since(dialStart)/time.Millisecond, ctx.Err())
		}
	}
	return nil
}

func (n *nodeSession) CallWithGpid(ctx context.Context, gpid *base.Gpid, args RpcRequestArgs, name string) (result RpcResponseResult, err error) {
	// either the ctx cancelled or the tomb killed will stop this rpc call.
	ctxWithTomb := n.tom.Context(ctx)
	if err := n.waitUntilSessionReady(ctxWithTomb); err != nil {
		return nil, err
	}

	seqId := atomic.AddInt32(&n.seqId, 1) // increment sequence id
	rcall, err := MarshallPegasusRpc(n.codec, seqId, gpid, args, name)
	if err != nil {
		return nil, err
	}
	rcall.OnRpcCall = time.Now()

	req := &requestListener{call: rcall, ch: make(chan bool, 1)}

	defer func() {
		// manually trigger gc
		rcall = nil
		req = nil
	}()

	select {
	// passes the request to loopForRequest
	case n.reqc <- req:
		select {
		// receive from loopForResponse, or loopRequest failed
		case <-req.ch:
			err = rcall.Err
			result = rcall.Result
			if rcall.TilNow() > LatencyTracingThreshold {
				n.logger.Printf("[%s(%s)] trace to %s: %s", rcall.Name, rcall.Gpid, n, rcall.Trace())
			}
			return
		case <-ctxWithTomb.Done():
			err = ctxWithTomb.Err()
			result = nil
			return
		}
	case <-ctxWithTomb.Done():
		err = ctxWithTomb.Err()
		result = nil
		return
	}
}

func (n *nodeSession) writeRequest(r *PegasusRpcCall) error {
	return n.conn.Write(r.RawReq)
}

// readResponse never returns nil `PegasusRpcCall` unless the tcp round trip failed.
// The pegasus server may in some cases respond with a not-ERR_OK error code (together with
// sequence id and rpc name) while without a transport-layer failure.
func (n *nodeSession) readResponse() (*PegasusRpcCall, error) {
	return ReadRpcResponse(n.conn, n.codec)
}

func (n *nodeSession) Close() error {
	n.mu.Lock()
	if n.ConnState() != rpc.ConnStateClosed {
		n.logger.Printf("close session %s", n)
		n.conn.Close()
		n.tom.Kill(nil)
	}
	n.mu.Unlock()

	return n.tom.Wait()
}
