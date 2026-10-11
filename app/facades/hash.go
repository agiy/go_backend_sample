package facades

import (
	"github.com/goravel/framework/contracts/hash"
)

// Hash は設定済みのpassword hashを返す。
func Hash() hash.Hash {
	return App().MakeHash()
}
