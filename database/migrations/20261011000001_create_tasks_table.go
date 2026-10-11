package migrations

import (
	"fmt"

	"github.com/agiy/go_backend_sample/app/facades"
)

// M20261011000001CreateTasksTable はTaskの保存先と入力制約を作成する。
type M20261011000001CreateTasksTable struct{}

// Signature はmigrationの履歴名を返す。
func (r *M20261011000001CreateTasksTable) Signature() string {
	return "20261011000001_create_tasks_table"
}

// Up はアプリを経由しない書き込みにも制約を適用する。
func (r *M20261011000001CreateTasksTable) Up() error {
	// PostgreSQLのCHECKを含むDDLを一文で適用し、失敗時に半端なtableを残さない。
	// btrimの文字集合はGoのstrings.TrimSpaceと同じUnicode White_Spaceを使う。
	if err := facades.Schema().Sql(`CREATE TABLE tasks (
		id BIGSERIAL PRIMARY KEY,
		title VARCHAR(200) NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		completed BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT tasks_title_not_blank CHECK (
			length(btrim(title, E' \t\n\r\v\f\u0085\u00a0\u1680\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200a\u2028\u2029\u202f\u205f\u3000')) > 0
		),
		CONSTRAINT tasks_description_length CHECK (char_length(description) <= 2000)
	)`); err != nil {
		return fmt.Errorf("tasks tableの作成: %w", err)
	}

	return nil
}

// Down はTaskのtableを削除する。
func (r *M20261011000001CreateTasksTable) Down() error {
	if err := facades.Schema().DropIfExists("tasks"); err != nil {
		return fmt.Errorf("tasks tableの削除: %w", err)
	}

	return nil
}
