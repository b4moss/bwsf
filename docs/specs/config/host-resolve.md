# host 解決

現行に存在する host 解決の仕様正本。テスト仕様: [`../../tests/config/host_resolve.md`](../../tests/config/host_resolve.md)。

関連 Issue: [#177](https://github.com/b4moss/bwsf/issues/177)

## `--host` 対象コマンド

`push` / `pull` / `list` / `clean` / `unlock` / `lock` / `auth login` / `auth logout`

## 解決順

優先度（高い方から）:

1. CLI `--host <id>`
2. プロジェクト設定の `host`（任意。リポジトリ共有用）
3. グローバル `hosts[]` のうち `is_default: true`

| 状況 | 結果 |
|------|------|
| 解決できた | その host を使う |
| 指定 id が `hosts[]` に無い | エラー |
| いずれも無く、`hosts` が空 | エラー（先に `setup` で host を追加するか、プロジェクト／CLI で指定） |
| `hosts` は非空だが `is_default: true` が無い／複数 | スキーマエラー（ロード時） |

プロジェクトの `host` は **id の参照のみ**（接続情報自体はグローバル `hosts[]` に持つ）。チームで同じ id を共有し、各メンバーのグローバルに同 id のエントリを置く想定。

## 関連

- グローバル設定: [`global-v2.md`](./global-v2.md)
- 履歴: [`../../_archived/history/v0.20.0-multi-host.md`](../../_archived/history/v0.20.0-multi-host.md) §1.1

----

以上
