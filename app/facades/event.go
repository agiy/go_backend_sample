package facades

import (
	"github.com/goravel/framework/contracts/event"
)

// Event は登録したeventを発行する入口を返す。
func Event() event.Instance {
	return App().MakeEvent()
}
