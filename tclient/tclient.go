package main

import (
	"fmt"
	"net/rpc"
	"runtime"
	"tshared/netutil"
)

func main () {
	runtime.GOMAXPROCS(4)
	var client, err = rpc.Dial("tcp", "localhost:9669")
	if err != nil {
		panic(err)
	}
	var globalServerAddr string
	if err = client.Call("RpcService.GlobalServerUrl", netutil.RpcNone, &globalServerAddr); err != nil {
		panic(err)
	}
	fmt.Printf("1: %+v\n", globalServerAddr)
}
