package facades

import (
	"github.com/goravel/framework/contracts/queue"
)

// Queue はjobを実行・投入する入口を返す。
func Queue() queue.Queue {
	return App().MakeQueue()
}
