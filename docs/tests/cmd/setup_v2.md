# cmd: `bwsf setup` v2（Issue #177 / v0.20.0 §2.5）

対象パッケージ: `app/src/cmd` / `app/src/core`（Setup） / 必要なら `app/src/utils`（対話）  
Issue: [#177](https://github.com/b4moss/bwsf/issues/177)  
製品正本: [`../specs/v0.20.0-multi-host.md`](../specs/v0.20.0-multi-host.md) §2.5

関連:

- 保存スキーマ → [`../config/global_v2.md`](../config/global_v2.md)
- 旧 api setup → [`setup_api.md`](./setup_api.md)（**v0.20 では本仕様が優先**）
- フォルダ ensure → [`setup_api_folder.md`](./setup_api_folder.md)（対象 host の `target_section` を使うよう追随）

---

### 前提（本仕様で固定）

| # | 方針 |
|---|------|
| S1 | **API のみ**。`backend=bw` の Login 付き setup 経路は廃止。Personal API Key は setup では扱わず、案内は `bwsf auth login`（[`auth_login_logout.md`](./auth_login_logout.md)） |
| S2 | host は対話で **「追加する / スキップ」** を選べる。スキップ可 → `hosts: []` のまま保存してよい |
| S3 | 初回追加（既存 `hosts` 空）の初期値: `id: "default"`, `is_default: true`。cloud なら `host_url: "https://vault.bitwarden.com"` |
| S4 | 既に host があるとき: 対話で **「追加 / 既存更新 / デフォルト変更 / スキップ」** |
| S5 | ファイル選択: 対話で **「`save_files` を設定 / 未設定」**。設定時は glob 入力（否定は `!` 接頭辞）。host をスキップしてもファイル選択は可 |
| S6 | 保存は常に新スキーマの `config.jsonc`（[`global_v2.md`](../config/global_v2.md)） |
| S7 | Bitwarden への `Login`（email/password）は行わない |
| S8 | フォルダ ensure（認証済み時）は解決／対象 host の `target_section` を使う。未認証ならスキップし auth を案内（現行 [`setup_api_folder.md`](./setup_api_folder.md) の精神を維持） |
| S9 | `--yes` は確認系（フォルダ作成等）に使う。マイグレーション確認との併用は [`migrate_v2.md`](../config/migrate_v2.md) |
| S10 | 非対話でも host スキップや最小保存ができること（フラグ名は実装で固定し本仕様の表を更新してよい。少なくとも「host を追加せず `hosts: []` + 任意 save_files」をテスト可能にする） |

---

## 1. host スキップ／初回追加

対話関数はスタブ可能。

#### テスト：正常系

| # | 操作 | 期待 |
|---|------|------|
| H1 | host「スキップ」+ save_files「未設定」 | `hosts: []`。`save_files` 未設定。Login なし。auth 案内あり |
| H2 | host「スキップ」+ save_files に `[".env*", "!.env.local"]` | `hosts: []` かつ当該 `save_files` |
| H3 | host「追加」・cloud・email 入力 | `hosts[0].id=default`, `type=bitwarden-cloud`, `host_url=https://vault.bitwarden.com`, `is_default=true`, `target_section` は default `dotenvs` または入力値 |
| H4 | host「追加」・selfhost・URL 必須 | `type=bitwarden-selfhost`, `host_url` が入力値 |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| H5 | selfhost で URL 空 | 保存せずエラー |
| H6 | 設定保存失敗 | エラー。Login を呼ばない |

---

## 2. 既存 host があるとき

初期状態: `default`（is_default）が1件。

#### テスト：正常系

| # | 選択 | 期待 |
|---|------|------|
| E1 | スキップ | hosts 変更なし（save_files だけ変えられる） |
| E2 | 追加 | 新 id が入り、`is_default` は全体でちょうど1のまま（新規を default にする／しないは対話結果に従い validate を満たす） |
| E3 | 既存更新 | 選んだ host の email / url / target_section 等が更新される |
| E4 | デフォルト変更 | 指定 id のみ `is_default: true` |

---

## 3. 非対話

#### テスト：正常系

| # | 条件 | 期待 |
|---|------|------|
| N1 | host 追加に必要なフラグ一式（type / url 要時 / email 等） | Login 無しで新スキーマ保存 |
| N2 | host スキップを明示する非対話経路 | `hosts: []` で成功しうる |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| N3 | host 追加モードなのに必須フラグ不足 | エラー |

（具体フラグ名は実装時に本表へ追記して FIX する。）

---

## 4. 退行・廃止

| # | 期待 |
|---|------|
| X1 | `backend=bw` 専用の password Login フローが setup から呼ばれない |
| X2 | 完了メッセージに認証は別コマンドである旨（`bwsf auth login`） |

---

## 5. 対象外

- `auth login` / `logout` 本体（#174 → [`auth_login_logout.md`](./auth_login_logout.md)）
- Keychain 書き込み（setup は設定ファイルのみ）
- `bwsf init`（#193）

---

## 実装メモ（仕様外・参考）

```text
setup
  → (migrate if flat — migrate_v2)
  → prompt host: add | skip | (if existing) add/update/default/skip
  → prompt save_files: set | unset
  → Save GlobalConfig jsonc
  → optional EnsureFolder on target_section if authenticated
  → hint auth
```
