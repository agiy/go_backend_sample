package facades

import (
	"github.com/goravel/framework/contracts/http"
)

// RateLimiter はHTTP requestのrate limitを評価する入口を返す。
func RateLimiter() http.RateLimiter {
	return App().MakeRateLimiter()
}
