# go_backend_sample

Goravelを使うWeb APIバックエンドの初期構成です。Makefileやコマンドを隠すwrapperは置かず、下記のコマンドを直接実行します。

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
| テスト | Go testing、Goravel HTTP testing、実PostgreSQL |
| 依存の脆弱性検査 | govulncheck v1.8.0 |

アプリ依存はルートの`go.mod`、開発ツールは`tools/go.mod`へ分離しています。グローバルな`go install ...@latest`は不要です。各moduleの`go.sum`も管理します。

## 初回セットアップ

Go 1.27.2と、起動済みのDocker（Docker Compose対応）が必要です。以下はリポジトリのルートで実行します。初回ビルドでは依存のダウンロードに時間がかかります。

```sh
go version
docker compose version
go mod download
go -C tools mod download

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

## Formatter・Linter

```sh
# 整形を適用する。
go tool -modfile=tools/go.mod golangci-lint fmt

# 整形されていない差分を表示する。
go tool -modfile=tools/go.mod golangci-lint fmt --diff

# 全application codeとtestを検査する。
go tool -modfile=tools/go.mod golangci-lint config verify
go tool -modfile=tools/go.mod golangci-lint run
```

`.golangci.yml`を正本にして、error未確認・型assertion未確認・error wrapping不足・context漏れ・SQL resource漏れ・危険なコード・複雑度・公開APIの説明・未使用コードなどを検出します。問題数の表示上限を設けず、testも検査し、moduleの自動変更は許可しません。

DB accessはGoravelのfacadeへ統一します。`app/`・`database/`からGORM、pgx、sqlx、Bun、SQLiteへの直接依存を禁止しています。間接依存のGORM・pgxはGoravelの内部実装として利用します。テストもSQLiteへ置き換えません。

`nolint`には対象linter名と理由を必須にします。現在の例外は、公式generatorが参照する`facades.Http()`の公開名を維持するための`revive`指定のみです。テンプレートやtest全体を除外する設定はありません。

GitHub Actionsの`CI`はformatter適用後に`git diff --exit-code`で差分を拒否し、lint・vet・race detector付きtest・build・`go mod tidy -diff`を実行します。

## テスト・build

DBを起動してから実行します。DBが起動していない場合は失敗し、DBテストをskipしません。

```sh
docker compose up -d --wait postgres
go test -race -count=1 ./... --env=.env.testing
go vet ./...
go build -trimpath -o bin/server .
go mod tidy -diff
go -C tools mod tidy -diff
```

HTTP応答、migrationの適用→rollback→再適用、失敗jobのUUID一意制約、jobのNOT NULL制約、アプリ接続ユーザーの非superuser属性を実PostgreSQLで検証します。migration testは`APP_ENV=testing`かつDB名が`go_backend_sample_test`の場合だけschemaを変更します。同じテストDBを使う複数のtest実行を同時に起動しないでください。

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

GoとAlpineのimageをdigestまで固定したmulti-stage buildで、実行ユーザーは非rootです。`.env`、開発ツール、Git metadataはimageに含めません。起動時に必要な環境変数を渡します。デプロイやproduction環境の構築は今回の範囲に含みません。

## ディレクトリと機能追加

```text
app/facades/            Goravel serviceへの入口
app/http/controllers/  HTTP controller
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

CRUD・ログイン・業務tableの追加は次の段階です。AI機能用のOpenAI driverは今回不要なため導入していません。

土台は[公式Goravel template](https://github.com/goravel/goravel/tree/v1.18.x)（commit `31651f646ccedf1701471bf01285111270a9d370`）です。元のMIT Licenseと著作権表記を`LICENSE`に保持しています。
