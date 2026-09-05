# コマンド

## 概要

現行の製品コマンド一覧（v0.20.0）です。コンパクトな棚卸しは [`docs/COMMANDS.md`](https://github.com/b4moss/bwsf/blob/main/docs/COMMANDS.md) にもあります。

| コマンド | 説明 | 主なフラグ |
|---|---|---|
| `bwsf setup` | ホストとグローバル `save_files` の設定（ログインなし。`auth login` を使用） | `--folder` `--host-type` `--url` `--email` `--skip-host` `--save-files` `--yes` |
| `bwsf auth login` | API Key 保存 → Identity 確認 → unlock | `--host` |
| `bwsf auth logout` | API Key と `vault_unlock` を削除 | `--host` `--all` |
| `bwsf unlock` | vault セッションを Unlock し `vault_unlock` を保存 | `--host` |
| `bwsf lock` | `vault_unlock` を削除（API Key は残す） | `--host` `--all` |
| `bwsf config show` | 現在のローカル設定を表示 | — |
| `bwsf push` | 管理対象ファイル（`.env*` / `*.tfvars` / `*.tfvars.json`）を Bitwarden にプッシュ | `--from` `--host` |
| `bwsf pull` | 管理対象ファイルを Bitwarden からプル | `--output` `--host` |
| `bwsf list` | 保存されている全プロジェクトを一覧表示 | `--host` |
| `bwsf clean` | リモートバックアップ確認後にローカルの管理対象ファイルを削除 | `--from` `--host` |

付帯（cobra 標準）: `bwsf -v` / `--version`、`bwsf help`、`bwsf completion`。

管理対象は、名前が `.env` で始まるファイル、または末尾が `.tfvars` / `.tfvars.json` のファイルです。名前に `.example` を含むものは除外されます。任意の `save_files` glob（`!` による除外）で、その後さらに絞り込みます。

bwsf は **API** のみを使用します。`bw` CLI バックエンドと `bwsf backend` は v0.20.0 で削除されました。

保管庫コマンドのホスト解決順: `--host` → プロジェクトの `host` → グローバルの `is_default`。

## bwsf setup

Bitwarden ホスト（スキップ可）と任意のグローバル `save_files` を設定します。

```bash
bwsf setup
```

設定は `~/.config/bwsf/config.jsonc` に書き込まれます。ホスト追加をスキップ（`hosts: []`）しても、`save_files` は設定できます。

任意: `dotenvs` 以外の Bitwarden フォルダ（`target_section`）を指定できます。

```bash
bwsf setup --folder my-envs
```

名前変更では既存ノートは自動移動されません。

`setup` はマスターパスワードでのログインを行いません。続けて `bwsf auth login` を実行してください。

### 非対話フラグ

| フラグ | 説明 |
|---|---|
| `--host-type` | `cloud` または `selfhosted`（`bitwarden-cloud` / `bitwarden-selfhost` に対応） |
| `--url` | セルフホストのサーバー URL（`--host-type=selfhosted` のとき必須） |
| `--email` | アカウントのメール |
| `--skip-host` | `hosts: []` のままにする |
| `--save-files` | グローバル `save_files` glob（`!` 接頭辞 = 除外） |
| `--yes` | 確認をすべて yes とみなす（フォルダ作成、レガシー移行など） |
| `--folder` | ホストの `target_section`（デフォルト: `dotenvs`） |

## bwsf init

**カレントディレクトリ**に `.bwsf/config.jsonc` を生成します（git root へは自動上昇しません）。

```bash
bwsf init
bwsf init --skip-host --yes
bwsf init --host default --save-files '.env*,!.env.local' --override-project-name my-api
```

グローバル設定が必要です（`bwsf setup`。`hosts: []` 可）。プロジェクトの `host` はグローバル `hosts[]` への id 参照のみ。既存ファイルがある場合は上書き確認（`--yes` でスキップ）。

### 非対話フラグ

| フラグ | 説明 |
|---|---|
| `--host <id>` | プロジェクト `host`（グローバルに存在する id） |
| `--skip-host` | プロジェクト `host` を書かない |
| `--save-files` | プロジェクト `save_files` glob（`!` = 除外） |
| `--override-project-name` | プロジェクト名の上書き（空ならキー省略） |
| `--yes` | 上書き確認をスキップ |


## bwsf auth

Personal API Key 認証を管理します。引数なしの `bwsf auth` はヘルプのみです。

```bash
bwsf auth login
bwsf auth login --host work
bwsf auth logout
bwsf auth logout --all
```

`auth login` は `client_id` / `client_secret` の入力（または保存済みキーの再利用）→ Identity 確認 → vault unlock（`vault_unlock` 保存）まで一気通貫です。キーはアカウント設定 → セキュリティ → キーで作成します。

`auth logout` は解決 host の API Key **と** `vault_unlock` を削除します。`bwsf lock` は vault セッションのみです。

## bwsf config show

`~/.config/bwsf/config.jsonc` にある現在のローカル設定を表示します。

```bash
bwsf config show
```

スキーマメタデータ、`save_files`、ホスト（id / type / url / email / target_section / default）を表示します。Bitwarden にはアクセスしません。設定ファイルが無い場合はエラー終了し、`bwsf setup` を案内します。

## bwsf push

現在のディレクトリから管理対象ファイルを Bitwarden 保管庫にプッシュします。

```bash
cd /path/to/your_project
bwsf push
bwsf push --host work
```

ホストの `target_section`（デフォルト: `dotenvs`）に同名プロジェクトの Note が既にある場合は、**上書き確認なしで更新**します。

## bwsf pull

```bash
cd /path/to/your_project
bwsf pull
bwsf pull --host work
```

## bwsf list

```bash
bwsf list
bwsf list --host work
```

## bwsf clean

```bash
cd /path/to/your_project
bwsf clean
bwsf clean --host work
```

リモートの Bitwarden バックアップを確認したうえで、ローカルの管理対象ファイルを削除します。
