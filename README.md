# GoDB Admin

MySQL、PostgreSQL などに対応する、Web データベース管理ツールです

## 特徴

- 🗄️ 複数のデータベースタイプに対応（MySQL、PostgreSQL、MariaDB）
- 🌳 サーバー/データベース/テーブルのツリー構造ナビゲーション
- 📊 テーブルデータ・詳細の表示
- 📥 CSVエクスポート機能（複数テーブル対応）
- 👥 ユーザー権限管理（GRANT文表示）
- ℹ️ サーバー情報の表示（バージョン、文字セット など）
- 🔐 パスワードのAES-256-GCM暗号化
- 🌐 多言語対応（日本語・英語）
- 💾 設定の永続化 (JSON形式)

## 必要要件

- Go 1.24.4以上
- MySQL 5.7以上、MariaDB、またはPostgreSQL 9.6以上

## インストール

```bash
# リポジトリをクローン
git clone <repository-url>
cd godbadmin

# 依存関係のインストール
go mod download

# Makefileを使用してビルド
make build

# または直接実行
make run
```

## 使い方

### サーバー起動

```bash
# デフォルトポート8000で起動（使用中なら自動的に8001, 8002...を試行）
make run

# または直接実行
go run main.go

# ポート指定
go run main.go -port 9000
```

ブラウザで http://localhost:8000 にアクセス

### サーバー設定の追加

1. トップページで「サーバー追加」ボタンをクリック
2. 以下の情報を入力:
   - **名前**: 識別用の名前
   - **データベースタイプ**: MySQL、PostgreSQL、MariaDBから選択
     - 選択すると自動的にデフォルトポートが入力されます
   - **ホスト**: データベースサーバーのホスト (例: localhost)
   - **ポート**: データベースのポート番号
     - MySQL/MariaDB: 3306
     - PostgreSQL: 5432
   - **ユーザー**: データベースユーザー名
   - **パスワード**: データベースパスワード
   - **データベース**: 接続先のデータベース名
     - 「データベース取得」ボタンで利用可能なデータベース一覧を取得可能
     - または手動で入力
3. 「保存」をクリック

### データベース・テーブルの操作

1. サーバー一覧からサーバーを選択
2. 「📊 データベース管理」ボタンをクリック
3. 左側のツリーでデータベース・テーブルを選択
4. 以下の操作が可能:
   - **データベース作成**: メニューから「データベース作成」を選択
   - **テーブルデータ表示**: テーブルをクリック（最大100件表示）
   - **テーブル詳細**: 「📋 テーブル詳細」ボタンでカラム情報とCREATE TABLE文を表示
   - **行詳細**: データ行の🔍アイコンをクリックして詳細表示

### サーバー情報・権限管理

1. サーバー一覧からサーバーを選択
2. 以下のボタンから各種情報を表示:
   - **ℹ️ サーバー情報**: バージョン、プロトコル、文字セット、SSL状態
   - **👥 ユーザー権限**: 全ユーザーとGRANT文の表示

## 設定ファイル

サーバー設定は `settings.json` に保存されます。

```json
{
  "servers": [
    {
      "id": "uuid-here",
      "name": "ローカルDB",
      "db_type": "mysql",
      "host": "localhost",
      "port": 3306,
      "user": "root",
      "password": "password",
      "database": "mydb"
    },
    {
      "id": "uuid-here-2",
      "name": "PostgreSQL開発",
      "db_type": "postgresql",
      "host": "localhost",
      "port": 5432,
      "user": "postgres",
      "password": "password",
      "database": "testdb"
    }
  ]
}
```

## プロジェクト構成

```
godbadmin/
├── main.go                     # エントリーポイント
├── config/                     # 設定管理
│   ├── config.go              # サーバー設定の永続化
│   └── crypto.go              # パスワード暗号化
├── handlers/                   # HTTPハンドラー
│   ├── server.go              # サーバー管理、情報、権限
│   └── database.go            # データベース、テーブル、行操作、エクスポート
├── db/                         # データベース接続
│   └── db.go                  # データベース操作
├── i18n/                       # 多言語化
│   ├── i18n.go                # 多言語化の初期化と関数
│   └── locales/
│       ├── ja.json            # 日本語翻訳
│       └── en.json            # 英語翻訳
├── templates/                  # HTMLテンプレート
│   ├── header.html            # 共通ヘッダー
│   ├── styles.html            # 共通スタイル
│   ├── servers.html           # サーバー管理
│   ├── server_form.html       # サーバー追加・編集
│   ├── server_info.html       # サーバー情報
│   ├── user_privileges.html   # ユーザー権限
│   ├── database_overview.html # データベース概要
│   ├── table_data.html        # テーブルデータ
│   ├── table_details.html     # テーブル詳細
│   ├── row_details.html       # 行詳細
│   └── export.html            # エクスポート
├── settings.json               # サーバー設定 (自動生成、暗号化キー含む)
└── Makefile                    # ビルド・デプロイ
```

## 技術スタック

- [Echo](https://echo.labstack.com/) - 高性能Webフレームワーク
- [sqlx](https://github.com/jmoiron/sqlx) - SQLツールキット
- [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) - MySQLドライバー
- [go-i18n](https://github.com/nicksnyder/go-i18n) - 多言語化ライブラリ
- crypto/aes - AES-256-GCM暗号化

## 開発

### Makefileコマンド

```bash
# アプリケーションを実行
make run

# マルチプラットフォームビルド（dist/フォルダに出力）
make build

# ビルド成果物を削除
make clean

# ヘルプを表示
make help
```

### ビルド対象プラットフォーム

`make build`で以下のプラットフォーム向けにビルドされます:

- macOS (Intel): `godbadmin-darwin-amd64`
- macOS (Apple Silicon): `godbadmin-darwin-arm64`
- Linux (x64): `godbadmin-linux-amd64`
- Linux (ARM64): `godbadmin-linux-arm64`
- Windows (x64): `godbadmin-windows-amd64.exe`

### テスト

```bash
go test ./...
```

### フォーマット

```bash
go fmt ./...
```

## セキュリティ上の注意

⚠️ **この実装はローカル開発環境での使用を想定しています**

- パスワードはAES-256-GCMで暗号化されますが、暗号化キーも同じファイルに保存されます
- 認証機能はありません
- 本番環境での使用には以下の対策が必要です：
  - 暗号化キーの環境変数化または専用キー管理システムの使用
  - 基本認証またはセッション管理の実装
  - HTTPS通信の使用

## ライセンス

MIT

## 貢献

プルリクエストを歓迎します。大きな変更の場合は、まずissueを開いて変更内容を議論してください。

## 実装済み機能

### サーバー管理
- ✅ サーバー追加・編集・削除
- ✅ パスワードのAES-256-GCM暗号化
- ✅ 接続テスト機能
- ✅ サーバー情報表示（バージョン、文字セット、SSL等）
- ✅ ユーザー権限管理（全ユーザー、GRANT文表示）

### データベース・テーブル操作
- ✅ データベース作成
- ✅ テーブルデータ表示（最大100件）
- ✅ テーブル詳細（カラム情報、CREATE TABLE文）
- ✅ 行詳細表示（プライマリキーベース）
- ✅ ツリー構造ナビゲーション
- ✅ リサイズ可能な2ペイン構造

### エクスポート
- ✅ CSVエクスポート（複数テーブル対応）
- ✅ エクスポート設定（区切り文字、囲み文字、エンコーディング）

### 多言語化
- ✅ 日本語・英語対応
- ✅ 言語切り替え機能
- ✅ Cookie経由の言語設定保持

## 今後の予定

### 優先度: 高
- [ ] テーブルデータの編集・削除
- [ ] 新規行の追加
- [ ] SQLクエリエディタ
- [ ] ページネーション・検索
- [ ] 全テンプレートの多言語化

### 優先度: 中
- [ ] インデックス情報の表示
- [ ] データのインポート（CSV、SQL）
- [ ] JSONエクスポート対応
- [ ] SQLダンプエクスポート対応
- [ ] テーブル作成・編集・削除
- [ ] データベース削除
- [ ] ユーザー管理（作成・編集・削除）

### 優先度: 低
- [ ] アプリケーション認証機能
- [ ] PostgreSQL完全対応（現在はMySQLのみ実装）
- [ ] クエリ履歴・お気に入り
- [ ] ダークモード対応

## API エンドポイント

### サーバー管理
- `POST /api/test-connection` - データベース接続テスト
  - ボディ: `{"host": "localhost", "port": 3306, "user": "root", "password": "pass", "db_type": "mysql"}`
  - レスポンス: `{"success": true}` または `{"success": false, "error": "..."}`

### データベース操作
- `GET /api/databases` - データベース一覧を取得
  - パラメータ: `host`, `port`, `user`, `password`
  - レスポンス: `{"success": true, "databases": ["db1", "db2"]}`

- `POST /api/database/create` - データベースを作成
  - ボディ: `{"server_id": "uuid", "db_name": "dbname", "charset": "utf8mb4", "collation": "utf8mb4_unicode_ci"}`
  - レスポンス: `{"success": true}`

### ユーザー権限
- `GET /api/user-grants` - 特定ユーザーのGRANT文を取得
  - パラメータ: `server_id`, `user`, `host`
  - レスポンス: `{"success": true, "grants": ["GRANT ALL ...", ...]}`

### 多言語化
- `GET /api/set-language?lang=ja` - 言語設定を変更（ja または en）
  - Cookie `lang` に設定を保存（有効期限: 1年）
  - 元のページにリダイレクト

## 主要ルート

### サーバー管理
- `GET /servers` - サーバー一覧・管理画面
- `GET /servers/new` - サーバー追加フォーム
- `POST /servers` - サーバー作成
- `GET /servers/:id/edit` - サーバー編集フォーム
- `POST /servers/:id` - サーバー更新
- `POST /servers/:id/delete` - サーバー削除
- `GET /servers/:id/info` - サーバー情報表示
- `GET /servers/:id/privileges` - ユーザー権限表示

### データベース・テーブル
- `GET /servers/:id/database` - データベース概要（パラメータ `?db=dbname` でデータベース選択）
- `GET /servers/:id/db/:db/table/:table` - テーブルデータ表示
- `GET /servers/:id/db/:db/table/:table/details` - テーブル詳細
- `GET /servers/:id/db/:db/table/:table/row` - 行詳細（PKパラメータ付き）

### エクスポート
- `GET /servers/:id/db/:db/export` - エクスポートページ（パラメータ `?table=tablename` でテーブル事前選択）
- `POST /servers/:id/db/:db/export` - エクスポート実行
