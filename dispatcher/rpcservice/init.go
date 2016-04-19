package rpcservice

import (
	"xim/utils/netutils"
	"xim/utils/rpcutils"
)

// StartRPCServer serve rpc server at netAddr.
func StartRPCServer(netAddr *netutils.NetAddr) {
	rpcutils.RegisterAndStartRPCServer(netAddr.Network, netAddr.LAddr,
		new(rpcutils.RPCServer),
		new(RPCDispatcher),
	)
}
