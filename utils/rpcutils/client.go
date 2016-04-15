package rpcutils

import (
	"errors"
	"log"
	"net/rpc"
	"time"

	"xim/utils/netutils"
)

// RPCClient represents rpc client.
type RPCClient struct {
	netAddr *netutils.NetAddr
	client  *rpc.Client
	quit    chan bool
}

// NewRPCClient creates a rpc client.
func NewRPCClient(netAddr *netutils.NetAddr, retry bool) (client *RPCClient, err error) {
	client = &RPCClient{
		netAddr: netAddr,
		quit:    make(chan bool, 1),
	}
	rpcClient, err := Connect(netAddr)
	client.client = rpcClient
	if retry {
		go client.RetryingReconnect()
		return client, nil
	}

	return client, err
}

// Connect dial a rpc service.
func Connect(netAddr *netutils.NetAddr) (client *rpc.Client, err error) {
	client, err = rpc.Dial(netAddr.Network, netAddr.LAddr)
	if err != nil {
		log.Printf("rpc.Dial(%s) error: %s.\n", netAddr, err)
	}
	log.Printf("rpc %s connected.\n", netAddr)
	return
}

// RetryingReconnect retry reconnect rpc when crashed.
func (cli *RPCClient) RetryingReconnect() {
	netAddr := cli.netAddr
	for {
		if err := cli.Ping(); err != nil {
			log.Printf("retry connecting %s.\n", netAddr)
			if client, err := Connect(netAddr); err == nil {
				cli.client = client
			}
		}
		select {
		case <-cli.quit:
			log.Printf("quit retry connecting %s.\n", netAddr)
			return
		case <-time.After(2 * time.Second):
		}
	}
}

// Ping call service's ping method.
func (cli *RPCClient) Ping() error {
	if cli.client == nil {
		return errors.New("rpc client not initialized")
	}
	call := <-cli.client.Go("RPCServer.Ping", new(NoArgs), new(NoReply), nil).Done
	return call.Error
}

// Time get server's time.
func (cli *RPCClient) Time() (t time.Time, err error) {
	if cli.client == nil {
		err = errors.New("rpc client not initialized")
		return
	}
	reply := new(RPCServerTimeReply)
	call := <-cli.client.Go("RPCServer.Time", new(NoArgs), reply, nil).Done
	err = call.Error
	if err == nil {
		t = reply.T
	}
	return
}

// Close close rpc client.
func (cli *RPCClient) Close() error {
	err := cli.client.Close()
	cli.quit <- true
	return err
}
