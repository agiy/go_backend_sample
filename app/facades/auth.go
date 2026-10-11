package facades

import (
	"github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/http"
)

// Auth は指定したHTTP contextの認証guardを返す。
func Auth(ctx ...http.Context) auth.Auth {
	return App().MakeAuth(ctx...)
}
