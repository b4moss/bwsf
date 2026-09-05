# 開発予定中の機能

- [x] `dotenvs` フォルダ以外の名称利用可能化（`bwsf setup --folder` / ホストの `target_section`）
- [x] `bwsf clean` コマンド：リモートバックアップ確認後にローカルの管理対象ファイルを削除
- [x] `.git` によるプロジェクトルート解決（#134）
- [x] `.bwsf/config.(json|jsonc)` プロジェクト設定（#133 / #177）
  - `override_project_name`、任意の `host`、`save_files`（`!` 除外対応。`not_save_files` は削除）
- [x] グローバルマルチホスト設定 v2（#177）— `~/.config/bwsf/config.jsonc`
- [ ] ホストごとの Keychain / unlock·lock（#153）
- [x] `auth login` / `logout`（#174）
- [ ] `bwsf init`（#193）

# バージョニングについて

`bwsf` はセマンティック・バージョニングを採用しています。

現在、v1.0.0 のリリースは未定です。

v0.x.0 のアップデートは、常に破壊的変更が考えられます。

プロダクションでの利用は、十分ご注意ください。
