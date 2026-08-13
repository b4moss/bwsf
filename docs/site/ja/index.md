---
layout: home

hero:
  name: bwsf
  text: 安全なファイル同期
  tagline: Bitwarden CLI で .env* と Terraform tfvars を管理
  actions:
    - theme: brand
      text: はじめる
      link: /ja/guide/getting-started
    - theme: alt
      text: GitHub で見る
      link: https://github.com/b4moss/bwsf

features:
  - icon: 🔐
    title: 安全なストレージ
    details: 管理対象ファイル（.env* / *.tfvars / *.tfvars.json）を Bitwarden のボールトに安全に保存します。
  - icon: 🔄
    title: 簡単同期
    details: シンプルなコマンドで、ローカルと Bitwarden 間で管理対象ファイルをプッシュ・プル。
  - icon: 📋
    title: マルチ環境
    details: 1つのプロジェクトで複数ファイル（.env、.env.staging、terraform.tfvars など）を管理。
  - icon: 🖥️
    title: クロスプラットフォーム
    details: macOS と Linux に対応。Windows サポートは計画中です。
---

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
