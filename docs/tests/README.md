# テスト仕様書（索引）

本ディレクトリは charter（TDD）に基づく **テスト仕様書** の置き場です。  
実装前にここを FIX し、その後 Red → Green → Refactor で進めます。

既存の横断メモ: [`../TEST.md`](../TEST.md)（リファクタ当時の一括仕様。新規は本ディレクトリを優先）。

## 構成

| パス | 対象 |
|------|------|
| [`core/`](./core/) | コア層（管理対象ファイル検出、push/pull/list/clean） |
| [`cmd/`](./cmd/) | CLI コマンド層（cobra 登録・ユーザー向け出力契約） |
| [`config/`](./config/) | 設定ファイル読み書き（グローバル config、JSONC 等） |
| [`project/`](./project/) | プロジェクトルート／名前解決（`.git` 基準） |
| [`utils/`](./utils/) | `bw` CLI ラッパの実行差し替え境界・単体ケース |

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

合意正本: [#133](https://github.com/b4moss/bwsf/issues/133)。キー命名は Issue コメントの決定。グローバル同系は [#177](https://github.com/b4moss/bwsf/issues/177)

## Issue #160 / v0.18.0 — coverage 75%+（Phase 2: `bw` 実行差し替え）

| 文書 | 内容 |
|------|------|
| [`utils/bw_exec_mock.md`](./utils/bw_exec_mock.md) | `runBw` 境界、ユニットで固定する分岐、e2e 委譲、モック非対象 |

合意正本: [#160](https://github.com/b4moss/bwsf/issues/160)。Phase 1（cmd DI・input 等）は既存契約の実装漏れ埋めのため本ディレクトリへの追加なし。
