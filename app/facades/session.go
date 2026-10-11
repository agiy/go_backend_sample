package facades

import (
	"github.com/goravel/framework/contracts/session"
)

// Session はHTTP sessionを管理する入口を返す。
func Session() session.Manager {
	return App().MakeSession()
}
