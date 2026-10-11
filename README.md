# go_backend_sample

Goravelを使うWeb APIバックエンドのサンプルです。Task管理のCRUD APIを通してModel・Form Request・Resource・migrationの流れを確認できます。Makefileやコマンドを隠すwrapperは置かず、下記のコマンドを直接実行します。

## 採用構成

| 用途 | 採用・バージョン |
| --- | --- |
| Go | 1.27.2 |
| Framework | Goravel v1.18.0（安定版） |
| HTTP / PostgreSQL driver | goravel/gin・goravel/postgres v1.18.0 |
| DB | PostgreSQL 18.6。Docker imageのdigestまで固定 |
| DB操作・schema・入力検証 | Goravel Orm / DB / Schema / Seeder / Form Request |
| 認証の土台 | Goravel Session guard、file session。ログイン機能は未実装 |
| 開発サーバー | Air v1.67.4 |
| Formatter / Linter | golangci-lint v2.14.0、gofumpt + goimports |
| Git hook | Lefthook v2.2.1。コミット時にstageした差分を検査 |
| CI設定 / Dockerfile検査 | actionlint v1.7.12 / Hadolint v2.15.1 |
| テスト | Go testing、Goravel HTTP testing、実PostgreSQL |
| 依存の脆弱性検査 | govulncheck v1.8.0 |

アプリ依存はルートの`go.mod`、開発ツールは`tools/go.mod`へ分離しています。actionlintはYAML依存のAPI競合を避けるため`tools/actionlint/go.mod`へ分離しています。グローバルな`go install ...@latest`は不要です。各moduleの`go.sum`も管理します。Hadolintは公式releaseのbinaryをSHA-256で確認して`bin/`へ配置します。

## 初回セットアップ

Go 1.27.2と、起動済みのDocker（Docker Compose対応）が必要です。以下はリポジトリのルートで実行します。初回ビルドでは依存のダウンロードに時間がかかります。

```sh
go version
docker compose version
go mod download
go -C tools mod download
go -C tools/actionlint mod download
```

続いてHadolintを導入します。macOS / Linuxのarm64・x86_64に対応します。`curl`と`shasum`を使用し、バージョンと検証用SHA-256の正本は`tools/hadolint.version`・`tools/hadolint.sha256`です。

```sh
(
  set -eu
  case "$(uname -s)/$(uname -m)" in
    Darwin/arm64) hadolint_asset=hadolint-macos-arm64 ;;
    Darwin/x86_64) hadolint_asset=hadolint-macos-x86_64 ;;
    Linux/aarch64|Linux/arm64) hadolint_asset=hadolint-linux-arm64 ;;
    Linux/x86_64) hadolint_asset=hadolint-linux-x86_64 ;;
    *) echo "未対応のOS・architectureです" >&2; exit 1 ;;
  esac
  mkdir -p bin
  curl --fail --location --retry 3 "https://github.com/hadolint/hadolint/releases/download/v$(cat tools/hadolint.version)/$hadolint_asset" -o "bin/$hadolint_asset"
  awk -v asset="*$hadolint_asset" '$2 == asset' tools/hadolint.sha256 | (cd bin && shasum -a 256 -c -)
  chmod +x "bin/$hadolint_asset"
  mv "bin/$hadolint_asset" bin/hadolint
)
```

Git hookを登録し、DBとアプリを初期化します。cloneした開発者ごとにhook登録が必要です。

```sh
go tool -modfile=tools/go.mod lefthook install

# 初回だけ作成する。既存の.envと鍵を上書きしない。
[ -f .env ] || cp .env.example .env
[ -f .env.testing ] || cp .env.testing.example .env.testing

go run . artisan key:generate
go run . --env=.env.testing artisan key:generate

docker compose up -d --wait postgres
go run . artisan migrate
go run . artisan migrate:status
```

`key:generate`は初回に実行します。既存の鍵を再生成すると暗号化済みdataに影響するため、日々の起動では実行しません。`.env`と`.env.testing`はGit管理対象外です。exampleのDB認証情報はローカル開発専用です。

## 開発サーバー

```sh
go tool -modfile=tools/go.mod air -c .air.toml
```

Go・route・migration・設定・`.env`の変更で再起動します。自動再起動が不要なときは`go run .`で起動できます。停止はCtrl+Cです。

```sh
curl --fail http://localhost:3000/health/live
curl --fail http://localhost:3000/health/ready
go run . artisan route:list
```

`/health/live`はプロセスが動いていれば200、`/health/ready`はPostgreSQLへ接続できれば200、接続できなければ503を返します。JSONはそれぞれ`{"status":"ok"}`、`{"status":"unavailable"}`です。

## Task ModelとCRUD API

ローカルでコードと動作を見るためのTask管理APIです。認証・利用者ごとの認可は未実装で、Taskは全リクエストで共有します。

| HTTP | path | 動作・成功時のstatus |
| --- | --- | --- |
| POST | `/api/tasks` | 作成。201と`Location` header |
| GET | `/api/tasks` | 一覧。200、ID昇順、pagination付き |
| GET | `/api/tasks/{id}` | 1件取得。200 |
| PUT | `/api/tasks/{id}` | 全置換。200 |
| DELETE | `/api/tasks/{id}` | 物理削除。204、bodyなし |

入力はJSON bodyの`title`・`description`・`completed`だけを使います。タイトルは前後のUnicode空白を除去して1〜200文字、説明は最大2000文字です。文字数はbyte数ではなくUnicode code point数で数えます。説明の既定値は`""`、完了状態は`false`です。PUTでもタイトルが必須で、省略した説明・完了状態は既定値へ戻します。部分更新のPATCHは実装していません。

文字列・booleanの型は厳密に扱い、`null`、文字列の`"false"`、数値の0、PostgreSQLが保存できないNUL文字を拒否します。IDやtimestampなど追加のbody項目はModelへ反映せず、query parameterで書き込み項目を補うこともできません。

一覧は`page`（既定1、1〜1,000,000）・`per_page`（既定20、1〜100）で取得します。空の一覧は`data: []`です。JSON構文・ID・paginationの形式が不正なら400、対象がなければ404、Content-TypeがJSON以外なら415、入力検証で拒否した場合は422を返します。内部エラーは500と一般的なメッセージを返し、詳細をGoravelのログへ記録します。

Modelは`app/models/task.go`です。`orm.Model`がID・作成日時・更新日時を持ち、アプリの項目を追加します。

```go
type Task struct {
    orm.Model
    Title       string `gorm:"column:title"`
    Description string `gorm:"column:description"`
    Completed   bool   `gorm:"column:completed"`
}
```

リクエストは`app/http/requests/task_request.go`、レスポンスのpresentationは`app/http/resources/task_resource.go`へ分けています。採用中のGoravel v1.18.0にはLaravelの`JsonResource`相当の組み込み機能がないため、Goの型で`TaskResource`・`TaskCollection`を定義しています。Goravelの`ResourceController`はCRUDルートを登録するための機能です。

ResourceはModelを埋め込まず、`id`・`title`・`description`・`completed`・`created_at`・`updated_at`だけを公開します。日時はUTCのRFC 3339形式で、小数秒はPostgreSQLと同じマイクロ秒精度（最大6桁）へ揃えます。Modelの項目が増えてもAPIの公開項目は自動で増えません。

```go
// controllerではModelをレスポンス用の型へ変換して返す。
return ctx.Response().Json(http.StatusOK, resources.NewTaskResource(task))
```

単一のTaskは`{"id": 1, "title": "買い物", ...}`のようにResourceを直接返します。一覧は`{"data": [...], "meta": {"page": 1, "per_page": 20, "total": 1}}`です。`data`はGoravelの必須形式ではなく、paginationなど他の情報も返す場合に使う、このAPIの規約です。

DBの正本は`database/migrations/20261011000001_create_tasks_table.go`です。NOT NULL、タイトルのVARCHAR(200)と空白だけを拒否するCHECK、説明の2000文字CHECK、説明・完了状態のDEFAULTを明記しています。更新時はGoravelのtransactionと行ロックを使い、存在確認から保存までの間に並行削除されることを防ぎます。`Save`で空文字・`false`も保存し、削除は行そのものを削除します。

開発サーバーを起動して次のコマンドで確認できます。最初の応答の`id`を確認して`task_id`へ入れてください。

```sh
go run . artisan migrate

curl --fail-with-body -i http://localhost:3000/api/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"買い物","description":"牛乳を買う"}'

# 作成したTaskのIDを指定する。
task_id=1

curl --fail-with-body "http://localhost:3000/api/tasks?page=1&per_page=20"
curl --fail-with-body "http://localhost:3000/api/tasks/$task_id"

curl --fail-with-body -X PUT "http://localhost:3000/api/tasks/$task_id" \
  -H 'Content-Type: application/json' \
  -d '{"title":"買い物完了","description":"","completed":true}'

curl --fail-with-body -i -X DELETE "http://localhost:3000/api/tasks/$task_id"
```

## Formatter・Linter

```sh
# 整形を適用する。
go tool -modfile=tools/go.mod golangci-lint fmt

# 整形されていない差分を表示し、違反があればexit code 1で終了する。
go tool -modfile=tools/go.mod golangci-lint fmt --diff

# 全application codeとtestを検査する。
go tool -modfile=tools/go.mod golangci-lint config verify
go tool -modfile=tools/go.mod golangci-lint run
```

`.golangci.yml`を正本にして、error未確認・型assertion未確認・error wrapping不足・context漏れ・SQL resource漏れ・危険なコード・複雑度・公開APIの説明・未使用コードなどを検出します。問題数の表示上限を設けず、testも検査し、moduleの自動変更は許可しません。

49種類のlinterを有効にしています。gofumptの追加規則3種類（同じ型の引数をまとめる・裸のreturnを明示する・複数行の関数呼び出しの閉じ括弧を揃える）、testifylintの全checker、gocriticのdiagnostic・style・performanceを追加しています。gocriticのexperimental・opinionatedは対象外です。

DB accessはGoravelのfacadeへ統一します。`app/`・`database/`からGORM、pgx、sqlx、Bun、SQLiteへの直接依存を禁止しています。間接依存のGORM・pgxはGoravelの内部実装として利用します。テストもSQLiteへ置き換えません。

`nolint`には対象linter名と理由を必須にします。現在の例外は、公式generatorが参照する`facades.Http()`の公開名を維持するための`revive`指定のみです。テンプレートやtest全体を除外する設定はありません。

GitHub Actionsの`CI`は`fmt --diff`で整形違反を拒否し、全体のlint・vet・race detector付きtest・build・各moduleの`go mod tidy -diff`を実行します。Lefthook設定・GitHub Actions・Dockerfileも検査します。

## コミット時の差分チェック

```sh
# stageしたファイルを手動で検査する。
go tool -modfile=tools/go.mod lefthook run pre-commit

# hookの設定と登録状態を確認する。
go tool -modfile=tools/go.mod lefthook validate
go tool -modfile=tools/go.mod lefthook check-install

# CI設定・Dockerfileを全体検査する。
go tool -modfile=tools/actionlint/go.mod actionlint
bin/hadolint --config .hadolint.yaml Dockerfile
```

通常の`git commit`でも同じ検査が自動で動きます。`lefthook.yml`が実行コマンドの正本です。整形違反は差分を表示してコミットを止めるため、`golangci-lint fmt`で修正してから必要な変更をstageし直します。

Goの整形はstageしたファイル全体、lintの指摘はstageした差分に含まれるファイル全体を対象にします。型・参照の解析にはpackage全体を使い、Goファイルの削除でもlintを起動します。`.golangci.yml`・アプリのmodule・共通ツールのmoduleを変更した場合は、Goの整形とlintを全体に実行します。moduleファイルの変更時は3つのmoduleの整合性も確認します。

一部だけstageしたファイルはLefthookが未stage分を一時退避し、検査後に戻します。formatterの自動適用・自動stageは行いません。別ファイルの未stage変更はpackage解析に影響する場合があるため、型エラーもコミットを止めます。

コミット時は変更したworkflowをactionlintで検査し、Dockerfile・Hadolint設定の変更時はDockerfileを検査します。Hadolintはstyleまでの全severityを失敗扱いにし、inline ignoreを許可しません。hookはローカルで回避できるため、CIでも全体検査を続けます。

## テスト・build

DBを起動してから実行します。DBが起動していない場合は失敗し、DBテストをskipしません。

```sh
docker compose up -d --wait postgres
go test -race -count=1 ./... --env=.env.testing
go vet ./...
go build -trimpath -o bin/server .
go mod tidy -diff
go -C tools mod tidy -diff
go -C tools/actionlint mod tidy -diff
```

HTTP応答、migrationの適用→rollback→再適用、失敗jobのUUID一意制約、jobのNOT NULL制約、アプリ接続ユーザーの非superuser属性を実PostgreSQLで検証します。TaskのCRUDでは独立したSQL読み戻し、zero valueへの更新、物理削除、入力の型・Unicode文字数境界、Resourceの公開項目・日時形式、pagination、DBへの直接書き込みでも制約が働くことを確認します。

migrationやTaskのtestは`APP_ENV=testing`かつDB名が`go_backend_sample_test`の場合だけDBを変更します。Taskのtestは専用DB内のTaskを全件削除して前後を隔離するため、テストDBに保存したTaskは残りません。同じテストDBを使う複数のtest実行を同時に起動しないでください。

## DBの運用

| 項目 | ローカル設定 |
| --- | --- |
| 接続 | 127.0.0.1:15432（`DB_PORT`で変更可能） |
| 開発DB | go_backend_sample |
| テストDB | go_backend_sample_test |
| アプリ接続role | goravel。NOSUPERUSER / NOCREATEDB / NOCREATEROLE |
| 認証・文字コード・時刻 | SCRAM-SHA-256 / UTF-8 / UTC |
| query / lock / idle transaction上限 | 10秒 / 3秒 / 30秒 |
| connection pool | idle 5 / open 20。環境変数で調整可能 |

ローカルのroleはmigrationを実行するため各DBのownerです。本番のruntime roleとmigration roleの分離はデプロイ先が決まった際に設定します。`DB_SSLMODE=disable`はローカル専用で、設定コードの既定値は`verify-full`です。

DB初期化の正本は`docker/postgres/init.sql`です。初期化SQLは空のvolumeに対して一度だけ実行されます。既存volumeへ変更を反映する場合はmigrationなどを使います。

```sh
# dataを残して停止する。
docker compose stop postgres

# containerとnetworkを削除する。named volumeのdataは残る。
docker compose down

# schema変更を適用・確認する。
go run . artisan migrate
go run . artisan migrate:status
```

schema変更は`database/migrations/`へ追加し、`bootstrap/migrations.go`へ登録します。ORMの自動schema更新は使用しません。必要なNOT NULL・UNIQUE・外部キー・CHECK・indexはmigrationに明記します。

## 依存の脆弱性検査と既知の制約

```sh
go tool -modfile=tools/go.mod govulncheck ./...
```

GitHub Actionsの`Dependency audit`は手動実行でき、検査結果をそのまま失敗として表示します。

2026-10-11の検査では、Goravel v1.18.0が使うOpenTelemetry log v0.20.0系に次の2件が残ります。このコマンドは現状exit code 3になり、脆弱性検査の合格とは扱いません。lint・testの合格とは別に判断します。

- [GO-2026-6508](https://pkg.go.dev/vuln/GO-2026-6508): OTLP gRPC log exporterのTLS certificate設定。
- [GO-2026-6615](https://pkg.go.dev/vuln/GO-2026-6615): log BatchProcessorのbuffer飽和時のbusy loop。

修正されたv0.21.0はlog APIが非互換で、Goravel v1.18.0をそのままbuildできません。合意した暫定対応として`config/telemetry.go`でOpenTelemetryのログ送信を無効に固定しています。`OTEL_LOGS_EXPORTER`を指定しても有効になりません。通常のGoravelログは`config/logging.go`で利用できます。修正版に対応するGoravelが利用可能になったら、依存更新・build・test・脆弱性検査を通してから再有効化します。

pgx・x/net・x/crypto・x/text・gRPC・OpenTelemetry trace側など、互換性を保って修正できた間接依存は更新済みです。

module単位では[x/cryptoのGO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932)も検出されますが、今回の検査では対象packageのimport・対象symbolへの到達はありません。脆弱性検査の出力から除外していません。

## Docker image

```sh
docker build -t go_backend_sample:local .
```

GoとAlpineのimageをdigestまで固定したmulti-stage buildで、apk packageのバージョンも固定しています。実行ユーザーは非rootのUID / GID `10001:10001`です。`.env`、開発ツール、Git metadataはimageに含めません。起動時に必要な環境変数を渡します。デプロイやproduction環境の構築は今回の範囲に含みません。

## ディレクトリと機能追加

```text
app/facades/            Goravel serviceへの入口
app/http/controllers/  HTTP controller
app/http/requests/     入力検証・正規化
app/http/resources/    公開項目・レスポンス形式
app/models/            永続化するModel
bootstrap/             provider・route・migration登録
config/                各serviceの設定
routes/                公開route
database/migrations/  schema変更履歴
docker/postgres/      ローカルDB初期化
tests/feature/        HTTP・DB統合テスト
tools/                開発ツール専用module
```

Goravelのgeneratorも直接実行します。

```sh
go run . artisan list
go run . artisan make:controller --help
go run . artisan make:request --help
go run . artisan make:model --help
go run . artisan make:migration --help
go run . artisan make:seeder --help
```

ログイン・利用者ごとの認可・業務固有のtableは次の段階です。AI機能用のOpenAI driverは今回不要なため導入していません。

土台は[公式Goravel template](https://github.com/goravel/goravel/tree/v1.18.x)（commit `31651f646ccedf1701471bf01285111270a9d370`）です。元のMIT Licenseと著作権表記を`LICENSE`に保持しています。
