# コマンド

## 概要

| コマンド | 説明 |
|---|---|
| `bwsf setup` | Bitwarden 接続の設定 |
| `bwsf config show` | 現在のローカル設定を表示 |
| `bwsf push` | 管理対象ファイル（`.env*` / `*.tfvars` / `*.tfvars.json`）を Bitwarden にプッシュ |
| `bwsf pull` | 管理対象ファイルを Bitwarden からプル |
| `bwsf list` | 保存されている全プロジェクトを一覧表示 |
| `bwsf clean` | リモートバックアップ確認後にローカルの管理対象ファイルを削除 |

管理対象は、名前が `.env` で始まるファイル、または末尾が `.tfvars` / `.tfvars.json` のファイルです。名前に `.example` を含むものは除外されます。

## bwsf setup

Bitwarden 接続の設定を行います。

```bash
bwsf setup
```

任意: `dotenvs` 以外の Bitwarden フォルダ名を指定できます。

```bash
bwsf setup --folder my-envs
```

フォルダ名は `~/.config/bwsf/config.json` に保存され、push / pull / list / clean で使われます。名前変更では既存ノートは自動移動されません。

### 対話入力

- **サーバー URL**: Bitwarden サーバー URL（Bitwarden Cloud の場合は空欄）
- **メールアドレス**: Bitwarden アカウントのメールアドレス
- **マスターパスワード**: Bitwarden のマスターパスワード

### 非対話フラグ

自動化（スモークテストなど）向け:

| フラグ | 説明 |
|---|---|
| `--host-type` | `cloud` または `selfhosted` |
| `--url` | セルフホストのサーバー URL（`--host-type=selfhosted` のとき必須） |
| `--email` | アカウントのメール |
| `--password` | マスターパスワード |
| `--yes` | 確認をすべて yes とみなす（フォルダ作成など） |
| `--folder` | Bitwarden フォルダ名（デフォルト: `dotenvs`） |

## bwsf config show

`~/.config/bwsf/config.json` にある現在のローカル設定を表示します。

```bash
bwsf config show
```

ホスト種別、セルフホスト URL、メール、実効フォルダ名を表示します。Bitwarden にはアクセスしません。設定ファイルが無い場合はエラー終了し、`bwsf setup` を案内します。

## bwsf push

現在のディレクトリから管理対象ファイルを Bitwarden 保管庫にプッシュします。

```bash
cd /path/to/your_project
bwsf push
```

### オプション

| オプション | 説明 |
|---|---|
| `--from <dir>` | 管理対象ファイルがあるディレクトリ（デフォルト: 現在のディレクトリ） |

### 動作

1. 現在のディレクトリ名をプロジェクト名として使用
2. 管理対象ファイルを検出（`.env*` / `*.tfvars` / `*.tfvars.json`。名前に `.example` を含むものは除外）
3. 同名プロジェクトの Note が既にある場合は **上書き確認なしで更新**
4. 設定フォルダ（デフォルト: `dotenvs`）にノートアイテムとして保存

### 使用例

```bash
# 現在のディレクトリからプッシュ
cd my-web-app
bwsf push

# 特定のディレクトリからプッシュ
bwsf push --from ./config
```

## bwsf pull

Bitwarden ボールトから管理対象ファイルを現在のディレクトリにプルします。

```bash
cd /path/to/your_project
bwsf pull
```

### オプション

| オプション | 説明 |
|---|---|
| `--output <dir>` | 出力ディレクトリ（デフォルト: 現在のディレクトリ） |

### 動作

1. 現在のディレクトリ名をプロジェクト名として使用
2. 設定フォルダ（デフォルト: `dotenvs`）内で一致するプロジェクトを検索
3. ローカルに同名ファイルが既にある場合、ファイル単位で上書きを確認
4. Note から管理対象ファイルを書き出す

### 使用例

```bash
# 現在のディレクトリにプル
cd my-web-app
bwsf pull

# 特定のディレクトリにプル
bwsf pull --output ./config
```

## bwsf list

Bitwarden ボールトに保存されている全プロジェクトを一覧表示します。

```bash
bwsf list
```

### 出力

標準出力にプロジェクト名を 1 行ずつ表示します。設定フォルダにアイテムが無い場合:

```
No items found in dotenvs folder
```

アイテムがある場合の例:

```
my-web-app
api-server
mobile-app
```

## bwsf clean

Bitwarden 側のバックアップを確認したうえで、ローカルの管理対象ファイルを削除します。

```bash
cd /path/to/your_project
bwsf clean
```

### オプション

| オプション | 説明 |
|---|---|
| `--from <dir>` | 削除対象の管理対象ファイルがあるディレクトリ（デフォルト: カレントディレクトリ） |

### 挙動

1. カレントディレクトリ名をプロジェクト名として使う
2. リモートに同名 Note アイテムが無い、または管理対象ファイルが空なら中止する
3. 内容が一致すれば確認なしでローカルを削除する
4. 1 ファイルでも差分があれば単一選択で分岐する（Abort / Overwrite remote then clean / Remove local）

## よくあるワークフロー

### 新規プロジェクトのセットアップ

```bash
# 管理対象ファイルを作成
echo "API_KEY=secret123" > .env
echo 'region = "ap-northeast-1"' > terraform.tfvars

# Bitwarden にプッシュ
bwsf push
```

### 新しいマシンでの同期

```bash
# プロジェクトをクローン
git clone https://github.com/yourorg/my-web-app.git
cd my-web-app

# Bitwarden から管理対象ファイルをプル
bwsf pull
```

### マルチ環境のセットアップ

```bash
# 複数の環境ファイルを作成
echo "API_URL=http://localhost:3000" > .env
echo "API_URL=https://staging.example.com" > .env.staging
echo "API_URL=https://api.example.com" > .env.production

# すべての管理対象ファイルをプッシュ
bwsf push

# 別のマシンで全ファイルをプル
bwsf pull
```
