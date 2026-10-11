// Package migrations はDB schemaの変更履歴を定義する。
package migrations

import (
	"fmt"

	"github.com/goravel/framework/contracts/database/schema"

	"github.com/agiy/go_backend_sample/app/facades"
)

// M20210101000001CreateJobsTable はdatabase queue用のtableを作成する。
type M20210101000001CreateJobsTable struct{}

// Signature はmigrationを識別する履歴名を返す。
func (r *M20210101000001CreateJobsTable) Signature() string {
	return "20210101000001_create_jobs_table"
}

// Up はqueueの保存先と失敗jobの一意制約を作成する。
func (r *M20210101000001CreateJobsTable) Up() error {
	if !facades.Schema().HasTable("jobs") {
		if err := facades.Schema().Create("jobs", func(table schema.Blueprint) {
			table.ID()
			table.String("queue")
			table.LongText("payload")
			table.UnsignedTinyInteger("attempts").Default(0)
			table.DateTimeTz("reserved_at").Nullable()
			table.DateTimeTz("available_at")
			table.DateTimeTz("created_at").UseCurrent()
			table.Index("queue")
		}); err != nil {
			return fmt.Errorf("jobs tableの作成: %w", err)
		}
	}

	if !facades.Schema().HasTable("failed_jobs") {
		if err := facades.Schema().Create("failed_jobs", func(table schema.Blueprint) {
			table.ID()
			table.String("uuid")
			table.Text("connection")
			table.Text("queue")
			table.LongText("payload")
			table.LongText("exception")
			table.DateTimeTz("failed_at").UseCurrent()
			table.Unique("uuid")
		}); err != nil {
			return fmt.Errorf("failed_jobs tableの作成: %w", err)
		}
	}

	return nil
}

// Down はこのmigrationで作成したtableを削除する。
func (r *M20210101000001CreateJobsTable) Down() error {
	if err := facades.Schema().DropIfExists("jobs"); err != nil {
		return fmt.Errorf("jobs tableの削除: %w", err)
	}

	if err := facades.Schema().DropIfExists("failed_jobs"); err != nil {
		return fmt.Errorf("failed_jobs tableの削除: %w", err)
	}

	return nil
}
