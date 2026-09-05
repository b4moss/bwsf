# config: JSONC 読み込み（Issue #155 / v0.18.0）

> **v0.20.0 / #177:** グローバルのパス XOR・Save 先 `.jsonc`・新スキーマは [`global_v2.md`](./global_v2.md) が優先。本文書は hujson 正規化の技術契約（コメント・末尾カンマ）として残す。

対象パッケージ: `app/src/config`（薄い JSONC 正規化ヘルパを同パッケージまたは共有 util に置く）  
Issue: [#155](https://github.com/b4moss/bwsf/issues/155) — 設定ファイル読み込みの JSONC 対応（`github.com/tailscale/hujson`）

関連:

- 既存 `LoadConfig` / `SaveConfig` / `Config`（`~/.config/bwsf/config.json`）
- 先送り: [#133](https://github.com/b4moss/bwsf/issues/133)（`.bwsf/config.(json|jsonc)` の探索・スキーマ）
- v0.20 グローバル: [`global_v2.md`](./global_v2.md)（[#177](https://github.com/b4moss/bwsf/issues/177)）

---

### 前提（本仕様で固定）

| # | 方針 |
|---|------|
| J1 | 読み込みパーサは **`github.com/tailscale/hujson`**。`Standardize` 後に `encoding/json` で `Config` へ Unmarshal する |
| J2 | **JSON / JSONC 両対応**（コメントなしの通常 JSON も同じ経路で読める） |
| J3 | **末尾カンマを許容**する（オブジェクト・配列） |
| J4 | 本 Issue の対象パスはグローバル **`~/.config/bwsf/config.json` のみ**（ファイル名は `.json` のまま。中身が JSONC でもよい）。`.jsonc` 拡張子の探索は #133 |
| J5 | **`SaveConfig` は strict JSON のまま**（`json.MarshalIndent`）。コメントや末尾カンマは書き出さない |
| J6 | Bitwarden Vault（Secure Note）の JSON 読み書きは **対象外**（現行スキーマ・`encoding/json` のまま） |
| J7 | パース失敗時は既存どおりエラーを返す（メッセージに `failed to parse config file` 相当を維持し、原因は `%w` で包む） |

---

### ヘルパ（推奨）

実装が `UnmarshalConfigJSONC(data []byte, dst *Config) error`（名称任意）のような純関数に寄せる場合、ユニットはここを厚くしてよい。

振る舞い:

1. hujson でパースし、strict JSON バイト列へ正規化する
2. `json.Unmarshal` で `Config` に載せる
3. 通常 JSON 入力でも成功する（退行なし）

#### テスト：正常系

| # | 入力の特徴 | 期待 |
|---|------------|------|
| H1 | 通常 JSON（コメントなし・末尾カンマなし） | `Config` 各フィールドが期待どおり |
| H2 | `//` 行コメント・ブロックコメントを含む JSONC | コメントは無視され、同値の `Config` |
| H3 | オブジェクト／配列の **末尾カンマ** | 成功し、同値の `Config` |
| H4 | コメントと末尾カンマの混在 | 成功し、同値の `Config` |

フィクスチャ例（`app/src/config/testdata/` 想定。ファイル名は実装時に合わせてよい）:

- `plain.json` — H1
- `comments.jsonc` — H2
- `trailing_comma.jsonc` — H3
- `comments_and_trailing_comma.jsonc` — H4

各フィクスチャの論理値は同一でよい（例: `host_type=cloud`, `email=user@example.com`, `folder_name` 未設定または明示）。

#### テスト: 異常系

| # | 入力 | 期待 |
|---|------|------|
| H5 | 壊れたトークン・閉じ忘れ等 | 非 nil エラー（`Config` は使わない） |
| H6 | 空バイト / 空白のみ | エラー（現行 `json.Unmarshal` と同様に失敗扱い） |

---

### `LoadConfig`

- ファイルパスは従来どおり `GetConfigPath()` → `~/.config/bwsf/config.json`（テストは HOME 差し替え）
- ファイル無し → `nil, nil`（現行どおり）
- ファイル有り → 上記ヘルパ経由でパース（**`json.Unmarshal` 直呼びはしない**）

#### テスト：正常系

- 既存: 通常 JSON の `config.json` を読める（退行: `TestLoadConfig_Success` 相当）
- 追加: HOME 下の `config.json` に JSONC フィクスチャ相当の内容を置き、`LoadConfig` が期待の `Config` を返す（コメント付き・末尾カンマ付きの少なくとも一方）

#### テスト: 異常系

- 既存: 壊れた内容 → パースエラー（`TestLoadConfig_InvalidJSON` 相当。JSONC 経路でも同様に失敗）
- ファイル無し → `nil, nil`（退行）

---

### `SaveConfig`（退行・書き出し契約）

本 Issue で書き出し形式は変えない。

#### テスト：正常系

| # | 手順 | 期待 |
|---|------|------|
| S1 | 通常の `Config` を `SaveConfig` | 出力ファイルは **strict JSON**（`json.MarshalIndent` 相当）。`//` や末尾カンマを含まない |
| S2 | JSONC から `LoadConfig` した `Config` を `SaveConfig` し、再度 `LoadConfig` | 再読込値が一致し、保存ファイルは strict JSON |

#### テスト: 異常系

- （退行）既存のディレクトリ作成・上書き成功系を壊さない

---

### 対象外（本仕様でテストしない）

- `.bwsf/config.json` / `.jsonc` の探索・選択・スキーマ（#133）
- Secure Note 本文の JSONC 化
- `bwsf config show` の表示文言変更（読み込みが通れば既存表示で足りる）

---

### 実装メモ（仕様外・参考）

```text
bytes → hujson.Parse / Standardize → encoding/json.Unmarshal → Config
SaveConfig → encoding/json.MarshalIndent → config.json
```
