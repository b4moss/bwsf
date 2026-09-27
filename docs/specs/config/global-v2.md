# グローバル設定 v2

現行のグローバル設定（パス・スキーマ・`hosts[]`・マイグレーション）の仕様正本。  
テスト仕様: [`../../tests/config/global_v2.md`](../../tests/config/global_v2.md) / [`../../tests/config/migrate_v2.md`](../../tests/config/migrate_v2.md)。

関連 Issue: [#177](https://github.com/b4moss/bwsf/issues/177)

## バックエンド

API のみ。`backend` フィールドは新スキーマに持たない。`bw` CLI 経路は廃止。

## パスとパーサー

| 項目 | 決定 |
|------|------|
| 正式パス | `~/.config/bwsf/config.jsonc` |
| 読み込み | `.json` と `.jsonc` のいずれか一方。両方あるとエラー |
| パーサー | `.json` → 厳密 JSON。`.jsonc` → JSONC |
| 書き込み | 常に `.jsonc`。中身は `json.MarshalIndent` の **pretty JSON**（コメントは保持しない） |
| `.json` のみから書き移すとき | `.jsonc` を書き、旧 `.json` は削除 |

プロジェクト（`.bwsf/config.*`）も同じパーサー規則。新規生成は常に `.jsonc`。

## トップレベル

```jsonc
{
  "schemaVersion": 1,
  "created_at": "2026-09-03T00:00:00+09:00",
  "updated_at": "2026-09-03T00:00:00+09:00",
  "app_version": "0.20.0",
  "settings": {
    "save_files": [".env", ".env.*"],
    "hosts": []
  }
}
```

| 項目 | 決定 |
|------|------|
| `schemaVersion` | `1` |
| `created_at` | 初回作成時のみ。ISO8601（TZ 付き） |
| `updated_at` / `app_version` | 書き込みのたびに更新。`app_version` はその時点の bwsf バージョン |

ファイル選択（`save_files`）は [`save-files.md`](./save-files.md)。

## `hosts[]` 要素

```jsonc
{
  "id": "default",
  "type": "bitwarden-cloud",
  "host_url": "https://vault.bitwarden.com",
  "email": "user@example.com",
  "target_section": "dotenvs",
  "is_default": true,
  "device_identifier": "optional-until-first-use"
}
```

| 項目 | 決定 |
|------|------|
| 配列 | **空配列を許容**（host 未設定。プロジェクト設定やファイル選択だけ使う利用者向け） |
| `id` | 必須（要素がある場合）。空白不可。`/` 禁止。重複禁止。印字可能（Unicode 可） |
| `type` | `"bitwarden-cloud"` \| `"bitwarden-selfhost"`（将来追加可） |
| `host_url` | 必須。cloud でも書く |
| cloud の setup 初期値 | `https://vault.bitwarden.com` |
| `email` | 任意 |
| `target_section` | 必須。空はエラー |
| `is_default` | `hosts` が1件以上のとき、全体でちょうど1つが `true`。空配列のときは不要 |
| `device_identifier` | host ごと。無ければ初回 API 利用時に生成して書き戻す |
| 初回 setup で作る場合の初期 `id` | `"default"`（特別扱いではない） |

## マイグレーション

| 項目 | 決定 |
|------|------|
| 検出時 | 実行前に確認対話 |
| 実行時 | バックアップ後に変換保存 |
| 拒否 | エラー終了（旧形式のまま動かさない） |
| 非対話 | `--yes` なら実行。なければエラー |
| 旧 `backend: "bw"` | 構造変換し `backend` は捨てる。API 前提の警告 |

| 旧 | 新 |
|----|----|
| （なし） | `hosts[0].id = "default"`, `is_default = true` |
| `host_type: "cloud"` | `type: "bitwarden-cloud"`, `host_url: "https://vault.bitwarden.com"` |
| `host_type: "selfhosted"` | `type: "bitwarden-selfhost"`, `host_url: <selfhosted_url>` |
| `email` | `email`（空なら省略） |
| `folder_name`（空なら `"dotenvs"`） | `target_section` |
| `device_identifier` | 当該 host へ |
| `backend` | 捨てる |
| ファイル選択 | 付けない |

旧 Keychain グローバルキーは `hosts/default/...` へ移行または読み替え（詳細は [`../infra/secretstore-hosts.md`](../infra/secretstore-hosts.md)）。

## 関連

- host 解決: [`host-resolve.md`](./host-resolve.md)
- setup: [`../cmd/setup.md`](../cmd/setup.md)
- 履歴: [`../../_archived/history/v0.20.0-multi-host.md`](../../_archived/history/v0.20.0-multi-host.md) §2

----

以上
