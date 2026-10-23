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
- フラット `bwsf auth` / `auth --clear` を削除 — `bwsf auth login` / `auth logout` を使用（`logout` は `vault_unlock` も削除）

## 製品仕様（OKF）

- OKF 索引: [`docs/index.md`](https://github.com/b4moss/bwsf/blob/main/docs/index.md)
- pillar: [`docs/README.md`](https://github.com/b4moss/bwsf/blob/main/docs/README.md)
- 仕様索引: [`docs/specs/`](https://github.com/b4moss/bwsf/blob/main/docs/specs/README.md)
- マルチホスト各ドメイン: [`config/global-v2`](https://github.com/b4moss/bwsf/blob/main/docs/specs/config/global-v2.md), [`config/host-resolve`](https://github.com/b4moss/bwsf/blob/main/docs/specs/config/host-resolve.md), [`config/save-files`](https://github.com/b4moss/bwsf/blob/main/docs/specs/config/save-files.md), [`cmd/auth`](https://github.com/b4moss/bwsf/blob/main/docs/specs/cmd/auth.md), [`cmd/unlock-lock`](https://github.com/b4moss/bwsf/blob/main/docs/specs/cmd/unlock-lock.md), [`cmd/init`](https://github.com/b4moss/bwsf/blob/main/docs/specs/cmd/init.md)
