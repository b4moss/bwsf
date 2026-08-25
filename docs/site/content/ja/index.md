---
title: ホーム
description: Bitwarden CLI で .env* と Terraform tfvars を管理
---

# bwsf

安全なファイル同期 — Bitwarden CLI で `.env*` と Terraform tfvars を管理します。

- [はじめる](/ja/guide/getting-started)
- [GitHub で見る](https://github.com/b4moss/bwsf)

## 主な機能

- **安全なストレージ** — 管理対象ファイル（`.env*` / `*.tfvars` / `*.tfvars.json`）を Bitwarden のボールトに安全に保存します。
- **簡単同期** — シンプルなコマンドで、ローカルと Bitwarden 間で管理対象ファイルをプッシュ・プル。
- **マルチ環境** — 1つのプロジェクトで複数ファイル（`.env`、`.env.staging`、`terraform.tfvars` など）を管理。
- **クロスプラットフォーム** — macOS と Linux に対応。Windows サポートは計画中です。

## クイックスタート

```bash
# Homebrew でインストール
brew tap b4m-oss/tap && brew install bwsf

# 初期設定
bwsf setup

# Bitwarden から管理対象ファイルをプル
cd /path/to/your_project
bwsf pull

# Bitwarden に管理対象ファイルをプッシュ
bwsf push
```

## 仕組み

bwsf は公式の Bitwarden CLI（`bw`）を使用して、管理対象ファイルを安全に保存・取得します。内容は Bitwarden フォルダ（デフォルト名: `dotenvs`、setup で変更可）内の**ノートアイテム**として保存されます。

各プロジェクトのファイルはカレントディレクトリ名で識別されるため、複数のプロジェクトを簡単に整理・管理できます。
