# コマンド

## 概要

| コマンド | 説明 |
|---|---|
| `bwsf setup` | Bitwarden 接続の設定 |
| `bwsf push` | .env ファイルを Bitwarden にプッシュ |
| `bwsf pull` | .env ファイルを Bitwarden からプル |
| `bwsf list` | 保存されている全プロジェクトを一覧表示 |

## bwsf setup

Bitwarden 接続の設定を行います。

```bash
bwsf setup
```

任意: `dotenvs` 以外の Bitwarden フォルダ名を指定できます。

```bash
bwsf setup --folder my-envs
```

フォルダ名は `~/.config/bwsf/config.json` に保存され、push / pull / list で使われます。名前変更では既存ノートは自動移動されません。

この対話式コマンドでは以下の入力を求められます：
- **サーバー URL**: Bitwarden サーバー URL（Bitwarden Cloud の場合は空欄）
- **メールアドレス**: Bitwarden アカウントのメールアドレス
- **マスターパスワード**: Bitwarden のマスターパスワード

## bwsf push

現在のディレクトリから `.env` ファイルを Bitwarden 保管庫 にプッシュします。

```bash
cd /path/to/your_project
bwsf push
```

### オプション

| オプション | 説明 |
|---|---|
| `--from <dir>` | ソースディレクトリを指定（デフォルト: 現在のディレクトリ） |

### 動作

1. 現在のディレクトリ名をプロジェクト名として使用
2. ディレクトリ内の `.env*` ファイルを検索
3. 同名のプロジェクトが Bitwarden に存在する場合、上書きを確認
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

Bitwarden ボールトから `.env` ファイルを現在のディレクトリにプルします。

```bash
cd /path/to/your_project
bwsf pull
```

### オプション

| オプション | 説明 |
|---|---|
| `--output <dir>` | 出力ディレクトリを指定（デフォルト: 現在のディレクトリ） |

### 動作

1. 現在のディレクトリ名をプロジェクト名として使用
2. 設定フォルダ（デフォルト: `dotenvs`）内で一致するプロジェクトを検索
3. ローカルに `.env` ファイルが既に存在する場合、上書きを確認
4. `.env` ファイルをダウンロードして作成

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

### 出力例

```
Projects in Bitwarden:
  • my-web-app
  • api-server
  • mobile-app
```

## よくあるワークフロー

### 新規プロジェクトのセットアップ

```bash
# .env ファイルを作成
echo "API_KEY=secret123" > .env

# Bitwarden にプッシュ
bwsf push
```

### 新しいマシンでの同期

```bash
# プロジェクトをクローン
git clone https://github.com/yourorg/my-web-app.git
cd my-web-app

# Bitwarden から .env をプル
bwsf pull
```

### マルチ環境のセットアップ

```bash
# 複数の環境ファイルを作成
echo "API_URL=http://localhost:3000" > .env
echo "API_URL=https://staging.example.com" > .env.staging
echo "API_URL=https://api.example.com" > .env.production

# すべてのファイルをプッシュ
bwsf push

# 別のマシンで全ファイルをプル
bwsf pull
```



