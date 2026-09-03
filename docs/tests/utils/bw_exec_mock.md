# utils: `bw` CLI 実行差し替え境界（Issue #160 Phase 2 / v0.18.0）

対象パッケージ: `app/src/utils`（`bwlist.go` / `bwpush.go` / `bwlogin.go`、および実行差し替えヘルパ）  
Issue: [#160](https://github.com/b4moss/bwsf/issues/160) — coverage 75%+ に向けた `bw` ラッパの単体テスト容易化

関連:

- 既存 `isBwAuthRequiredMessage` / `ErrBitwardenLocked` / session 抽出ヘルパ（`bwlist_test.go` / `bwlogin_test.go`）
- レイヤ方針: `docs/TEST.md`（CLI 薄いラッパ / core DI / infra・utils が `bw` 具体）
- Phase 1（cmd DI・input/spinner 等）とは独立。本仕様は **Phase 2 の exec 差し替え**のみを FIX する

---

### 前提（本仕様で固定）

| # | 方針 |
|---|------|
| B1 | `bwlist` / `bwpush` / `bwlogin` 内の `exec.Command("bw", ...)`（および必要なら `LookPath("bw")`）は、テスト差し替え可能な **単一の実行境界**（名称例: `runBw`）経由に統一する。本番デフォルトは現行どおり実 `bw` を起動する |
| B2 | 境界は **stdin 付き呼び出し**を表現できる（`create item` / `encode` / `edit` のパイプ経路）。stdout・stderr・error（非ゼロ exit）をテストが制御できること |
| B3 | 単体テストはこの境界を差し替え、**実 `bw` バイナリ・ネットワーク・Vault には依存しない** |
| B4 | 差し替えはパッケージ変数 + `t.Cleanup` で本番実装に戻す。`t.Parallel()` は使わない（`BW_SESSION` / spinner 等のグローバル副作用があるため） |
| B5 | 公開 API の契約（戻り値・エラー型・`IsLockedError` / auth 正規化）は現行を維持する。本仕様はテスト容易化のための内部境界であり、ユーザー向け CLI 振る舞いの変更ではない |
| B6 | テスト用モック（`app/src/testutil`）や実 `bw` を起動する e2e は本仕様の対象外。モック実装そのもののカバレッジ稼ぎはしない（モックは coverage 測定パッケージ外） |
| B7 | ユニットで固定する分岐は下表。対話 UI・TTY・Keychain（darwin）・実ログインの往復は e2e / 手動に委譲する |

---

### 実行境界（推奨形）

実装名は任意。次を満たせばよい。

```text
runBw(name string, args []string, stdin []byte) → (stdout, stderr []byte, err error)
```

- 本番: `name` は通常 `"bw"`。`LookPath` 失敗時は現行と同様に呼び出し元が扱う
- テスト: `args`（および必要なら `stdin`）を見て固定応答を返す
- `bw encode` → 続けて `bw create` / `bw edit` のように **連続呼び出し**がある経路は、呼び出し順・args を記録して検証してよい

---

### ユニットで固定する分岐

#### 共通（auth / ロック）

既存 `isBwAuthRequiredMessage` は維持。runner 経由の呼び出しでも、stdout/stderr に該当文言があれば現行どおり locked / auth エラーへ正規化する。

| # | 条件 | 期待 |
|---|------|------|
| A1 | stderr/stdout が Master password / not logged in / Vault is locked 等 | `IsLockedError` 相当、または現行の auth エラー経路 |
| A2 | それ以外の非ゼロ exit | 汎用エラー（メッセージに出力を含めてよい） |

#### `bwlist`（folders / items / unlock）

| # | 対象 | 条件 | 期待 |
|---|------|------|------|
| L1 | `GetFolderID` / `GetDotenvsFolderID` | `list folders` が有効 JSON（対象名あり） | 該当 folder ID |
| L2 | 同上 | 対象名なし | 現行どおり not found / 空扱い |
| L3 | 同上 | 不正 JSON または非ゼロ exit（A1 以外） | エラー |
| L4 | `CreateFolder` / `CreateDotenvsFolder` | `create folder` 成功（sync 失敗は無視可） | nil エラー |
| L5 | `ListItemsInFolder` | `list items` が有効 JSON | `[]Item` にパース |
| L6 | `ListItemsInFolder` | A1 の auth 文言 | locked / auth エラー |
| L7 | `BwUnlock` | unlock 成功（session 文字列） | 成功・session 設定の現行挙動 |
| L8 | `BwUnlock` | unlock 失敗 | 失敗の現行挙動 |

#### `bwpush`（get / create / update）

| # | 対象 | 条件 | 期待 |
|---|------|------|------|
| P1 | `GetItemByName` | sync 後 `list items` に名前一致 | `*FullItem` |
| P2 | `GetItemByName` | 名前不一致 | nil, nil または現行の not found |
| P3 | `GetItemByID` | `get item` 成功 JSON | `*FullItem` |
| P4 | `GetItemByID` | A1 / 非ゼロ | エラー |
| P5 | `CreateNoteItem` | direct create（stdin JSON）成功 | nil |
| P6 | `CreateNoteItem` | direct 失敗 → encode 経路成功 | nil（フォールバック） |
| P7 | `CreateNoteItem` | 両経路失敗 | エラー |
| P8 | `UpdateNoteItem` | get → encode/edit 成功 | nil |
| P9 | `UpdateNoteItem` | get または edit 失敗 | エラー |

#### `bwlogin`

| # | 対象 | 条件 | 期待 |
|---|------|------|------|
| G1 | `CheckBwCommand` | LookPath 相当が成功 | `(true, path)` |
| G2 | `CheckBwCommand` | LookPath 失敗 | `(false, "")` 相当 |
| G3 | `BwLogin` | 未ログイン → login 成功 → session 抽出可 | 成功 |
| G4 | `BwLogin` | already logged in → unlock 経路で session 取得 | 成功 |
| G5 | `BwLogin` | serverURL ありで `config server` が必要な経路 | 現行どおり config/logout 後に login |
| G6 | `BwLogin` | session 抽出不可 / 非ゼロ | 失敗 |

`environWithout` / `truncateForErr` / `looksLikeSessionKey` / `extractSessionKey` は Phase 1 または既存テストでよい（本仕様の必須ではない）。

---

### テスト：正常系（まとめ）

- runner 差し替えで L1 / L4 / L5 / L7 / P1 / P3 / P5 / P6 / P8 / G1 / G3 / G4 のうち、公開 API から到達するものをカバーする
- encode 経路（P6）は **stdin 付きの連続呼び出し**が少なくとも 1 ケースあること

### テスト: 異常系（まとめ）

- A1 / A2、L3 / L6、P4 / P7 / P9、G2 / G6
- 不正 JSON はパースエラーとして呼び出し元に返る（パニックしない）

---

### e2e / 手動に委譲するもの

- 実 Bitwarden / Vaultwarden への login・sync・CRUD
- Keychain（darwin `SessionStore`）
- promptui / TTY 対話
- `infra.RealBwClient` の薄い委譲そのもの（変換ロジックを別途薄く見る場合は本仕様外で可）
- coverage 数値そのもの（Issue #160 のチェックリストで管理。本仕様は境界とケースの正本）

---

### 完了の見方

- 上記ユニットケースが green
- 本番経路（runner 未差し替え）の退行がないこと（既存 unit / 必要なら e2e）
- Issue #160 Phase 2（75%+）達成の手段として使う。カバレッジゲート化はしない
