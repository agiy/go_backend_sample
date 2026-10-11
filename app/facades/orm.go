package facades

import (
	"github.com/goravel/framework/contracts/database/orm"
)

// Orm はDB driverを隠蔽するGoravelのORMを返す。
func Orm() orm.Orm {
	return App().MakeOrm()
}
