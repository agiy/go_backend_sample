// Package models はGoravel ORMで永続化するModelを定義する。
package models

import (
	"github.com/goravel/framework/database/orm"
)

// Task はタイトル・説明・完了状態を持つ作業項目を表す。
// schemaの制約はmigrationで管理し、SoftDeletesを埋め込まず物理削除する。
type Task struct {
	orm.Model
	Title       string `gorm:"column:title"`
	Description string `gorm:"column:description"`
	Completed   bool   `gorm:"column:completed"`
}
