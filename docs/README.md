# bwsf

プロダクトの意味的な pillar 正本（目的・スコープ・技術方針のハブ）。  
OKF の版索引は [`index.md`](./index.md)（`okf_version: "0.1"`）。本文はここに書く。

## 概要

Bitwardenを.envの管理に用いるCLIツール。

- 依存：bwコマンド

## 前提

- Bitwarden Cloudか、セルフホストのBitwarden環境があること
- それらで利用可能な、Bitwardenアカウント（メールアドレスとパスワードがあること）
- bwコマンドがインストールされていること

## 仕様

Bitwarden内に、dotenvsというフォルダを、.env専用のフォルダとして利用する
（従って、dotenvsという名称は、予約語となる）

dotenvsディレクトリのログインアイテムのドメイン名に相当する部分が「プロジェクト名」となる
例：「my-todo-app」など

## 対応ファイル

### 対象ファイル

`.env`で始まるすべてのファイルが対象となります。

例:
- `.env`
- `.env.local`
- `.env.development`
- `.env.staging`
- `.env.production`

### 除外ファイル

ファイル名に`.example`を含むファイルは除外されます。

例:
- `.env.example` → 除外
- `.env.local.example` → 除外
- `.env.example.local` → 除外

### 注意事項

- 上記の条件を満たすファイルが1つ以上あれば、pushコマンドを実行可能です
- `.env`ファイルが存在しなくても、`.env.local`などがあれば問題ありません

## 機能

### bwsf pull --output <dirname>

カレントディレクトリ名をプロジェクト名として、dotenvsの中を検索
一致するプロジェクトに保存されている.env情報を取り出し、カレントディレクトリまたは、指定ディレクトリに展開する
既に、.envファイルがディレクトリ内に存在する場合は、上書きするかどうかをy/Nで質問

### bwsf push --from <dirname>

カレントディレクトリ名をプロジェクト名として、dotenvsの中を検索
一致するプロジェクト名のBitwardenログインアイテムがあれば、上書きするかどうかをy/Nで質問

### bwsf list

Bitwardenのdotenvs内にあるアイテムの一覧をリスト表示

## 索引

- [roadmap](./roadmap.md) — マイルストーン
- [specs](./specs/) — 現行仕様（ドメイン別）
- [plans](./plans/) — これからやる内容
- [tests](./tests/) — テスト仕様（specs と同じドメイン切り）
- [憲章](./charter/) — 開発ルール
- [OKF v0.1](./charter/okf/) — 知識バンドルの版定義
- [override-charter](./override-charter.md) — 憲章オーバーライド
- [site](./site/) — 公開ドキュメントサイト（シェル）

----

以上
