// Package tests はGoravelを起動してHTTP・DB統合テストの共通環境を提供する。
package tests

import (
	"github.com/goravel/framework/testing"

	"github.com/agiy/go_backend_sample/bootstrap"
)

func init() {
	bootstrap.Boot()
}

// TestCase はGoravelのHTTPテスト支援を共有する。
type TestCase struct {
	testing.TestCase
}
