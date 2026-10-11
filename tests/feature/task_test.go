package feature

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json/v2"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	testinghttp "github.com/goravel/framework/contracts/testing/http"
	"github.com/goravel/framework/support/carbon"
	"github.com/stretchr/testify/require"

	"github.com/agiy/go_backend_sample/app/facades"
	"github.com/agiy/go_backend_sample/app/http/resources"
	"github.com/agiy/go_backend_sample/app/models"
	"github.com/agiy/go_backend_sample/tests"
)

// APIのwire formatを独立した型で読み、Modelの内部実装に依存せず検証する。
type taskRecord struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type taskPageRecord struct {
	Data []taskRecord `json:"data"`
	Meta struct {
		Page    int   `json:"page"`
		PerPage int   `json:"per_page"`
		Total   int64 `json:"total"`
	} `json:"meta"`
}

func TestTaskResourceTimestamp(t *testing.T) {
	// OSの時計精度に依存せず、Linuxのナノ秒とPostgreSQLのマイクロ秒の差を再現する。
	createdAt := time.Date(2026, time.October, 11, 22, 0, 0, 123456789, time.FixedZone("JST", 9*60*60))
	updatedAt := time.Date(2026, time.October, 11, 22, 1, 0, 987654321, time.FixedZone("JST", 9*60*60))
	task := models.Task{
		CreatedAt: carbon.NewDateTime(carbon.New(createdAt)),
		UpdatedAt: carbon.NewDateTime(carbon.New(updatedAt)),
	}
	body, err := json.Marshal(resources.NewTaskResource(task))
	require.NoError(t, err)
	var actual struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}
	require.NoError(t, json.Unmarshal(body, &actual))
	require.Equal(t, "2026-10-11T13:00:00.123456Z", actual.CreatedAt)
	require.Equal(t, "2026-10-11T13:01:00.987654Z", actual.UpdatedAt)
}

func TestTaskCRUD(t *testing.T) {
	db := prepareTasks(t)
	response, err := taskHTTP(t).Post("/api/tasks?completed=true", strings.NewReader(`{"id":-1,"title":"　買い物　","created_at":"2000-01-01 00:00:00"}`))
	require.NoError(t, err)
	response.AssertCreated()
	created := decodeTask(t, response)
	require.Positive(t, created.ID)
	require.Equal(t, "買い物", created.Title)
	require.Empty(t, created.Description)
	require.False(t, created.Completed, "queryで省略したbody項目を上書きしません")
	require.NotEmpty(t, created.CreatedAt)
	require.NotEqual(t, "2000-01-01 00:00:00", created.CreatedAt)
	require.NotEmpty(t, created.UpdatedAt)
	path := fmt.Sprintf("/api/tasks/%d", created.ID)
	response.AssertHeader("Location", path)

	response, err = taskHTTP(t).Put(path, strings.NewReader(`{"title":"買い物完了","description":"牛乳を買う","completed":true}`))
	require.NoError(t, err)
	response.AssertOk()
	require.True(t, decodeTask(t, response).Completed)

	t.Run("falseと空文字への変更も永続化する", func(t *testing.T) {
		subResponse, subErr := taskHTTP(t).Put(path, strings.NewReader(`{"title":"明日の買い物","description":"","completed":false}`))
		require.NoError(t, subErr)
		subResponse.AssertOk()
		updated := decodeTask(t, subResponse)
		require.Equal(t, created.ID, updated.ID)
		require.Equal(t, created.CreatedAt, updated.CreatedAt)
		require.False(t, updated.Completed)
		require.Empty(t, updated.Description)

		// HTTPの応答とは別のDB接続経路で、commitされた値を読み戻す。
		var title, description string
		var completed bool
		require.NoError(t, db.QueryRowContext(t.Context(), "SELECT title, description, completed FROM tasks WHERE id = $1", created.ID).
			Scan(&title, &description, &completed))
		require.Equal(t, "明日の買い物", title)
		require.Empty(t, description)
		require.False(t, completed)
	})

	t.Run("PUTで省略した項目は既定値に戻す", func(t *testing.T) {
		subResponse, subErr := taskHTTP(t).Put(path, strings.NewReader(`{"title":"準備","description":"残してはいけない","completed":true}`))
		require.NoError(t, subErr)
		subResponse.AssertOk()
		subResponse, subErr = taskHTTP(t).Put(path, strings.NewReader(`{"title":"やり直し"}`))
		require.NoError(t, subErr)
		subResponse.AssertOk()
		subResponse, subErr = taskHTTP(t).Get(path)
		require.NoError(t, subErr)
		subResponse.AssertOk()
		actual := decodeTask(t, subResponse)
		require.Equal(t, "やり直し", actual.Title)
		require.Empty(t, actual.Description)
		require.False(t, actual.Completed)
	})

	t.Run("検証エラーは既存Taskを書き換えない", func(t *testing.T) {
		subResponse, subErr := taskHTTP(t).Put(path, strings.NewReader(`{"title":"　","completed":true}`))
		require.NoError(t, subErr)
		subResponse.AssertUnprocessableEntity()
		subResponse, subErr = taskHTTP(t).Get(path)
		require.NoError(t, subErr)
		subResponse.AssertOk()
		require.Equal(t, "やり直し", decodeTask(t, subResponse).Title)
	})

	response, err = taskHTTP(t).Delete(path, nil)
	require.NoError(t, err)
	response.AssertNoContent()
	var count int
	require.NoError(t, db.QueryRowContext(t.Context(), "SELECT count(*) FROM tasks WHERE id = $1", created.ID).Scan(&count))
	require.Zero(t, count, "soft deleteではなく行を削除します")
	response, err = taskHTTP(t).Get(path)
	require.NoError(t, err)
	response.AssertNotFound()
	response, err = taskHTTP(t).Put(path, strings.NewReader(`{"title":"再作成しない"}`))
	require.NoError(t, err)
	response.AssertNotFound()
	response, err = taskHTTP(t).Delete(path, nil)
	require.NoError(t, err)
	response.AssertNotFound()
}

func TestTaskValidation(t *testing.T) {
	db := prepareTasks(t)
	cases := []struct {
		name  string
		body  map[string]any
		field string
	}{
		{name: "タイトル必須", body: map[string]any{}, field: "title"},
		{name: "Unicode空白だけを拒否", body: map[string]any{"title": " \t\n\u0085\u00a0\u1680\u2003\u202f\u205f\u3000"}, field: "title"},
		{name: "タイトル201文字を拒否", body: map[string]any{"title": strings.Repeat("あ", 201)}, field: "title"},
		{name: "説明2001文字を拒否", body: map[string]any{"title": "作業", "description": strings.Repeat("あ", 2001)}, field: "description"},
		{name: "タイトルの数値を拒否", body: map[string]any{"title": 123}, field: "title"},
		{name: "説明のnullを拒否", body: map[string]any{"title": "作業", "description": nil}, field: "description"},
		{name: "文字列のfalseを拒否", body: map[string]any{"title": "作業", "completed": "false"}, field: "completed"},
		{name: "数値のbooleanを拒否", body: map[string]any{"title": "作業", "completed": 0}, field: "completed"},
		{name: "booleanのnullを拒否", body: map[string]any{"title": "作業", "completed": nil}, field: "completed"},
		{name: "PostgreSQLが保存できないNUL文字を拒否", body: map[string]any{"title": "作業", "description": "a\x00b"}, field: "description"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.body)
			require.NoError(t, err)
			response, err := taskHTTP(t).Post("/api/tasks", bytes.NewReader(body))
			require.NoError(t, err)
			response.AssertUnprocessableEntity()
			var actual struct {
				Errors map[string]map[string]string `json:"errors"`
			}
			decodeResponse(t, response, &actual)
			require.Contains(t, actual.Errors, test.field)
		})
	}
	t.Run("不正なJSONとobject以外を拒否", func(t *testing.T) {
		for _, body := range []string{`{"title":`, `[]`, `null`, `{"title":"a","title":"b"}`} {
			response, err := taskHTTP(t).Post("/api/tasks", strings.NewReader(body))
			require.NoError(t, err)
			response.AssertBadRequest()
		}
	})
	t.Run("queryのタイトルでbodyの欠落を補わない", func(t *testing.T) {
		response, err := taskHTTP(t).Post("/api/tasks?title=spoof", strings.NewReader(`{}`))
		require.NoError(t, err)
		response.AssertUnprocessableEntity()
	})
	t.Run("JSON以外の書き込みを拒否", func(t *testing.T) {
		response, err := taskHTTP(t).WithHeader("Content-Type", "text/plain").Post("/api/tasks", strings.NewReader(`{"title":"作業"}`))
		require.NoError(t, err)
		response.AssertUnsupportedMediaType()
	})
	var count int
	require.NoError(t, db.QueryRowContext(t.Context(), "SELECT count(*) FROM tasks").Scan(&count))
	require.Zero(t, count, "拒否した入力はDBへ書き込みません")

	t.Run("日本語の上限文字数は受け付ける", func(t *testing.T) {
		body, err := json.Marshal(map[string]any{"title": strings.Repeat("あ", 200), "description": strings.Repeat("い", 2000)})
		require.NoError(t, err)
		response, err := taskHTTP(t).Post("/api/tasks", bytes.NewReader(body))
		require.NoError(t, err)
		response.AssertCreated()
		created := decodeTask(t, response)
		require.Equal(t, strings.Repeat("あ", 200), created.Title)
		require.Equal(t, strings.Repeat("い", 2000), created.Description)
	})
}

func TestTaskPagination(t *testing.T) {
	prepareTasks(t)
	response, err := taskHTTP(t).Get("/api/tasks")
	require.NoError(t, err)
	response.AssertOk()
	var empty taskPageRecord
	decodeResponse(t, response, &empty)
	require.NotNil(t, empty.Data, "空の一覧もnullではなく配列で返します")
	require.Empty(t, empty.Data)
	require.Equal(t, 1, empty.Meta.Page)
	require.Equal(t, 20, empty.Meta.PerPage)
	require.Zero(t, empty.Meta.Total)
	var ids []int64
	for _, title := range []string{"最初", "次", "最後"} {
		body, marshalErr := json.Marshal(map[string]any{"title": title})
		require.NoError(t, marshalErr)
		createdResponse, createErr := taskHTTP(t).Post("/api/tasks", bytes.NewReader(body))
		require.NoError(t, createErr)
		createdResponse.AssertCreated()
		ids = append(ids, decodeTask(t, createdResponse).ID)
	}
	response, err = taskHTTP(t).Get("/api/tasks?page=2&per_page=2")
	require.NoError(t, err)
	response.AssertOk()
	var page taskPageRecord
	decodeResponse(t, response, &page)
	require.Equal(t, 2, page.Meta.Page)
	require.Equal(t, 2, page.Meta.PerPage)
	require.EqualValues(t, 3, page.Meta.Total)
	require.Len(t, page.Data, 1)
	require.Equal(t, ids[2], page.Data[0].ID)
	require.Equal(t, "最後", page.Data[0].Title)

	t.Run("上限や形式が不正なpaginationを拒否", func(t *testing.T) {
		for _, query := range []string{"page=0", "page=1000001", "page=x", "per_page=0", "per_page=101", "per_page=-1"} {
			response, err := taskHTTP(t).Get("/api/tasks?" + query)
			require.NoError(t, err)
			response.AssertBadRequest()
		}
	})
	t.Run("不正なIDを拒否", func(t *testing.T) {
		for _, id := range []string{"0", "-1", "abc", "9223372036854775808"} {
			response, err := taskHTTP(t).Get("/api/tasks/" + id)
			require.NoError(t, err)
			response.AssertBadRequest()
		}
	})
}

func TestTaskDatabaseConstraints(t *testing.T) {
	db := prepareTasks(t)
	cases := []struct {
		name  string
		query string
		args  []any
		state string
	}{
		{name: "タイトル欠落", query: "INSERT INTO tasks DEFAULT VALUES", state: "23502"},
		{name: "空白タイトル", query: "INSERT INTO tasks (title) VALUES ($1)", args: []any{" \t\n\u0085\u00a0\u1680\u2003\u202f\u205f\u3000"}, state: "23514"},
		{name: "タイトル文字数", query: "INSERT INTO tasks (title) VALUES ($1)", args: []any{strings.Repeat("あ", 201)}, state: "22001"},
		{name: "説明文字数", query: "INSERT INTO tasks (title, description) VALUES ('作業', $1)", args: []any{strings.Repeat("あ", 2001)}, state: "23514"},
		{name: "説明null", query: "INSERT INTO tasks (title, description) VALUES ('作業', NULL)", state: "23502"},
		{name: "完了状態null", query: "INSERT INTO tasks (title, completed) VALUES ('作業', NULL)", state: "23502"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			tx, err := db.BeginTx(t.Context(), nil)
			require.NoError(t, err)
			defer func() { require.NoError(t, tx.Rollback()) }()
			_, err = tx.ExecContext(t.Context(), test.query, test.args...)
			require.Error(t, err)
			state, ok := errors.AsType[interface {
				error
				SQLState() string
			}](err)
			require.True(t, ok)
			require.Equal(t, test.state, state.SQLState())
		})
	}
}

func prepareTasks(t *testing.T) *sql.DB {
	t.Helper()
	// table全体を掃除するため、テスト専用DBと環境を確認して直列で実行する。
	require.Equal(t, "testing", facades.Config().GetString("app.env"))
	require.Equal(t, "go_backend_sample_test", facades.Orm().DatabaseName())
	require.NoError(t, facades.Artisan().Call("migrate --no-ansi"))
	require.True(t, facades.Schema().HasTable("tasks"))
	db, err := facades.Orm().DB()
	require.NoError(t, err)
	_, err = db.ExecContext(t.Context(), "DELETE FROM tasks")
	require.NoError(t, err)
	t.Cleanup(func() {
		// t.ContextはCleanup前にcancelされるので、後始末には独立した期限を設定する。
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := db.ExecContext(ctx, "DELETE FROM tasks")
		require.NoError(t, err)
	})

	return db
}

func taskHTTP(t *testing.T) testinghttp.Request {
	t.Helper()
	return new(tests.TestCase).Http(t).WithContext(t.Context())
}

func decodeTask(t *testing.T, response testinghttp.Response) taskRecord {
	t.Helper()
	var task taskRecord
	decodeResponse(t, response, &task)
	var document map[string]any
	decodeResponse(t, response, &document)
	require.NotContains(t, document, "data", "単一のTaskは直接返します")
	require.Len(t, document, 6, "Resourceに指定した6項目だけを公開します")
	_, createdErr := time.Parse(time.RFC3339Nano, task.CreatedAt)
	require.NoError(t, createdErr, "ResourceはRFC 3339形式で日時を返します")
	require.True(t, strings.HasSuffix(task.CreatedAt, "Z"), "日時はUTCで返します")
	_, updatedErr := time.Parse(time.RFC3339Nano, task.UpdatedAt)
	require.NoError(t, updatedErr)
	return task
}

func decodeResponse(t *testing.T, response testinghttp.Response, target any) {
	t.Helper()
	content, err := response.Content()
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal([]byte(content), target))
}
