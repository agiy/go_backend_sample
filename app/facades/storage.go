package facades

import (
	"github.com/goravel/framework/contracts/filesystem"
)

// Storage は設定済みのfilesystemを返す。
func Storage() filesystem.Storage {
	return App().MakeStorage()
}
