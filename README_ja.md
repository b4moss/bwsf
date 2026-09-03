# bwsf

[![CI](https://github.com/b4moss/bwsf/actions/workflows/test.yml/badge.svg)](https://github.com/b4moss/bwsf/actions/workflows/test.yml)
[![Coverage](https://img.shields.io/codecov/c/github/b4moss/bwsf)](https://codecov.io/gh/b4moss/bwsf)
[![Go Reference](https://pkg.go.dev/badge/github.com/b4moss/bwsf.svg)](https://pkg.go.dev/github.com/b4moss/bwsf)
[![Release](https://img.shields.io/github/v/release/b4moss/bwsf)](https://github.com/b4moss/bwsf/releases)
[![License](https://img.shields.io/github/license/b4moss/bwsf)](https://github.com/b4moss/bwsf/blob/main/LICENSE)
[![OpenSSF Scorecard](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fapi.scorecard.dev%2Fprojects%2Fgithub.com%2Fb4moss%2Fbwsf&query=%24.score&label=OpenSSF%20Scorecard&suffix=%2F10)](https://scorecard.dev/viewer/?uri=github.com/b4moss/bwsf)

bwsf（Bitwarden Secured Files）は、[Bitwarden](https://bitwarden.com/)を使用して `.env*` および Terraform の `*.tfvars` / `*.tfvars.json` を管理するCLIツールです。
[Official site](https://bwsf.oss.b4m.jp/)

[English version is here.](./README.md)

## 🚨🚨重要なお知らせ🚨🚨

**v0.20.0 破壊的変更:** グローバル設定は `~/.config/bwsf/config.jsonc`（多ホストスキーマ）へ。旧 flat の `config.json` は初回実行時に確認（または `--yes`）で移行します。`not_save_files` は廃止（`save_files` の `!` 接頭辞を使用）。`bw` CLI バックエンドと `bwsf backend` は廃止（API のみ）。

v0.17.0、v0.17.1では、正常に Bitwarden にログインできない事象が確認されています。v0.17系統をご利用の方は、先に v0.17.2 以降へアップデートしてください。

## 概要

bwsf は Bitwarden 上でプロジェクトファイル（`.env*` / `*.tfvars` / `*.tfvars.json`。名前に `.example` を含むものは除外）を管理します。

簡単な使用方法は以下の通りです：

| コマンド | |
|----|----|
| bwsf setup | ホストとグローバル `save_files` の設定（API のみ） |
| bwsf auth | Personal API Key の保存と認証 |
| bwsf config show | 現在のローカル設定を表示 |
| bwsf push | 管理対象ファイルを Bitwarden ホストにプッシュ |
| bwsf pull | Bitwarden ホストから管理対象ファイルをプル |
| bwsf list | Bitwarden ホストに保存されているプロジェクトの一覧を表示 |
| bwsf clean | リモートバックアップを確認したうえでローカルの管理対象ファイルを削除 |

## 動機

私たちは長い間、Bitwardenをパスワードマネージャーとして使用しています。
また、.envファイルをBitwardenホストに保存し、シェルスクリプトとして管理しています。
このプロジェクトは、手作りのシェルスクリプトをGoで書かれたモダンなCLIコマンドに移行するものです。

## 要件

Bitwarden CLI（`bw`）は不要です。bwsf は **API** で Bitwarden と通信します。

必要なもの:

- Bitwarden アカウント（Cloud またはセルフホスト / Vaultwarden）
- **Personal API Key**（アカウント設定 → セキュリティ → キー）
- OS の秘密保管（**macOS Keychain** / **Linux secret service**）へのアクセス

### Homebrew

bwsf 本体のインストールに Homebrew を使います。

### 対応OS

- macOS
- Linux
- [計画中] Windows

## インストール

| OS | コマンド |
|----|----|
| macOS / Linux| brew tap b4m-oss/tap && brew install bwsf |

過去バージョンのインストール。versioned formula は「現行 minor の全 patch」と「1つ前の minor の最新 patch」のみ tap に残します（それ以外は削除）。それより古い版は [GitHub Releases](https://github.com/b4moss/bwsf/releases) を参照してください。

```shell
brew install bwsf@0.18.0
```

## インストールの確認

```shell
bwsf -v
# bwsf version 0.20.0
```

## 使い方

### 初期セットアップ

```shell
bwsf setup
```

Bitwarden ホスト（スキップ可）と任意のグローバル `save_files` を設定します。設定は `~/.config/bwsf/config.jsonc` に保存されます。

デフォルトでは、ノートは Bitwarden の `dotenvs` フォルダ（各 host の `target_section`）に保存されます。別のフォルダ名を使う場合:

```shell
bwsf setup --folder my-envs
```

フォルダ名を変更しても既存ノートは自動では移動しません。必要なら Bitwarden 上で手動移動してください。

複数ホストがある場合、`push` / `pull` / `list` / `clean` で `--host <id>` を指定できます。

保存内容の確認:

```shell
bwsf config show
```

### Bitwardenホストから管理対象ファイルをプル

```shell
cd /path/to/your_project
bwsf pull
```

bwsf はカレントディレクトリ名に一致する Note を Bitwarden ホストから検索します。
存在する場合、管理対象ファイルをカレントディレクトリへ書き出します。
同名のローカルファイルがある場合は、ファイル単位で上書き確認します。

### Bitwardenホストに管理対象ファイルをプッシュ

```shell
cd /path/to/your_project
bwsf push
```

bwsf はカレントディレクトリの管理対象ファイルを Bitwarden ホストへプッシュします。
設定フォルダ（デフォルト: `dotenvs`）に同じ名前の Note がある場合は、**上書き確認なしで更新**します。

### Bitwardenホストのプロジェクト一覧

```shell
bwsf list
```

Bitwarden ホスト上のプロジェクト名を、標準出力に 1 行ずつ表示します。

### ローカル管理対象ファイルの削除

```shell
cd /path/to/your_project
bwsf clean
```

リモートのバックアップを確認したうえで、ローカルの管理対象ファイルを削除します。

### 典型的な流れ

```shell
bwsf setup                 # ホスト + 任意の save_files（＋任意で folder）
bwsf auth                  # Personal API Key を保存し Identity トークン取得
bwsf push                  # マスターパスワードで unlock して同期
bwsf pull
bwsf list
```

`bwsf auth` は `client_id` / `client_secret` の入力を求め、OS の秘密保管（**macOS Keychain** / **Linux secret service**）に保存し、Identity のアクセストークンを取得します（トークンはプロセスのメモリ上のみ）。削除は `bwsf auth --clear` です。

Personal API Key は Bitwarden Web 保管庫の アカウント設定 → セキュリティ → キー から作成できます。

`push` / `pull` / `list` のたびに **マスターパスワード** の入力を求め、メモリ上でボルト鍵を復元し、コマンド終了時に鍵・トークンを破棄します。

補足: unlock は Community SDK のパスワード login 経路（config の email + マスターパスワード）で鍵を復元します。Identity の Personal API Key トークンとは別管理です。

### v0.19 からのアップグレード

初回実行時、旧 `~/.config/bwsf/config.json`（flat）を検出すると確認後（または `--yes`）に移行します。バックアップは元ファイル横に作成されます。プロジェクト設定に `not_save_files` が残っている場合は `save_files` + `!` へ手動で書き換えてください（残るとロードエラー）。

## アンインストール

```shell
brew uninstall bwsf
```

## FAQ

<details>
<summary>Q. Bitwardenアカウントを持っていません。</summary>

bwsfを使用するには、Bitwardenアカウントが必要です。

[Bitwarden Cloud](https://bitwarden.com/)にアクセスして、アカウントを登録できます。

無料で、クレジットカードも不要です。

</details>

<details>
<summary>Q. Bitwardenのセルフホストユーザーです。</summary>

もちろん、bwsfはBitwardenのセルフホストユーザーでも利用可能です。

初期セットアップ時にセルフホストのURLを入力できます。

</details>

<details>
<summary>Q. .envファイルはBitwardenホストにどのように保存されますか？</summary>

管理対象ファイルはJSON形式に変換されます。bwsfはBitwardenのNoteアイテムを作成し、NoteセクションにそのJSONを保存します。

</details>

<details>
<summary>Q. Bitwardenのアカウント情報はどこに保存されますか？</summary>

bwsfは設定データを`~/.config/bwsf/`に保存します（正式なファイル: `config.jsonc`）。

ただし、セキュリティ情報（パスワードなど）は一切保存されません。

- **macOS & Linux**: Homebrew
- **Windows**: Chocolaty

## 開発

### 要件

**Docker** が開発マシンにインストールされている必要があります。

### 開発環境の起動

```
git clone https://github.com/b4moss/bwsf.git
cd bwsf
make run
```

## ライセンス

[MIT License](./LICENSE)
