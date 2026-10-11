// Package feature は実際のHTTPルートとPostgreSQLの接続・schema契約を検証する。
package feature

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/agiy/go_backend_sample/app/facades"
	"github.com/agiy/go_backend_sample/tests"
)

func TestHealth(t *testing.T) {
	testCase := new(tests.TestCase)
	t.Run("生存確認", func(t *testing.T) {
		response, err := testCase.Http(t).WithContext(t.Context()).Get("/health/live")
		require.NoError(t, err)
		response.AssertOk().AssertExactJson(map[string]any{"status": "ok"})
	})
	t.Run("PostgreSQL接続確認", func(t *testing.T) {
		response, err := testCase.Http(t).WithContext(t.Context()).Get("/health/ready")
		require.NoError(t, err)
		response.AssertOk().AssertExactJson(map[string]any{"status": "ok"})
	})
}

func TestPostgresMigrations(t *testing.T) {
	// schemaを変更するため、開発DBやproductionでは実行しない。共有DBを使うこのtestは直列で実行する。
	require.Equal(t, "testing", facades.Config().GetString("app.env"), ".env.testingを指定してください")
	require.Equal(t, "go_backend_sample_test", facades.Orm().DatabaseName(), "テスト専用DB以外のschema変更を拒否します")

	require.NoError(t, facades.Artisan().Call("migrate --no-ansi"))
	require.True(t, facades.Schema().HasTable("jobs"))
	require.True(t, facades.Schema().HasTable("failed_jobs"))
	require.True(t, facades.Schema().HasTable("tasks"))
	require.NoError(t, facades.Artisan().Call("migrate:reset --no-ansi"))
	require.False(t, facades.Schema().HasTable("jobs"))
	require.False(t, facades.Schema().HasTable("failed_jobs"))
	require.False(t, facades.Schema().HasTable("tasks"))
	require.NoError(t, facades.Artisan().Call("migrate --no-ansi"))
	require.True(t, facades.Schema().HasTable("jobs"))
	require.True(t, facades.Schema().HasTable("failed_jobs"))
	require.True(t, facades.Schema().HasTable("tasks"))

	db, err := facades.Orm().DB()
	require.NoError(t, err)
	var superuser bool
	require.NoError(t, db.QueryRowContext(t.Context(), "SELECT rolsuper FROM pg_roles WHERE rolname = current_user").Scan(&superuser))
	require.False(t, superuser, "アプリはsuperuserでDBに接続しません")

	t.Run("失敗jobのUUID重複をDBが拒否する", func(t *testing.T) {
		tx, err := db.BeginTx(t.Context(), nil)
		require.NoError(t, err)
		defer func() { require.NoError(t, tx.Rollback()) }()

		const insert = "INSERT INTO failed_jobs (uuid, connection, queue, payload, exception) VALUES ($1, 'postgres', 'default', '{}', 'test')"
		_, err = tx.ExecContext(t.Context(), insert, "duplicate-test")
		require.NoError(t, err)
		_, err = tx.ExecContext(t.Context(), insert, "duplicate-test")
		require.Error(t, err)
		state, ok := errors.AsType[interface {
			error
			SQLState() string
		}](err)
		require.True(t, ok)
		require.Equal(t, "23505", state.SQLState())
	})
	t.Run("jobのpayload欠落をDBが拒否する", func(t *testing.T) {
		tx, err := db.BeginTx(t.Context(), nil)
		require.NoError(t, err)
		defer func() { require.NoError(t, tx.Rollback()) }()

		_, err = tx.ExecContext(t.Context(), "INSERT INTO jobs (queue, available_at) VALUES ('default', CURRENT_TIMESTAMP)")
		require.Error(t, err)
		state, ok := errors.AsType[interface {
			error
			SQLState() string
		}](err)
		require.True(t, ok)
		require.Equal(t, "23502", state.SQLState())
	})
}
