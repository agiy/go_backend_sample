package facades

import (
	"github.com/goravel/framework/contracts/process"
)

// Process は外部processの実行を構成する入口を返す。
func Process() process.Process {
	return App().MakeProcess()
}
