package rpcservice

import (
	"xim/utils/netutils"
	"xim/utils/rpcutils"
)

// StartRPCServer serve rpc server at netAddr.
func StartRPCServer(netAddr *netutils.NetAddr, rcvrs ...interface{}) {
	rpcutils.RegisterAndStartRPCServer(netAddr.Network, netAddr.LAddr,
		append([]interface{}{new(rpcutils.RPCServer)}, rcvrs...)...,
	)
}
