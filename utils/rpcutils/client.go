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
	netAddr   *netutils.NetAddr
	Client    *rpc.Client
	connected bool
	quit      chan bool
	retry     chan bool
}

// NewRPCClient creates a rpc client.
func NewRPCClient(netAddr *netutils.NetAddr, retry bool) (client *RPCClient, err error) {
	client = &RPCClient{
		netAddr: netAddr,
		quit:    make(chan bool, 1),
	}
	rpcClient, err := Connect(netAddr)
	if err == nil {
		client.Client = rpcClient
		client.connected = true
	}
	if retry {
		client.retry = make(chan bool, 1)
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
	} else {
		log.Printf("rpc %s connected.\n", netAddr)
	}
	return
}

// Retry trigger retry connect.
func (cli *RPCClient) Retry() {
	if cli.retry != nil {
		cli.retry <- true
		cli.retry <- false
	}
}

// RetryingReconnect retry reconnect rpc when crashed.
func (cli *RPCClient) RetryingReconnect() {
	netAddr := cli.netAddr
	for {
		if err := cli.Ping(); err != nil {
			cli.connected = false
			log.Printf("retry connecting %s.\n", netAddr)
			if client, err := Connect(netAddr); err == nil {
				cli.Client = client
				cli.connected = true
			}
		}
		waiting := true
		for waiting {
			select {
			case <-cli.quit:
				log.Printf("quit retry connecting %s.\n", netAddr)
				return
			case retry := <-cli.retry:
				waiting = !retry
				break
			case <-time.After(5 * time.Second):
				waiting = false
				break
			}
		}
	}
}

// Ping call service's ping method.
func (cli *RPCClient) Ping() error {
	if cli.Client == nil {
		return errors.New("rpc client not initialized")
	}
	call := <-cli.Client.Go("RPCServer.Ping", new(NoArgs), new(NoReply), nil).Done
	return call.Error
}

// Time get server's time.
func (cli *RPCClient) Time() (t time.Time, err error) {
	if cli.Client == nil {
		err = errors.New("rpc client not initialized")
		return
	}
	reply := new(RPCServerTimeReply)
	call := <-cli.Client.Go("RPCServer.Time", new(NoArgs), reply, nil).Done
	err = call.Error
	if err == nil {
		t = reply.T
	}
	return
}

// Connected returns whether client is connected.
func (cli *RPCClient) Connected() bool {
	return cli.connected
}

// Close close rpc client.
func (cli *RPCClient) Close() error {
	cli.quit <- true
	err := cli.Client.Close()
	return err
}
