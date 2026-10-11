package facades

import (
	"github.com/goravel/framework/contracts/database/seeder"
)

// Seeder は初期dataを投入する入口を返す。
func Seeder() seeder.Facade {
	return App().MakeSeeder()
}
