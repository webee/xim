package dispatcher

import (
	"encoding/json"
	"xim/logic"
	"xim/utils/netutils"
	"xim/utils/rpcutils"
)

// Dispatcher represents a dispatcher.
type Dispatcher struct {
	dispatcherRPCClient *rpcutils.RPCClient
}

var (
	dispatcher          *Dispatcher
	dispatcherRPCClient *rpcutils.RPCClient
)

// InitDispatcherRPC connect to dispatcher rpc.
func InitDispatcherRPC(netAddr *netutils.NetAddr) {
	dispatcherRPCClient, _ = rpcutils.NewRPCClient(netAddr, true)
	dispatcher = &Dispatcher{
		dispatcherRPCClient,
	}
}

// PutMsg push a msg to channel.
func PutMsg(from *logic.UserLocation, channel string, msg json.RawMessage) (msgID string, err error) {
	return
}
