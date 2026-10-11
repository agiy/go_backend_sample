package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/agiy/go_backend_sample/database/migrations"
)

// Migrations は実行対象のmigrationを順番に返す。
func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateJobsTable{},
	}
}
