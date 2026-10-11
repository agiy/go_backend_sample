package facades

import (
	"github.com/goravel/framework/contracts/schedule"
)

// Schedule は定期処理を登録する入口を返す。
func Schedule() schedule.Schedule {
	return App().MakeSchedule()
}
