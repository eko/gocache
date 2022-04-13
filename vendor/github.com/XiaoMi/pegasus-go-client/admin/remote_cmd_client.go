/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package admin

import (
	"context"

	"github.com/XiaoMi/pegasus-go-client/idl/base"
	"github.com/XiaoMi/pegasus-go-client/idl/cmd"
	"github.com/XiaoMi/pegasus-go-client/session"
)

// RemoteCmdClient is a client to call remote command to a PegasusServer.
type RemoteCmdClient struct {
	session session.NodeSession
}

// NewRemoteCmdClient returns an instance of RemoteCmdClient.
func NewRemoteCmdClient(addr string, nodeType session.NodeType) *RemoteCmdClient {
	return &RemoteCmdClient{
		session: session.NewNodeSession(addr, nodeType),
	}
}

// Call a remote command.
func (c *RemoteCmdClient) Call(ctx context.Context, command string, arguments []string) (cmdResult string, err error) {
	rcmd := &RemoteCommand{
		Command:   command,
		Arguments: arguments,
	}
	return rcmd.Call(ctx, c.session)
}

// RemoteCommand can be called concurrently by multiple sessions.
type RemoteCommand struct {
	Command   string
	Arguments []string
}

// Call a remote command to an existing session.
func (c *RemoteCommand) Call(ctx context.Context, session session.NodeSession) (cmdResult string, err error) {
	thriftArgs := &cmd.RemoteCmdServiceCallCommandArgs{
		Cmd: &cmd.Command{Cmd: c.Command, Arguments: c.Arguments},
	}
	res, err := session.CallWithGpid(ctx, &base.Gpid{}, thriftArgs, "RPC_CLI_CLI_CALL")
	if err != nil {
		return "", err
	}
	ret, _ := res.(*cmd.RemoteCmdServiceCallCommandResult)
	return ret.GetSuccess(), nil
}
