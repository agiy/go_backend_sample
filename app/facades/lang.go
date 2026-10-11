package facades

import (
	"context"

	"github.com/goravel/framework/contracts/translation"
)

// Lang は指定したcontextの翻訳を返す。
func Lang(ctx context.Context) translation.Translator {
	return App().MakeLang(ctx)
}
