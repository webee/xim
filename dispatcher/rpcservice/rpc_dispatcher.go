package rpcservice

import (
	"encoding/json"
	"log"
	"sync"
	"time"
	"xim/dispatcher/msgchan"
	"xim/logic"
)

// RPCDispatcher represents the rpc dispatcher.
type RPCDispatcher struct {
}

// NewRPCDispatcher returns a new rpc.
func NewRPCDispatcher() *RPCDispatcher {
	return &RPCDispatcher{}
}

// RPCDispatcherPutMsgArgs is the msg args.
type RPCDispatcherPutMsgArgs struct {
	User    logic.UserLocation
	Channel string
	Msg     json.RawMessage
}

// RPCDispatcherPutMsgReply is the msg reply.
type RPCDispatcherPutMsgReply struct {
	MsgID string
}

// RPCServer methods.
const (
	RPCDispatcherPutMsg = "RPCDispatcher.PutMsg"
)

var (
	rwmx     = new(sync.RWMutex)
	channels = make(map[string]*msgchan.MsgChan)
)

// PutMsg put msg to channel.
func (r *RPCDispatcher) PutMsg(args *RPCDispatcherPutMsgArgs, reply *RPCDispatcherPutMsgReply) error {
	var err error
	log.Println(RPCDispatcherPutMsg, "is called:", args.User, args.Channel, string(args.Msg))
	msgChan := getMsgChan(args.Channel)
	reply.MsgID, err = msgChan.Put(args.User, args.Msg)
	return err
}

func getMsgChan(channel string) (msgChan *msgchan.MsgChan) {
	rwmx.RLock()
	msgChan, ok := channels[channel]
	rwmx.RUnlock()
	if !ok || msgChan.Closed() {
		rwmx.Lock()
		if msgChan, ok = channels[channel]; !ok || msgChan.Closed() {
			msgChan = msgchan.NewMsgChan(channel, 10*time.Second)
			channels[channel] = msgChan
			msgChan.OnClose(func() {
				rwmx.Lock()
				defer rwmx.Unlock()
				delete(channels, channel)
				log.Printf("delete #%s from channels.", channel)
			})
		}
		rwmx.Unlock()
	}
	return
}
