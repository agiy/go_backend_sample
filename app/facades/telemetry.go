package facades

import (
	"github.com/goravel/framework/contracts/telemetry"
)

// Telemetry は計測・traceを設定する入口を返す。
func Telemetry() telemetry.Telemetry {
	return App().MakeTelemetry()
}
