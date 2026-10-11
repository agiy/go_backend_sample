package facades

import (
	"github.com/goravel/framework/contracts/config"
)

// Config は環境変数と設定値を読み取る入口を返す。
func Config() config.Config {
	return App().MakeConfig()
}
