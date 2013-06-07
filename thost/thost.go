package main

import (
	"log"
	"net"
	"net/rpc"
	"os"
	"runtime"
	"thost/rpcservice"
)

var (
	rpcService = &rpcservice.RpcService {}
	rpcListener net.Listener
	serverConn *rpc.Client
	errLoggedServerConn = false
)

func logf (msg string, args ... interface{}) {
	log.Printf(msg + "\n", args...)
}

func ServerConn () *rpc.Client {
	var serverAddr string
	var err error
	if serverConn == nil {
		rpcService.GlobalServerUrl(nil, &serverAddr)
		if serverConn, err = rpc.Dial("tcp", serverAddr); err != nil {
			serverConn = nil
			if !errLoggedServerConn {
				errLoggedServerConn = true
				logf("Failed to connect to %s, will keep retrying. Error: %+v", serverAddr, err)
			}
		} else {
			errLoggedServerConn = false
			logf("Connected to %s.", serverAddr)
		}
	}
	return serverConn
}

func main () {
	var err error
	var clientConn net.Conn
	var errLoggedListen, errLoggedClientConn = false, false
	runtime.GOMAXPROCS(4)
	if err = rpc.Register(rpcService); err != nil {
		os.Exit(1)
	}
	for {
		if rpcListener == nil {
			if rpcListener, err = net.Listen("tcp", ":9669"); err != nil {
				rpcListener = nil
				if !errLoggedListen {
					errLoggedListen = true
					logf("Failed to establish TCP listener, will keep retrying. Error: %+v", err)
				}
			} else {
				errLoggedListen = false
				logf("TCP listener established.")
			}
		}
		if rpcListener != nil {
			if clientConn, err = rpcListener.Accept(); err == nil {
				errLoggedClientConn = false
				go rpc.ServeConn(clientConn)
			} else if !errLoggedClientConn {
				errLoggedClientConn = true
				rpcListener.Close()
				rpcListener = nil
				logf("Failed to accept TCP connection, will re-establish TCP listener. Error: %+v", err)
			}
		}
	}
}
