package main

import (
	"fmt"

	"github.com/garryfan2013/goget/proxy"
	_ "github.com/garryfan2013/goget/proxy/local"
	_ "github.com/garryfan2013/goget/proxy/rpc_proxy"
	"github.com/garryfan2013/goget/rest"
	"github.com/garryfan2013/goget/rpc"
)

func main() {
	pm, err := proxy.GetProxyManager(proxy.ProxyLocal)
	if err != nil {
		fmt.Println(err)
		return
	}

	rpcServer, err := rpc.NewRpcServer(pm)
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		rpcServer.Serve()
	}()

	pm, err = proxy.GetProxyManager(proxy.ProxyRPC)
	if err != nil {
		fmt.Println(err)
		return
	}

	restServer, err := rest.NewRestServer(pm)
	if err != nil {
		fmt.Println(err)
		return
	}

	restServer.Serve()
}
