// Package bootstrap は設定・provider・route・migrationを起動時に登録する。
package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"

	"github.com/agiy/go_backend_sample/config"
	"github.com/agiy/go_backend_sample/routes"
)

// Boot はGoravelを設定し、commandやHTTP処理に利用できる状態へ初期化する。
func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).
		WithRouting(func() {
			routes.Web()
			routes.Grpc()
		}).
		WithProviders(Providers).
		WithConfig(config.Boot).
		Create()
}
