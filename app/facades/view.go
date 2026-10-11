package facades

import (
	"github.com/goravel/framework/contracts/view"
)

// View はtemplateを描画する入口を返す。
func View() view.View {
	return App().MakeView()
}
