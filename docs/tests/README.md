# テスト仕様書（索引）

本ディレクトリは charter（TDD）に基づく **テスト仕様書** の置き場です。  
実装前にここを FIX し、その後 Red → Green → Refactor で進めます。

既存の横断メモ: [`../TEST.md`](../TEST.md)（リファクタ当時の一括仕様。新規は本ディレクトリを優先）。

## 構成

| パス | 対象 |
|------|------|
| [`core/`](./core/) | コア層（管理対象ファイル検出、push/pull/list/clean） |

## Issue #102 / v0.15.0 — `*.tfvars` 対応

| 文書 | 内容 |
|------|------|
| [`core/managed_files_tfvars.md`](./core/managed_files_tfvars.md) | 管理対象ファイル検出（`.env*` + `.tfvars` / `.tfvars.json`、`.example` 除外） |
| [`core/ops_tfvars.md`](./core/ops_tfvars.md) | push / pull / list / clean での tfvars 往復・退行 |

合意正本: [#102](https://github.com/b4moss/bwsf/issues/102)（Q1〜Q3）
