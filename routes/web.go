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

	task := controllers.NewTaskController()
	facades.Route().Get("/api/tasks", task.Index)
	facades.Route().Post("/api/tasks", task.Store)
	facades.Route().Get("/api/tasks/{id}", task.Show)
	facades.Route().Put("/api/tasks/{id}", task.Update)
	facades.Route().Delete("/api/tasks/{id}", task.Destroy)
}
