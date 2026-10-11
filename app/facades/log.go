package facades

import (
	"github.com/goravel/framework/contracts/log"
)

// Log はGoravelで設定したloggerを返す。
func Log() log.Log {
	return App().MakeLog()
}
