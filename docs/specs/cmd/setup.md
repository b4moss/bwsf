# bwsf setup

現行の `bwsf setup` の仕様正本。テスト仕様: [`../../tests/cmd/setup_v2.md`](../../tests/cmd/setup_v2.md)。

関連 Issue: [#177](https://github.com/b4moss/bwsf/issues/177)

## 決定

| 項目 | 決定 |
|------|------|
| host | 対話で「追加する / スキップ」を選べる。**スキップ可**（`hosts: []` のまま保存してよい） |
| 追加する場合の初期値 | `id: "default"`, `is_default: true`（既存が空のとき） |
| 既に host があるとき | 対話で「追加 / 既存更新 / デフォルト変更 / スキップ」 |
| ファイル選択 | 対話で「`save_files` を設定 / 未設定」。設定時は glob を入力（否定は `!` 接頭辞）。host をスキップしてもファイル選択は可 |

Login は行わない。認証は `auth login`。

## 関連

- グローバル設定: [`../config/global-v2.md`](../config/global-v2.md)
- save_files: [`../config/save-files.md`](../config/save-files.md)
- コマンド一覧: [`commands.md`](./commands.md)
- 履歴: [`../../_archived/history/v0.20.0-multi-host.md`](../../_archived/history/v0.20.0-multi-host.md) §2.5

----

以上
