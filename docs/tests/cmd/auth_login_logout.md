# cmd: `bwsf auth login` / `bwsf auth logout`（Issue #174 / v0.20.0 §4）

対象パッケージ: `app/src/cmd`（必要なら `infra`）  
Issue: [#174](https://github.com/b4moss/bwsf/issues/174)  
製品正本: [`../specs/v0.20.0-multi-host.md`](../specs/v0.20.0-multi-host.md) §4、host 解決は §1.1

関連:

- host 解決 → [`../config/host_resolve.md`](../config/host_resolve.md)（本仕様で `auth login` / `auth logout` の `--host` を追加）
- Keychain キー → [`../infra/secretstore_hosts.md`](../infra/secretstore_hosts.md)
- vault セッション CLI → [`unlock_lock.md`](./unlock_lock.md)
- プロセス終了時のメモリ破棄 → [`session_lifecycle.md`](./session_lifecycle.md)
- 自動 restore → [`../core/vault_unlock_restore.md`](../core/vault_unlock_restore.md)

---

### 前提（本仕様で固定）

| # | 方針 |
|---|------|
| A1 | コマンドは **`bwsf auth login`** / **`bwsf auth logout`**。親 `bwsf auth`（引数なし）は **ヘルプのみ**（認証・削除を実行しない） |
| A2 | 旧フラット `bwsf auth`（実行して認証）と **`--clear` は削除** |
| A3 | host 解決は §1.1: **CLI `--host`** → **プロジェクト `host`** → **`is_default`**（[`ResolveHost`](../config/host_resolve.md) を unlock/lock と共有） |
| A4 | `login` / `logout` はいずれも `--host` を受け付ける。`logout` に **`--all`** あり |
| A5 | `login` の責務: Personal API Key の取得・保存 → Identity 確認 → **unlock まで一気通貫**（`vault_unlock` も書く）。MP 自体は保存しない |
| A6 | `logout` の責務: 解決 host の **API Key + `vault_unlock` を削除**。`lock` は vault のみ（[`unlock_lock.md`](./unlock_lock.md)） |
| A7 | `logout --all` は登録済み **全 host** の API Key + `vault_unlock` を削除。`hosts` が空なら **no-op 成功** |
| A8 | auth 経路では host を対話追加しない。未解決ならエラー（先に `bwsf setup`） |
| A9 | ヘルプ文で境界を明示: **auth login/logout = API Key**（logout はセッションも）、**unlock/lock = vault セッションのみ** |
| A10 | 複数 host への同時 login 可（Keychain が host 単位）。対話・Identity・MP 入力はスタブ可能 |

確定フラグ: `login` / `logout` の `--host`、`logout` の `--all`。`--all` と `--host` の併用は **エラー**（[`unlock_lock.md`](./unlock_lock.md) L6 と同型）。

---

## 1. コマンド登録・破壊的変更

#### テスト：正常系

| # | 操作 | 期待 |
|---|------|------|
| R1 | root のサブコマンド一覧 | `auth` が含まれる |
| R2 | `auth` のサブコマンド一覧 | `login` と `logout` が含まれる |
| R3 | 引数なし `bwsf auth` / `bwsf auth --help` | ヘルプ表示のみ。Keychain・Identity を触らない。exit 0 |
| R4 | `login --help` / `logout --help` | `--host` が文書化される。`logout` に `--all` がある |
| R5 | 親または子の Long / ヘルプ | API Key（auth）と vault セッション（unlock/lock）の境界が読める |

#### テスト: 異常系 / 退行

| # | 条件 | 期待 |
|---|------|------|
| R6 | `--clear` フラグ | 親・子いずれも存在しない（unknown flag またはヘルプに無い） |
| R7 | 旧フラット実行相当（サブコマンド無しで認証する経路） | 存在しない（R3 のヘルプのみ） |

---

## 2. `bwsf auth login`

対話の API Key / 再利用確認 / MP 入力はスタブ可能。初期ホスト: グローバルに `default`（is_default）と `work`。

#### テスト：正常系

| # | 条件 | 期待 |
|---|------|------|
| N1 | default host 解決・新規 API Key・正しい MP | 成功。`hosts/<id>/api_client_*` と `hosts/<id>/vault_unlock` が非空。プロセスメモリは終了時 ClearSession してよい |
| N2 | `--host work` | `work` のみ更新。`default` の API Key / `vault_unlock` は不変 |
| N3 | プロジェクト `.bwsf` に `host: work`、CLI 無し | `work` を login（Key + unlock） |
| N4 | Keychain に既存 Key あり・再利用 Yes | 秘密の再入力なしで Identity → unlock まで進む |
| N5 | 既存 Key あり・再利用 No → 新規入力 | 新しい Key で上書き保存され、unlock まで進む |
| N6 | 既に `vault_unlock` がある状態で再 login | 新しい不透明データで上書き（成功） |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| E1 | host 解決失敗（未知 id / hosts 空で default 無し） | 非ゼロ終了。Keychain を変更しない |
| E2 | Identity 認証失敗 | 非ゼロ。`vault_unlock` を新たに書かない（実装が「失敗時は unlock に進まない」を固定）。API Key の部分保存有無は実装で固定しテストに書く |
| E3 | Identity 成功・MP 不正 | エラー。API Key は残してよい。既存の有効 `vault_unlock` は誤って消さない（[`unlock_lock.md`](./unlock_lock.md) N7 と同型） |
| E4 | Identity 成功・MP 空 | エラー。Unlock しない |
| E5 | Unlock に必要な host 情報不足（例: email 未設定で失敗する実装） | 明確なエラー。setup 案内可 |

実装固定（本仕様）:

- E2: Identity 失敗時は unlock しない。API Key を Identity 前に書いた場合でも `vault_unlock` は書かない。
- E3/E4: login の unlock 部分は unlock コマンドと同じ失敗時ポリシー（既存 blob 維持）。

---

## 3. `bwsf auth logout`

#### テスト：正常系

| # | 条件 | 期待 |
|---|------|------|
| L1 | 解決 host に API Key と `vault_unlock` あり | 両方削除 |
| L2 | API Key のみ / `vault_unlock` のみ / どちらも無し | 成功（冪等）。欠けるキーをエラーにしない |
| L3 | `--host work` | `work` のみ削除。`default` の API Key / `vault_unlock` は残る |
| L4 | `logout --all`・hosts に `default` と `work` | 両方の API Key と `vault_unlock` が削除 |
| L5 | `logout --all`・`hosts: []` | **no-op 成功**（exit 0） |
| L6 | `logout --all` と `--host` の併用 | **エラー**（どちらも消さない） |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| L7 | `--host` 未知 id（`--all` 無し） | エラー。他 host のキーを消さない |

---

## 4. unlock/lock との境界（退行）

#### テスト：正常系

| # | 操作 | 期待 |
|---|------|------|
| B1 | `lock` 後 | `api_client_*` は残る（logout 相当にしない） |
| B2 | `logout` 後に `unlock` | API Key 無しとして **`auth login` を案内**。`vault_unlock` も無い |
| B3 | `login` 成功後 | 別途 `unlock` 無しでも当該 host の `vault_unlock` がある（一気通貫） |
| B4 | 未 login で `unlock` | `auth login` 案内（旧 `bwsf auth` 文言にしない） |

---

## 対象外

- `bwsf unlock` / `lock` 本体の契約再定義（[`unlock_lock.md`](./unlock_lock.md)）
- Keychain キー名・flat 移行（[`../infra/secretstore_hosts.md`](../infra/secretstore_hosts.md)）
- push/pull/list/clean の自動 restore（[`../core/vault_unlock_restore.md`](../core/vault_unlock_restore.md)）
- グローバル `hosts[]` の追加・更新（[`setup_v2.md`](./setup_v2.md)）
- `bwsf init`（#193）
