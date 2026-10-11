package facades

import (
	"github.com/goravel/framework/contracts/database/db"
)

// DB はGoravelのSQL query builderを返す。
func DB() db.DB {
	return App().MakeDB()
}
