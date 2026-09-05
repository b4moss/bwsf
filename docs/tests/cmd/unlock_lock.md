# cmd: `bwsf unlock` / `bwsf lock`（Issue #153 / v0.20.0 §3）

対象パッケージ: `app/src/cmd`（必要なら `infra` / `core`）  
Issue: [#153](https://github.com/b4moss/bwsf/issues/153)  
製品正本: [`../specs/v0.20.0-multi-host.md`](../specs/v0.20.0-multi-host.md) §3、host 解決は §1.1

関連:

- host 解決 → [`../config/host_resolve.md`](../config/host_resolve.md)（本仕様で unlock/lock の `--host` を追加）
- Keychain キー → [`../infra/secretstore_hosts.md`](../infra/secretstore_hosts.md)
- 自動 restore → [`../core/vault_unlock_restore.md`](../core/vault_unlock_restore.md)
- プロセス終了時のメモリ破棄 → [`session_lifecycle.md`](./session_lifecycle.md)

---

### 前提（本仕様で固定）

| # | 方針 |
|---|------|
| U1 | コマンド **`bwsf unlock`** / **`bwsf lock`** を登録する |
| U2 | host 解決は §1.1: **CLI `--host`** → **プロジェクト `host`** → **`is_default`**（[`ResolveHost`](../config/host_resolve.md) を共有） |
| U3 | `unlock` / `lock` はいずれも `--host` を受け付ける |
| U4 | `unlock` の責務は **vault セッションのみ**（API Key の取得・削除はしない。それは [`auth_login_logout.md`](./auth_login_logout.md)） |
| U5 | `unlock` 成功時: MP で Unlock → Keychain に当該 host の `vault_unlock` を保存。MP 自体は保存しない |
| U6 | `lock` は解決した **1 host** の `vault_unlock` のみ削除。API Key は残す |
| U7 | `lock --all` は登録済み **全 host** の `vault_unlock` を削除。`hosts` が空なら **no-op 成功** |
| U8 | 未認証（API Key 無し）で `unlock` した場合は auth 案内エラー（MP に進まない、または Unlock 前に失敗） |
| U9 | ヘルプ文で auth（API Key）と unlock/lock（vault セッション）の境界を分かる範囲で明示する（login/logout 文言の詳細は [`auth_login_logout.md`](./auth_login_logout.md)） |

---

## 1. コマンド登録

#### テスト：正常系

| # | 操作 | 期待 |
|---|------|------|
| R1 | root のサブコマンド一覧 | `unlock` と `lock` が含まれる |
| R2 | `unlock --help` / `lock --help` | `--host` が文書化される。`lock` に `--all` がある |

---

## 2. `bwsf unlock`

対話の MP 入力はスタブ可能。

#### テスト：正常系

| # | 条件 | 期待 |
|---|------|------|
| N1 | default host 解決・API Key あり・正しい MP | 成功。`hosts/<id>/vault_unlock` が非空で保存される。プロセスメモリは終了時 ClearSession してよい |
| N2 | `--host work` | `work` の `vault_unlock` のみ更新。他 host の `vault_unlock` は不変 |
| N3 | プロジェクト `.bwsf` に `host: work`、CLI 無し | `work` を unlock |
| N4 | 既に有効な `vault_unlock` がある状態で再 unlock | 新しい不透明データで上書き（成功） |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| N5 | host 解決失敗（未知 id / hosts 空で default 無し） | 非ゼロ終了。Keychain を変更しない |
| N6 | API Key 無し | auth 案内。`vault_unlock` を書かない |
| N7 | MP 不正 | エラー。既存の有効 `vault_unlock` を誤って消さない（失敗時は維持）。実装が「失敗時は触らない」をテストで固定 |
| N8 | MP 空 | エラー。Unlock しない |

---

## 3. `bwsf lock`

#### テスト：正常系

| # | 条件 | 期待 |
|---|------|------|
| L1 | 解決 host に `vault_unlock` あり | 当該キーのみ削除。API Key は残る |
| L2 | 解決 host に `vault_unlock` 無し | 成功（冪等 no-op） |
| L3 | `--host work` | `work` のみ削除。`default` の `vault_unlock` は残る |
| L4 | `lock --all`・hosts に `default` と `work` | 両方の `vault_unlock` が削除。API Key は残る |
| L5 | `lock --all`・`hosts: []` | **no-op 成功**（exit 0） |
| L6 | `lock --all` と `--host` の併用 | 実装で拒否するか `--all` 優先かを固定。テストで一方に決める（推奨: 併用はエラー） |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| L7 | `--host` 未知 id（`--all` 無し） | エラー。他 host のキーを消さない |

---

## 4. auth との境界（退行）

#### テスト：正常系

| # | 操作 | 期待 |
|---|------|------|
| B1 | `lock` 後も `hosts/<id>/api_client_*` が残る | API Key はログアウト相当にしない |
| B2 | `unlock` は API Key を新規要求しない（Keychain にあれば） | 資格情報プロンプトを出さない |

---

## 対象外

- `auth login` / `auth logout` / `logout --all` / 旧 flat `auth`・`--clear` 削除（#174 → [`auth_login_logout.md`](./auth_login_logout.md)）
- `bwsf init`（#193）
- push/pull/list/clean 内の自動 restore 詳細（[`vault_unlock_restore.md`](../core/vault_unlock_restore.md)）
