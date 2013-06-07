package main

import (
	"fmt"
	"net"
	"net/rpc"
	"runtime"
	"tserver/server/usermanager"
)

var (
	rpcSocket net.Listener
)

func main () {
	fmt.Println("Foo zeh Sörvah")
	var err error
	var conn net.Conn
	runtime.GOMAXPROCS(4)
	rpc.Register(usermanager.UserMan)
	if rpcSocket, err = net.Listen("tcp", ":6996"); err != nil {
		panic(err)
	}
	defer rpcSocket.Close()
	for {
		if conn, err = rpcSocket.Accept(); err != nil {
			panic(err)
		}
		go rpc.ServeConn(conn)
	}
}
