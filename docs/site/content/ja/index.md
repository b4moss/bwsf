---
title: ホーム
description: Bitwarden API で .env* と Terraform tfvars を管理
---

# bwsf

安全なファイル同期 — Bitwarden **API**（Personal API Key）で `.env*` と Terraform tfvars を管理します。

- [はじめる](/ja/guide/getting-started)
- [GitHub で見る](https://github.com/b4moss/bwsf)
- [OKF docs ハブ](https://github.com/b4moss/bwsf/blob/main/docs/index.md) · [pillar](https://github.com/b4moss/bwsf/blob/main/docs/README.md) · [specs](https://github.com/b4moss/bwsf/blob/main/docs/specs/README.md)

## 主な機能

- **安全なストレージ** — 管理対象ファイル（`.env*` / `*.tfvars` / `*.tfvars.json`）を Bitwarden のボールトに安全に保存します。
- **簡単同期** — シンプルなコマンドで、ローカルと Bitwarden 間で管理対象ファイルをプッシュ・プル。
- **マルチ環境** — 1つのプロジェクトで複数ファイル（`.env`、`.env.staging`、`terraform.tfvars` など）を管理。
- **クロスプラットフォーム** — macOS と Linux に対応。Windows サポートは計画中です。

## クイックスタート

```bash
# Homebrew でインストール
brew tap b4m-oss/tap && brew install bwsf

# 初期設定 + 認証
bwsf setup
bwsf auth login

# Bitwarden から管理対象ファイルをプル
cd /path/to/your_project
bwsf pull

# Bitwarden に管理対象ファイルをプッシュ
bwsf push
```

## 仕組み

bwsf は Bitwarden **API** を使用します（`bw` CLI は不要）。管理対象ファイルは Bitwarden フォルダ（`target_section`、デフォルト名: `dotenvs`、setup で変更可）内の **Secure Note** として保存されます。

プロジェクト名は、`.git` ルートのディレクトリ名（無ければ cwd のベース名）。`.bwsf/config.jsonc` の `override_project_name` で上書きできます。
