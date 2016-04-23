package broker

import (
	"log"
	"xim/utils/netutils"
	"xim/utils/poolutils"
	"xim/utils/rpcutils"
)

type rpcClientPool struct {
	poolutils.ConcurResPool
}

type rpcClientResouce struct {
	rpcClient *rpcutils.RPCClient
}

func (c *rpcClientResouce) Close() {
	_ = c.rpcClient.Close()
}

func newRPCClientResouce(netAddr *netutils.NetAddr) (poolutils.Resource, error) {
	rpcClient, err := rpcutils.NewRPCClient(netAddr, false)
	return &rpcClientResouce{
		rpcClient: rpcClient,
	}, err
}

func newRPCClientPool(netAddr *netutils.NetAddr, concurrency uint) *rpcClientPool {
	return &rpcClientPool{
		ConcurResPool: *poolutils.NewConcurObjPool(
			concurrency,
			func() (poolutils.Resource, error) {
				return newRPCClientResouce(netAddr)
			}),
	}
}

func (p *rpcClientPool) Get() (int, *rpcutils.RPCClient, error) {
	id, res, err := p.ConcurResPool.Get()
	log.Println("get rpc pool:", id)
	rpcClientRes := res.(*rpcClientResouce)
	return id, rpcClientRes.rpcClient, err
}

func (p *rpcClientPool) Put(id int) {
	log.Println("put rpc pool:", id)
	p.ConcurResPool.Put(id)
}
