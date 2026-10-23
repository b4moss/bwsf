# 製品仕様（specs）

**現行バージョンに存在する**機能の仕様正本。`docs/tests/` と同じドメイン切り。  
**SemVer フォルダ（`vX.Y.Z/`）では切らない**（版フォルダは `plans/`。履歴は `_archived/history/`）。

| 置き場 | 役割 |
|--------|------|
| **`docs/specs/`（ここ）** | 現行機能の仕様正本（ドメイン別） |
| [`docs/tests/`](../tests/README.md) | テスト仕様書（TDD 用。実装前の Red 契約） |
| [`docs/plans/`](../plans/) | これからやる内容（未実装） |
| [`docs/site/`](../site/) | ユーザー向けガイド（公開サイト。OKF バンドル外のシェル） |

仕様とテストは分けて管理する。テスト仕様は本ディレクトリを参照して書いてよいが、決定の再解釈はここに戻す。

## 索引（ドメイン）

### config

| 文書 | 内容 | Issue |
|------|------|-------|
| [`config/host-resolve.md`](./config/host-resolve.md) | host 解決順（`--host` / プロジェクト / default） | #177 |
| [`config/global-v2.md`](./config/global-v2.md) | グローバル設定パス・スキーマ・`hosts[]`・マイグレーション | #177 |
| [`config/save-files.md`](./config/save-files.md) | `save_files` + `!` 否定、`not_save_files` 廃止 | #177 |

### cmd

| 文書 | 内容 | Issue |
|------|------|-------|
| [`cmd/setup.md`](./cmd/setup.md) | `bwsf setup`（host スキップ可・save_files 対話） | #177 |
| [`cmd/unlock-lock.md`](./cmd/unlock-lock.md) | `unlock` / `lock` / `lock --all` | #153 |
| [`cmd/auth.md`](./cmd/auth.md) | `auth login` / `logout` | #174 |
| [`cmd/init.md`](./cmd/init.md) | `bwsf init`（プロジェクト設定生成） | #193 |
| [`cmd/commands.md`](./cmd/commands.md) | コマンド一覧（実装棚卸し） | — |

### infra / core

| 文書 | 内容 | Issue |
|------|------|-------|
| [`infra/secretstore-hosts.md`](./infra/secretstore-hosts.md) | Keychain キー `hosts/<id>/…` | #153 |
| [`core/vault-unlock-restore.md`](./core/vault-unlock-restore.md) | vault 系の自動 restore | #153 |

## 履歴

旧 SemVer ファイル名の一括正本は [`../_archived/history/v0.20.0-multi-host.md`](../_archived/history/v0.20.0-multi-host.md) に退避。現行の編集は上表のドメインファイルへ。

----

以上
