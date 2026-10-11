package facades

import (
	"github.com/goravel/framework/contracts/console"
)

// Artisan は登録済みのconsole commandを実行する入口を返す。
func Artisan() console.Artisan {
	return App().MakeArtisan()
}
