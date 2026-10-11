package facades

import (
	"github.com/goravel/framework/contracts/http/client"
)

// Http はHTTP clientのrequest builderを返す。
//
//nolint:revive // Goravelのgeneratorがfacades.Http()を参照するため公開名を維持する。
func Http() client.Request {
	return App().MakeHttp()
}
