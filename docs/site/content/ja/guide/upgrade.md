# アップグレード

現在、`bwsf` は開発段階にあり、活発にアップデートされます。

brew からインストールした場合、手動でアップデートする必要があります。

```bash
brew upgrade bwsf
```

## v0.20.0 の破壊的変更

- グローバル設定パス: `~/.config/bwsf/config.jsonc`（schemaVersion 1、マルチホスト `settings.hosts`）
- 旧 flat の `config.json` は初回ロード時に確認後（または `--yes`）で移行。元ファイル横に `.bak-<timestamp>` バックアップを作成
- `not_save_files` を削除（グローバル・プロジェクトとも）。代わりに `save_files` の `!` 接頭辞を使用
- `backend` フィールドと `bwsf backend` を削除 — API のみ（`bw` CLI 経路は廃止）
- 保管庫コマンドは `--host <id>` を受け付け（解決順: CLI → プロジェクトの `host` → `is_default`）

製品仕様: [v0.20.0-multi-host.md](https://github.com/b4moss/bwsf/blob/main/docs/specs/v0.20.0-multi-host.md)
