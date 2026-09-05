# config: グローバル設定 flat → v2 マイグレーション（Issue #177 / v0.20.0 §2.6）

対象パッケージ: `app/src/config`（確認対話の差し替え口は `cmd` / `utils` でも可）  
Issue: [#177](https://github.com/b4moss/bwsf/issues/177)  
製品正本: [`../specs/v0.20.0-multi-host.md`](../specs/v0.20.0-multi-host.md) §2.6

関連: [`global_v2.md`](./global_v2.md)

---

### 前提（本仕様で固定）

| # | 方針 |
|---|------|
| M1 | 旧 flat（トップレベル `host_type` / `selfhosted_url` / `email` / `folder_name` / `backend` / `device_identifier` 等、`schemaVersion` 無し）を検出したら、**変換前に確認**する |
| M2 | 対話で拒否 → **エラー終了**。旧形式のまま動作させない |
| M3 | 非対話（TTY 無し、または確認できない環境）では **`--yes` が無ければエラー**。`--yes` なら確認なしで実行 |
| M4 | 実行時は **バックアップ後**に新スキーマを保存する（例: 元パス横に `.bak-<timestamp>`。名前は実装で固定しテストする） |
| M5 | 写像は正本 §2.6 表どおり。`backend` は捨てる。ファイル選択（`save_files`）は付けない |
| M6 | 旧 `backend: "bw"` でも構造変換は行うが、**API 前提の警告**を出す |
| M7 | 変換後の保存先は `config.jsonc`。元が `config.json` ならバックアップ後に新形式へ置き換え（旧 `.json` は残さない／バックアップ側に残す） |
| M8 | **Keychain グローバルキーの `hosts/default/...` 移行は本仕様の対象外**（[#153](https://github.com/b4moss/bwsf/issues/153)）。設定ファイルの写像のみ固定する |

### フィールド写像

| 旧 | 新 |
|----|----|
| （id 無し） | `hosts[0].id = "default"`, `is_default = true` |
| `host_type: "cloud"` | `type: "bitwarden-cloud"`, `host_url: "https://vault.bitwarden.com"` |
| `host_type: "selfhosted"` | `type: "bitwarden-selfhost"`, `host_url: <selfhosted_url>` |
| `email`（空なら省略） | `email` |
| `folder_name`（空なら `"dotenvs"`） | `target_section` |
| `device_identifier` | 当該 host へ |
| `backend` | 捨てる |
| （ファイル選択） | 付けない |

`schemaVersion: 1`、`created_at` / `updated_at` / `app_version` は [`global_v2.md`](./global_v2.md) の保存契約に従う。

---

## 1. 検出

#### テスト：正常系（検出判定）

| # | ファイル内容 | 期待 |
|---|--------------|------|
| D1 | 旧 flat（`host_type` あり、`schemaVersion` なし） | マイグレーション候補として検出 |
| D2 | 新スキーマ（`schemaVersion: 1`） | マイグレーションしない（通常ロード） |
| D3 | ファイル無し | マイグレーションしない（`nil, nil`） |

---

## 2. 確認フロー

確認関数はテストでスタブ可能にする。

#### テスト：正常系

| # | 条件 | 期待 |
|---|------|------|
| C1 | 対話・ユーザー承認 | バックアップ作成 → 新 `config.jsonc` 保存 → 再ロードで新スキーマ |
| C2 | `--yes`（非対話可） | 確認プロンプト無しで C1 と同様に成功 |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| C3 | 対話・ユーザー拒否 | エラー。元ファイルは変更されない（バックアップも作らない、または作っても新ファイルは書かない——実装は「元を壊さない」をテストで固定） |
| C4 | 非対話かつ `--yes` 無し | エラー。元ファイル不変 |

---

## 3. 写像結果

共通の旧入力例:

```json
{
  "host_type": "cloud",
  "selfhosted_url": "",
  "email": "user@example.com",
  "folder_name": "team-envs",
  "backend": "api",
  "device_identifier": "dev-1"
}
```

#### テスト：正常系

| # | 旧入力の特徴 | 期待ホスト |
|---|--------------|------------|
| X1 | cloud + email + folder + device | `id=default`, `type=bitwarden-cloud`, `host_url=https://vault.bitwarden.com`, `email` あり, `target_section=team-envs`, `is_default=true`, `device_identifier=dev-1` |
| X2 | selfhosted + URL | `type=bitwarden-selfhost`, `host_url` が旧 URL |
| X3 | `folder_name` 空／省略 | `target_section=dotenvs` |
| X4 | `email` 空 | `email` キー省略または空扱い（ロード後に未設定） |
| X5 | `backend: "bw"` | 構造は X1 相当に変換。`backend` キー無し。警告ログ／メッセージに API 前提の旨 |
| X6 | 変換後 | トップに旧 `host_type` 等が残らない。`settings.save_files` は未設定（キー無しまたは空でフィルタ未適用） |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| X7 | selfhosted なのに URL 空など、写像後に validate 失敗 | エラー（壊れた新ファイルを正式パスに残さない） |

---

## 4. バックアップ

#### テスト：正常系

| # | 手順 | 期待 |
|---|------|------|
| B1 | 成功マイグレーション | バックアップファイルが存在し、内容が移行前の旧 JSON と一致 |
| B2 | 移行成功後の正式パス | 新スキーマの `config.jsonc` を読む |

---

## 5. 対象外

- Keychain キーの移行／読み替え → [`../infra/secretstore_hosts.md`](../infra/secretstore_hosts.md)（#153）
- プロジェクト `.bwsf` の `not_save_files` 自動変換（手動で `save_files` + `!` へ。残存はスキーマエラー）
- `bwsf init`（#193）

---

## 実装メモ（仕様外・参考）

```text
Load → detect flat
  → confirm (or --yes)
  → backup original
  → map §2.6 → Validate → Save config.jsonc
  → reject / no --yes → error, keep original
```
