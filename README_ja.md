# bwsf

[![CI](https://github.com/b4moss/bwsf/actions/workflows/test.yml/badge.svg)](https://github.com/b4moss/bwsf/actions/workflows/test.yml)
[![Coverage](https://img.shields.io/codecov/c/github/b4moss/bwsf)](https://codecov.io/gh/b4moss/bwsf)
[![Go Reference](https://pkg.go.dev/badge/github.com/b4moss/bwsf.svg)](https://pkg.go.dev/github.com/b4moss/bwsf)
[![Release](https://img.shields.io/github/v/release/b4moss/bwsf)](https://github.com/b4moss/bwsf/releases)
[![License](https://img.shields.io/github/license/b4moss/bwsf)](https://github.com/b4moss/bwsf/blob/main/LICENSE)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/b4moss/bwsf/badge)](https://securityscorecards.dev/viewer/?uri=github.com/b4moss/bwsf)

bwsf（Bitwarden Secured Files）は、[Bitwarden](https://bitwarden.com/)を使用して `.env*` および Terraform の `*.tfvars` / `*.tfvars.json` を管理するCLIツールです。
[Official site](https://bwsf.oss.b4m.jp/)

[English version is here.](./README.md)

## 🚨🚨重要なお知らせ🚨🚨

v0.17.0、v0.17.1では、正常に Bitwarden にログインできない事象が確認されています。

v0.17系統をご利用の方は、v0.17.2(日本時間2026/9/3リリース)にアップデートをお願いします。

## 概要

bwsf は Bitwarden 上でプロジェクトファイル（`.env*` / `*.tfvars` / `*.tfvars.json`。名前に `.example` を含むものは除外）を管理します。

簡単な使用方法は以下の通りです：

| コマンド | |
|----|----|
| bwsf setup | Bitwarden ホストとアカウントの設定 |
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

**`bw`** コマンドがマシンにインストールされている必要があります。

[bwコマンドのインストール方法については、こちらのドキュメントをお読みください。](https://bitwarden.com/help/cli/#download-and-install)

** Homebrew **: インストールに必要です。

### 対応OS

- macOS
- Linux
- [計画中] Windows

## インストール

| OS | コマンド |
|----|----|
| macOS / Linux| brew tap b4m-oss/tap && brew install bwsf |

過去バージョンのインストール（公開済みの全リリースが tap にあります）:

```shell
brew install bwsf@0.15.0
```

## インストールの確認

```shell
bwsf -v
# bwsf version 0.16.0
```

## 使い方

### 初期セットアップ

```shell
bwsf setup
```

Bitwardenホストとアカウント情報を設定します。

デフォルトでは、ノートは Bitwarden の `dotenvs` フォルダに保存されます。別のフォルダ名を使う場合:

```shell
bwsf setup --folder my-envs
```

フォルダ名は `~/.config/bwsf/config.json` に保存され、push / pull / list / clean で参照されます。フォルダ名を変更しても既存ノートは自動では移動しません。必要なら Bitwarden 上で手動移動してください。

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

bwsfは設定データを`~/.config/bwsf/`に保存します。

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
