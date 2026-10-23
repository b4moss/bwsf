# bwsf init

現行の `bwsf init` の仕様正本。テスト仕様: [`../../tests/cmd/init.md`](../../tests/cmd/init.md)。

関連 Issue: [#193](https://github.com/b4moss/bwsf/issues/193)

## 決定

| 項目 | 決定 |
|------|------|
| 生成 | 常に `.bwsf/config.jsonc` |
| 内容 | 対話（`host` / スキップ、`save_files` / 未設定、`override_project_name` 等） |
| `host` | **任意**。グローバル `hosts[]` から選ぶかスキップ。書いた場合はチーム共有用の id 参照（[`../config/host-resolve.md`](../config/host-resolve.md)） |
| 既存あり | 確認後上書き。`--yes` でスキップ |
| グローバル設定ファイル無し | エラー（先に `setup`。`hosts` 空のグローバルは「有り」とみなす） |
| `hosts` が空のとき | `host` プロンプトはスキップ可（プロジェクトに `host` キーを書かない） |

```jsonc
{
  "host": "default",
  "override_project_name": "my-api",
  "save_files": [".env*", "!.env.local"]
}
```

`host` 無しの例:

```jsonc
{
  "override_project_name": "my-api",
  "save_files": [".env*", "!.env.local"]
}
```

## 関連

- setup: [`setup.md`](./setup.md)
- save_files: [`../config/save-files.md`](../config/save-files.md)
- 履歴: [`../../_archived/history/v0.20.0-multi-host.md`](../../_archived/history/v0.20.0-multi-host.md) §5

----

以上
