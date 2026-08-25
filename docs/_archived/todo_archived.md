# Archived TODOs

## docs/todo.md retirement (v0.15.0 docs sync)

`docs/todo.md` was removed. Remaining open items were moved to GitHub Issues:

| Former todo.md item | Issue |
|---|---|
| Homebrew 本家への登録 | #129 (Until v1.0.0) |
| bwコマンド依存からの脱却 | #53 (Until v1.0.0) |
| マスターパスワード入力の負担軽減 | #130 (After v1.0.0) |
| Windows 版ビルド＆ Chocolatey 公開 | #131 (After v1.0.0) |

Already done / covered elsewhere: docs site (#71), local Vaultwarden smoke (#108–#110), remote smoke (#112), custom folder name (#54).

## v0.10.0
- ✅ Linux版ビルド＆Homebrew公開
- ✅ Bugfix: user GPG unverified

## v0.9.1

- ✅ Bugfix: プルリク作成時に、テストが発生しない

## v0.9.0

- ✅ local/develop/staging/productionなどの、各環境への対応
- ✅ dotenvsフォルダがない場合、新規作成するかどうか

## v0.8.0

- ✅ E2Eテスト（モック）

## v0.7.0

- ✅ enhance: stdoutの待機でアニメーションをつける
- ✅ プルリクマージ時に、CI/CDでテストを行う（逆に、リリース時は行わない）

## v0.6.0

- ✅ GPG署名
- ✅ ドキュメント整備
- ✅ Goバージョンの差異を統一

## v0.5.0

- ✅ brew tapでの、配布。まずはmacOSビルドのみ
- ✅ ローカルのmacOSで実行されるかを確認


## v0.4.0

- ✅ 単体テストの実装
- ✅ ソースコード内へのバージョン指定

## v0.3.0

- ✅ bw push
- ✅ bw pull

## v0.2.0

### Feat: bwコマンドの存在をチェック

- 現在の端末に、bwコマンドがインストールされているかどうかを確認
- インストールされていれば「[INFO] ✅ bw command is installed!」と標準出力
- インストールされていなければ「[ERROR] ❌ bw command is not installed...」と標準出力

### Feat: Bitwardenセットアップ

- Bitwardenホストにログインする
- Bitwarden Cloudか、セルフホストかを確認する。対話で選択式。
- セルフホストの場合は、URLを入力させる
- Eメールアドレス、パスワードを入力させる。パスワードは入力を隠す。
- bwコマンドを使って、ログインを試みる
- 失敗したらエラーメッセージをそのまま表示
- ログインできたら「[INFO] ✅ Sign in to Bitwarden was successful!」と標準出力

### Fix: .configを永続化

現状、.configをDocker内に作っているが、デバッグがしづらいので、永続化したい
docker containerも立ち上がったままにしたい

### Feat: bwenv listコマンド

- bwenv listと入力すると
- Bitwardenホスト内の「dotenvs」フォルダを探す
- dotenvsフォルダがなければエラー
- dotenvsフォルダが存在すれば、その中のアイテムの名称一覧を出力

### Enhance: 標準出力に色をつけたい

- 赤：ERROR
- 緑：何かを実行して、SUCCESSした時
- 黄：WARNING
- 水色：重要なINFO、または決定事項
- 薄紫：対話の質問

### Enhance: cloudかselfhostedかの質問

- 現状: 文字入力させている
- 改善: 選択肢を表示させ、矢印キーで選択、Enterで確定させたい