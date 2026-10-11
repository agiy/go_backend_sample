// Package controllers はHTTPの入力と応答を扱う。
package controllers

import (
	"context"
	"net/http"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"

	"github.com/agiy/go_backend_sample/app/facades"
)

// HealthController はプロセスの生存とDBへの接続可否を公開する。
type HealthController struct{}

// NewHealthController は起動確認用のcontrollerを作る。
func NewHealthController() *HealthController {
	return &HealthController{}
}

// Live はDB停止時にもプロセスの生存を確認できるようにする。
func (r *HealthController) Live(ctx contractshttp.Context) contractshttp.Response {
	return ctx.Response().Json(http.StatusOK, contractshttp.Json{"status": "ok"})
}

// Ready はDBへ接続できる場合だけリクエストを受けられると判断する。
func (r *HealthController) Ready(ctx contractshttp.Context) contractshttp.Response {
	// 接続済みpoolへのPingが通常リクエストを圧迫しないよう上限を設ける。
	pingContext, cancel := context.WithTimeout(ctx.Context(), time.Second)
	defer cancel()

	orm := facades.Orm()
	if orm == nil {
		return ctx.Response().Json(http.StatusServiceUnavailable, contractshttp.Json{"status": "unavailable"})
	}
	db, err := orm.DB()
	if err != nil {
		return ctx.Response().Json(http.StatusServiceUnavailable, contractshttp.Json{"status": "unavailable"})
	}
	if err := db.PingContext(pingContext); err != nil {
		return ctx.Response().Json(http.StatusServiceUnavailable, contractshttp.Json{"status": "unavailable"})
	}

	return ctx.Response().Json(http.StatusOK, contractshttp.Json{"status": "ok"})
}
