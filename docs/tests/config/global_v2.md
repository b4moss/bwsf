# config: グローバル設定 v2（Issue #177 / v0.20.0 §2.1–2.2 / §2.4）

対象パッケージ: `app/src/config`（必要なら表示整形も同パッケージ）  
Issue: [#177](https://github.com/b4moss/bwsf/issues/177)  
製品正本: [`../specs/v0.20.0-multi-host.md`](../specs/v0.20.0-multi-host.md) §2.1・§2.2・§2.4

関連:

- マイグレーション → [`migrate_v2.md`](./migrate_v2.md)
- `save_files` / `!` → [`save_files_bang.md`](./save_files_bang.md)
- host 解決 → [`host_resolve.md`](./host_resolve.md)
- setup → [`../cmd/setup_v2.md`](../cmd/setup_v2.md)
- 旧 JSONC 読み込み契約 → [`jsonc_load.md`](./jsonc_load.md)（**v0.20 では本仕様が優先**）

---

### 前提（本仕様で固定）

| # | 方針 |
|---|------|
| G1 | 正式パスは **`~/.config/bwsf/config.jsonc`**。読み込みは `.json` と `.jsonc` の **いずれか一方**。両方あるとエラー |
| G2 | `.json` → 厳密 JSON（コメント・末尾カンマ不可）。`.jsonc` → JSONC（hujson。コメント・末尾カンマ可） |
| G3 | 書き込みは **常に `.jsonc`**。コメント保持は可能なら行い、不可なら pretty JSON でよい |
| G4 | `.json` のみから書き移すとき: `.jsonc` を書き、旧 `.json` を削除する |
| G5 | トップレベルは `schemaVersion` / `created_at` / `updated_at` / `app_version` / `settings`。旧 flat キー（`host_type` 等）を新ファイルに残さない |
| G6 | `schemaVersion` は `1`。未知・欠落・0 以下はロードエラー |
| G7 | `created_at` は初回作成時のみ（ISO8601・TZ 付き）。`updated_at` / `app_version` は書き込みのたびに更新 |
| G8 | `settings.hosts` は **空配列を許容**。非空時は `is_default: true` が全体でちょうど1つ |
| G9 | host 要素: `id`（空白不可・`/` 禁止・重複禁止・印字可能 Unicode 可）、`type`（`bitwarden-cloud` \| `bitwarden-selfhost`）、`host_url` 必須、`target_section` 必須（空エラー）、`email` / `device_identifier` 任意 |
| G10 | 新スキーマに `backend` フィールドは持たない。出現したらスキーマエラー（旧ファイルはマイグレーション側） |
| G11 | ファイル無し → `LoadConfig` 相当は `nil, nil`（現行どおり）。未作成を「空スキーマ」として成功扱いしない |
| G12 | Keychain キー書き換えは **対象外**（[#153](https://github.com/b4moss/bwsf/issues/153)） |

### スキーマ例

```jsonc
{
  "schemaVersion": 1,
  "created_at": "2026-09-03T00:00:00+09:00",
  "updated_at": "2026-09-03T00:00:00+09:00",
  "app_version": "0.20.0",
  "settings": {
    "save_files": [".env", ".env.*"],
    "hosts": [
      {
        "id": "default",
        "type": "bitwarden-cloud",
        "host_url": "https://vault.bitwarden.com",
        "email": "user@example.com",
        "target_section": "dotenvs",
        "is_default": true,
        "device_identifier": "optional-until-first-use"
      }
    ]
  }
}
```

`hosts: []` のみ（ファイル選択用途）も合法。

---

## 1. パス探索

テストは `HOME` 差し替え。

#### テスト：正常系

| # | 配置 | 期待 |
|---|------|------|
| P1 | `config.jsonc` のみ | それを読む |
| P2 | `config.json` のみ | それを読む（厳密 JSON） |
| P3 | どちらも無し | `nil, nil` |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| P4 | 同一 dir に `.json` と `.jsonc` の両方 | エラー（両方ある旨） |

---

## 2. パース

#### テスト：正常系

| # | 入力 | 期待 |
|---|------|------|
| R1 | `.json` の通常 JSON（新スキーマ） | 各フィールドが期待どおり |
| R2 | `.jsonc`（`//`・ブロックコメント・末尾カンマ） | 成功し同値 |
| R3 | `hosts: []`、`save_files` 省略 | 成功 |

#### テスト: 異常系

| # | 入力 | 期待 |
|---|------|------|
| R4 | `.json` にコメントまたは末尾カンマ | エラー |
| R5 | 壊れたトークン | エラー（`failed to parse` 相当を維持可） |
| R6 | 空バイト / 空白のみ | エラー |

---

## 3. バリデーション（ホスト・メタ）

ロード後または専用 `Validate` でよい。

#### テスト：正常系

| # | 入力 | 期待 |
|---|------|------|
| V1 | host 1件・`is_default: true` | OK |
| V2 | host 2件・ちょうど1つが default | OK |
| V3 | `hosts: []`（`is_default` 無し） | OK |
| V4 | `id` に Unicode（例: `東京`）・`/` 無し | OK |
| V5 | `email` / `device_identifier` 省略 | OK |

#### テスト: 異常系

| # | 入力 | 期待 |
|---|------|------|
| V6 | `schemaVersion` 欠落または `!= 1` | エラー |
| V7 | 非空 `hosts` で `is_default` が0または2以上 | エラー |
| V8 | `id` 空・空白のみ・`/` 含む | エラー |
| V9 | `id` 重複 | エラー |
| V10 | 未知の `type` | エラー |
| V11 | `host_url` 欠落または空 | エラー |
| V12 | `target_section` 欠落または空 | エラー |
| V13 | 新スキーマに `backend` キーあり | エラー |
| V14 | 新スキーマに `not_save_files` あり | エラー（詳細は [`save_files_bang.md`](./save_files_bang.md)） |

---

## 4. 保存（`SaveConfig`）

#### テスト：正常系

| # | 手順 | 期待 |
|---|------|------|
| S1 | 新規保存 | パスは `config.jsonc`。`schemaVersion=1`。`created_at` / `updated_at` が TZ 付き。`app_version` が現行バイナリ版 |
| S2 | 既存を再保存 | `created_at` は不変。`updated_at` / `app_version` は更新 |
| S3 | 元が `config.json` のみの新スキーマを保存 | `config.jsonc` が書き出され、旧 `config.json` は削除 |
| S4 | 保存ファイルを再ロード | ラウンドトリップで論理値が一致 |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| S5 | 無効な Host を保存しようとする | エラー（ディスクに不正スキーマを残さない） |

---

## 5. `bwsf config show`（v0.20 表示契約）

旧表示契約: [`../cmd/config_show.md`](../cmd/config_show.md)（flat フィールド）。**v0.20 では本節が優先。**

#### テスト：正常系

| # | 条件 | 期待 |
|---|------|------|
| C1 | host 1件（default）+ `save_files` あり | 人間向けラベルで hosts（id / type / url / email / target_section / default 印）と `save_files` が出る。秘密は出ない |
| C2 | `hosts: []` | hosts 空であることが分かる。Bitwarden アクセスなし |
| C3 | 読み取り専用 | 成功パスでファイル書き込みなし |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| C4 | ファイル無し | エラー（`setup` を促す） |
| C5 | 壊れた／スキーマ不正 | ロードエラーを伝播 |

---

## 6. 対象外

- 旧 flat → 新スキーマ変換（[`migrate_v2.md`](./migrate_v2.md)）
- Keychain / unlock / auth / `bwsf init`
- プロジェクト `.bwsf/config.*` の探索順（#133。キー変更は [`save_files_bang.md`](./save_files_bang.md)）

---

## 実装メモ（仕様外・参考）

```text
HOME/.config/bwsf/{config.json|config.jsonc} XOR
  → parse (.json strict | .jsonc hujson)
  → validate schemaVersion + hosts
  → Save → always config.jsonc (+ delete .json if needed)
```
