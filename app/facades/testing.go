package facades

import (
	"github.com/goravel/framework/contracts/testing"
)

// Testing はGoravelのtest支援機能を返す。
func Testing() testing.Testing {
	return App().MakeTesting()
}
