// Package facades はapplicationに登録されたGoravelのserviceを公開する。
package facades

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

// App はGoravelのapplication containerを返す。
func App() contractsfoundation.Application {
	return foundation.App
}
