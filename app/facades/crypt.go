package facades

import (
	"github.com/goravel/framework/contracts/crypt"
)

// Crypt はAPP_KEYによる暗号化の入口を返す。
func Crypt() crypt.Crypt {
	return App().MakeCrypt()
}
