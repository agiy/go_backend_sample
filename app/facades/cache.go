package facades

import (
	"github.com/goravel/framework/contracts/cache"
)

// Cache は設定済みのcache storeを返す。
func Cache() cache.Cache {
	return App().MakeCache()
}
