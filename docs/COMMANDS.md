# bwsf コマンド一覧（v0.20.0）

実装（`app/src/cmd/`）を正とした現状コマンドの棚卸しです。  
詳細な説明・ワークフローはドキュメントサイトの [Commands](https://bwsf.oss.b4m.jp/en/guide/commands) / [コマンド](https://bwsf.oss.b4m.jp/ja/guide/commands) を参照してください。

製品正本（多ホスト）: [`specs/v0.20.0-multi-host.md`](./specs/v0.20.0-multi-host.md)

## 本体コマンド

| コマンド | 役割 | 主なフラグ |
|---|---|---|
| `bwsf setup` | ホスト / グローバル `save_files` 設定（Login なし。認証は `auth login`） | `--folder` `--host-type` `--url` `--email` `--skip-host` `--save-files` `--yes` |
| `bwsf auth` | サブコマンド親（引数なしはヘルプのみ） | （なし） |
| `bwsf auth login` | Personal API Key 保存 → Identity 確認 → unlock まで一気通貫 | `--host` |
| `bwsf auth logout` | API Key + `vault_unlock` を削除 | `--host` `--all` |
| `bwsf unlock` | 解決 host の vault セッションを Unlock し `vault_unlock` を Keychain に保存 | `--host` |
| `bwsf lock` | 解決 host の `vault_unlock` を削除（API Key は残す） | `--host` `--all` |
| `bwsf config show` | ローカル設定（`~/.config/bwsf/config.jsonc`）の表示 | （なし） |
| `bwsf push` | 管理対象ファイルを Bitwarden へプッシュ | `--from` `--host` |
| `bwsf pull` | Bitwarden から管理対象ファイルを取得 | `--output` `--host` |
| `bwsf list` | 設定フォルダ内のプロジェクト一覧 | `--host` |
| `bwsf clean` | リモートバックアップ確認後にローカル管理対象を削除 | `--from` `--host` |

`bwsf backend` は **廃止**（API のみ。実行するとエラー）。

### 管理対象ファイル

- 名前が `.env` で始まるもの
- 末尾が `.tfvars` / `.tfvars.json` のもの
- 名前に `.example` を含むものは除外
- 追加フィルタ: グローバル／プロジェクトの `save_files`（`!` 接頭辞で除外）。プロジェクト設定があれば完全オーバーライド

### ホスト解決

`--host` → プロジェクト `host` → グローバル `is_default`

## 付帯（cobra 標準）

| コマンド / フラグ | 役割 |
|---|---|
| `bwsf -v` / `--version` | バージョン表示 |
| `bwsf help` | ヘルプ |
| `bwsf completion` | シェル補完スクリプト生成 |
