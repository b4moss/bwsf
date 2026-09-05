# テスト仕様書（索引）

本ディレクトリは charter（TDD）に基づく **テスト仕様書** の置き場です。  
実装前にここを FIX し、その後 Red → Green → Refactor で進めます。

既存の横断メモ: [`../TEST.md`](../TEST.md)（リファクタ当時の一括仕様。新規は本ディレクトリを優先）。

## 構成

| パス | 対象 |
|------|------|
| [`infra/`](./infra/) | インフラ層（`ApiBwClient`、Identity、SecretStore、保管庫 CRUD 等） |
| [`core/`](./core/) | コア層（管理対象ファイル検出、再試行・push/pull/list/clean 接続等） |
| [`cmd/`](./cmd/) | CLI 層（cobra 登録・setup / auth / セッション寿命・ユーザー向け出力契約） |
| [`config/`](./config/) | 設定ファイル読み書き（グローバル config、JSONC 等） |
| [`project/`](./project/) | プロジェクトルート／名前解決（`.git` 基準） |
| [`utils/`](./utils/) | `bw` CLI ラッパの実行差し替え境界・単体ケース |

## Issue #53 Step 3

| 文書 | 内容 |
|------|------|
| [`infra/apiclient_unlock.md`](./infra/apiclient_unlock.md) | Unlock / ClearSession / 鍵状態 |
| [`core/unlock_retry_api.md`](./core/unlock_retry_api.md) | 認証切れと未 unlock の分岐再試行 |
| [`cmd/setup_api.md`](./cmd/setup_api.md) | `backend=api` 時の setup 分離 |
| [`cmd/session_lifecycle.md`](./cmd/session_lifecycle.md) | コマンド終了時のセッション破棄 |

Step 3 の実装計画正本: [Issue #53 Step 3 実装計画](https://github.com/b4moss/bwsf/issues/53#issuecomment-5276317436)

## Issue #53 Step 4

| 文書 | 内容 |
|------|------|
| [`infra/apiclient_vault.md`](./infra/apiclient_vault.md) | sync / folder / Secure Note CRUD（`ErrAPINotImplemented` 撤去） |
| [`cmd/setup_api_folder.md`](./cmd/setup_api_folder.md) | api setup からの設定フォルダ作成 |
| [`core/vault_ops_api.md`](./core/vault_ops_api.md) | push / pull / list を api adapter に接続したときの伝播・退行 |

Step 4 の実装計画正本: [Issue #53 Step 4 実装計画](https://github.com/b4moss/bwsf/issues/53#issuecomment-5277440346)

### Step 4 で扱わない（仕様書にも本実装テストを置かない）

- `clean` の api 対応
- 組織ボルト / SSO
- `BACKEND=api` の `make smoke` 本実装

## Issue #53 Step 5

仕上げ（default=`api`、docs、spike 削除、`bw` は残置）。**新規テスト仕様書は起こさない。**

実装計画正本: [Issue #53 Step 5 実装計画](https://github.com/b4moss/bwsf/issues/53#issuecomment-5278247151)

## Issue #102 / v0.15.0 — `*.tfvars` 対応

| 文書 | 内容 |
|------|------|
| [`core/managed_files_tfvars.md`](./core/managed_files_tfvars.md) | 管理対象ファイル検出（`.env*` + `.tfvars` / `.tfvars.json`、`.example` 除外） |
| [`core/ops_tfvars.md`](./core/ops_tfvars.md) | push / pull / list / clean での tfvars 往復・退行 |

合意正本: [#102](https://github.com/b4moss/bwsf/issues/102)（Q1〜Q3）

## Issue #125 / v0.15.0 — `bwsf config show`

| 文書 | 内容 |
|------|------|
| [`cmd/config_show.md`](./cmd/config_show.md) | `config` / `config show` の登録・表示項目・未設定時エラー |

方針正本: 同文書の前提 C1〜C6（[#125](https://github.com/b4moss/bwsf/issues/125)）

## Issue #134 / v0.16.0 — `.git` をプロジェクトルートとして扱う

| 文書 | 内容 |
|------|------|
| [`project/git_root.md`](./project/git_root.md) | `.git` 有無・サブディレクトリ・フラグ有無での Name/Dir/Warn 振る舞い（現行との差分） |

合意正本: [#134](https://github.com/b4moss/bwsf/issues/134)（Q1〜Q3, Q7）。`override_project_name` は [#133](https://github.com/b4moss/bwsf/issues/133) / [`config/project_local.md`](./config/project_local.md)

## Issue #155 / v0.18.0 — 設定ファイルの JSONC 対応（hujson）

| 文書 | 内容 |
|------|------|
| [`config/jsonc_load.md`](./config/jsonc_load.md) | `LoadConfig` の JSONC 読み込み、末尾カンマ、`SaveConfig` は strict JSON、Vault 対象外 |

合意正本: [#155](https://github.com/b4moss/bwsf/issues/155)。`.jsonc` 拡張子・`.bwsf/` は [#133](https://github.com/b4moss/bwsf/issues/133)

## Issue #133 / v0.18.0 — `.bwsf/config.(json|jsonc)` プロジェクト設定

| 文書 | 内容 |
|------|------|
| [`config/project_local.md`](./config/project_local.md) | 探索・0/1/複数選択、`override_project_name`、`save_files`/`not_save_files`、core フィルタ |

合意正本: [#133](https://github.com/b4moss/bwsf/issues/133)。グローバル同系（v0.20.0 多ホスト）および `save_files` / `!` は [#177](https://github.com/b4moss/bwsf/issues/177) / [`config/save_files_bang.md`](./config/save_files_bang.md) / 製品仕様 [`../specs/v0.20.0-multi-host.md`](../specs/v0.20.0-multi-host.md)

## Issue #160 / v0.18.0 — coverage 75%+（Phase 2: `bw` 実行差し替え）

| 文書 | 内容 |
|------|------|
| [`utils/bw_exec_mock.md`](./utils/bw_exec_mock.md) | `runBw` 境界、ユニットで固定する分岐、e2e 委譲、モック非対象 |

合意正本: [#160](https://github.com/b4moss/bwsf/issues/160)。Phase 1（cmd DI・input 等）は既存契約の実装漏れ埋めのため本ディレクトリへの追加なし。

## Issue #177 / v0.20.0 — グローバル設定 v2 / 多ホスト（§2）

製品正本: [`../specs/v0.20.0-multi-host.md`](../specs/v0.20.0-multi-host.md)。実装順は §2（本 Issue）→ #153 → #174 → #193。

| 文書 | 内容 |
|------|------|
| [`config/global_v2.md`](./config/global_v2.md) | 新スキーマ I/O（`.json` XOR `.jsonc`）、hosts 検証、Save、`config show` v2 |
| [`config/migrate_v2.md`](./config/migrate_v2.md) | 旧 flat 検出・確認／`--yes`・バックアップ・§2.6 写像（Keychain は #153） |
| [`config/save_files_bang.md`](./config/save_files_bang.md) | `save_files` + `!` 否定、`not_save_files` 廃止、プロジェクト完全オーバーライド |
| [`config/host_resolve.md`](./config/host_resolve.md) | §1.1 解決順、`push`/`pull`/`list`/`clean` の `--host` |
| [`cmd/setup_v2.md`](./cmd/setup_v2.md) | setup の host スキップ／既存 host 操作／`save_files` 対話、bw setup 廃止 |

## Issue #193 / v0.20.0 — `bwsf init`（§5）

製品正本: [`../specs/v0.20.0-multi-host.md`](../specs/v0.20.0-multi-host.md) §5。前提は #177（グローバル設定ファイル。`hosts: []` 可）。

| 文書 | 内容 |
|------|------|
| [`cmd/init.md`](./cmd/init.md) | `bwsf init` 対話、`.bwsf/config.jsonc` 生成、host / save_files / override、上書きと `--yes` |

### 旧テスト仕様との関係（v0.20 実装時）

| 旧文書 | 扱い |
|--------|------|
| [`config/jsonc_load.md`](./config/jsonc_load.md) | 読み込み技術は維持。パス・Save 先・スキーマは **global_v2 が優先** |
| [`config/project_local.md`](./config/project_local.md) | 探索・override・候補選択は維持。**フィルタ／`not_save_files` は save_files_bang が優先** |
| [`cmd/setup_api.md`](./cmd/setup_api.md) / [`cmd/setup_api_folder.md`](./cmd/setup_api_folder.md) | api・Login 無しの精神は維持。**フロー全体は setup_v2 が優先**（bw 退行テストは廃止） |
| [`cmd/config_show.md`](./cmd/config_show.md) | 登録形は維持。**表示フィールドは global_v2 §5 が優先** |
