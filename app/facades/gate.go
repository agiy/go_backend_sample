package facades

import (
	"github.com/goravel/framework/contracts/auth/access"
	"github.com/goravel/framework/contracts/http"
)

// Gate は登録済みの認可policyを評価する入口を返す。
func Gate(_ ...http.Context) access.Gate {
	return App().MakeGate()
}
