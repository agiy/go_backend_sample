package routes

import (
	"github.com/agiy/go_backend_sample/app/facades"
	"github.com/agiy/go_backend_sample/app/http/controllers"
)

// Web はHTTPルートを登録する。
func Web() {
	health := controllers.NewHealthController()
	facades.Route().Get("/health/live", health.Live)
	facades.Route().Get("/health/ready", health.Ready)
}
