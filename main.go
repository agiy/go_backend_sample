// Package main はGoravel applicationとArtisan commandの起点を提供する。
package main

import (
	"github.com/agiy/go_backend_sample/bootstrap"
)

func main() {
	app := bootstrap.Boot()

	app.Start()
}
