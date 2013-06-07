package rpcservice

import (
	"tshared/netutil"
)

type RpcService struct {
}

func (this *RpcService) GlobalServerUrl (none *netutil.RpcNoneType, ptr *string) error {
	*ptr = "localhost:6996"
	return nil
}
