package facades

import (
	"github.com/goravel/framework/contracts/validation"
)

// Validation は入力を検証する入口を返す。
func Validation() validation.Validation {
	return App().MakeValidation()
}
