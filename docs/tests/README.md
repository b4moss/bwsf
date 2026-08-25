# テスト仕様書（索引）

本ディレクトリは charter（TDD）に基づく **テスト仕様書** の置き場です。  
実装前にここを FIX し、その後 Red → Green → Refactor で進めます。

既存の横断メモ: [`../TEST.md`](../TEST.md)（リファクタ当時の一括仕様。新規は本ディレクトリを優先）。

## 構成

| パス | 対象 |
|------|------|
| [`core/`](./core/) | コア層（管理対象ファイル検出、push/pull/list/clean） |
| [`cmd/`](./cmd/) | CLI コマンド層（cobra 登録・ユーザー向け出力契約） |
| [`project/`](./project/) | プロジェクトルート／名前解決（`.git` 基準） |

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

合意正本: [#134](https://github.com/b4moss/bwsf/issues/134)（Q1〜Q3, Q7）。`override_project_name` 実装は [#133](https://github.com/b4moss/bwsf/issues/133) へ先送り（空スロットのみ）
