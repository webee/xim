package rpcservice

import (
	"log"
	"time"
)

// RPCServer represents the rpc server.
type RPCServer struct {
}

// RPCServer methods.
const (
	RPCServerPing = "RPCServer.Ping"
	RPCServerTime = "RPCServer.Time"
)

// Ping tests if the rpc service is ok.
func (s *RPCServer) Ping(args *NoArgs, reply *NoReply) error {
	log.Println(RPCServerPing, "is called.")
	return nil
}

// Time returns current server local time.
func (s *RPCServer) Time(args *NoArgs, reply *RPCServerTimeReply) error {
	log.Println(RPCServerTime, "is called.")
	reply.T = time.Now()
	return nil
}
