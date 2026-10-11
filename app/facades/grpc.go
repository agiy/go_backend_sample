package facades

import (
	"github.com/goravel/framework/contracts/grpc"
)

// Grpc はgRPC serverを構成する入口を返す。
func Grpc() grpc.Grpc {
	return App().MakeGrpc()
}
