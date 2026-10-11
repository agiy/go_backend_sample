package facades

import (
	"github.com/goravel/framework/contracts/route"
)

// Route はHTTP routeを登録する入口を返す。
func Route() route.Route {
	return App().MakeRoute()
}
