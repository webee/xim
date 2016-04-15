package rpcutils

import (
	"time"
)

// NoArgs used by rpc with no args.
type NoArgs struct {
}

// NoReply used by rpc with no reply.
type NoReply struct {
}

// RPCServerTimeReply is the RPCServer.Time reply type.
type RPCServerTimeReply struct {
	T time.Time
}
