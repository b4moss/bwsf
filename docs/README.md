# bwsf

プロダクトの意味的な pillar 正本（目的・スコープ・技術方針のハブ）。  
OKF の版索引は [`index.md`](./index.md)（`okf_version: "0.1"`）。詳細な振る舞いの正本は [`specs/`](./specs/)（`tests/` と同じドメイン切り）。

## 概要

Bitwarden を用いてプロジェクトの管理対象ファイル（`.env*` / Terraform `*.tfvars` / `*.tfvars.json`）を同期する CLI ツール。

- **バックエンド**: Bitwarden **API のみ**（Personal API Key）。`bw` CLI 経路と `bwsf backend` は廃止
- **設定**: グローバル `~/.config/bwsf/config.jsonc`、任意のプロジェクト `.bwsf/config.jsonc`
- **秘密保管**: OS の Keychain / secret service（host 単位の API Key と `vault_unlock`）

## 前提

- Bitwarden Cloud またはセルフホスト（Vaultwarden 等）の環境
- Bitwarden アカウントと Personal API Key（アカウント設定 → セキュリティ → キー）
- OS の秘密保管（macOS Keychain / Linux secret service）
- `bw` CLI は **不要**

## スコープ

### やること

- ホスト（接続先）と `save_files` の設定（`setup`）／プロジェクト設定の生成（`init`）
- Personal API Key の保存と vault セッション管理（`auth login` / `logout`、`unlock` / `lock`）
- 管理対象ファイルの push / pull / list / clean（Secure Note、フォルダは host の `target_section`、既定 `dotenvs`）
- マルチホスト（`--host` / プロジェクト `host` / `is_default`。`list` はプロジェクト `host` を見ない）

### やらぬこと

- Bitwarden 以外の秘密保管バックエンド
- `bw` CLI への依存
- 組織ボルト / SSO などの Bitwarden 組織機能一式

## プロジェクト名

- 既定: カレントから辿った **`.git` ルートのディレクトリ名**（無ければ cwd のベース名）
- 上書き: プロジェクト設定の `override_project_name`

## 管理対象ファイル

基盤ルール（通過後に `save_files` で絞り込み可）:

- 名前が `.env` で始まるもの
- 末尾が `.tfvars` / `.tfvars.json` のもの
- 名前に `.example` を含むものは除外

詳細: [`specs/config/save-files.md`](./specs/config/save-files.md)

## 主なコマンド

| コマンド | 役割 |
|----------|------|
| `bwsf setup` | ホスト / グローバル `save_files`（Login なし） |
| `bwsf auth login` / `logout` | API Key +（logout 時は）`vault_unlock` |
| `bwsf unlock` / `lock` | vault セッションのみ |
| `bwsf init` | `.bwsf/config.jsonc` 生成 |
| `bwsf config show` | ローカル設定表示 |
| `bwsf push` / `pull` / `list` / `clean` | 保管庫との同期・一覧・ローカル削除 |

棚卸し: [`specs/cmd/commands.md`](./specs/cmd/commands.md)

## 技術方針

- CLI: Go + cobra（実装は `app/src/`）
- 設定: JSONC 読み込み、書き込みは pretty JSON の `.jsonc`
- TDD / 薄い DDD: [`charter/`](./charter/)
- 公開ガイド: [`site/`](./site/)（Nuxt。OKF バンドル外のシェル）

## 索引

- [roadmap](./roadmap.md) — マイルストーン
- [specs](./specs/) — 現行仕様（ドメイン別）
- [plans](./plans/) — これからやる内容
- [tests](./tests/) — テスト仕様（specs と同じドメイン切り）
- [憲章](./charter/) — 開発ルール
- [OKF v0.1](./charter/okf/) — 知識バンドルの版定義
- [override-charter](./override-charter.md) — 憲章オーバーライド
- [site](./site/) — 公開ドキュメントサイト（シェル）

----

以上
