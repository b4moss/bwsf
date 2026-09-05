# cmd: api セッション寿命（Issue #53 Step 3 → Issue #153 / v0.20.0 §3）

対象パッケージ: `app/src/cmd`  
製品正本（v0.20）: [`../specs/v0.20.0-multi-host.md`](../specs/v0.20.0-multi-host.md) §3

関連:

- Keychain / `vault_unlock` → [`../infra/secretstore_hosts.md`](../infra/secretstore_hosts.md)
- 自動 restore → [`../core/vault_unlock_restore.md`](../core/vault_unlock_restore.md)
- CLI unlock / lock → [`unlock_lock.md`](./unlock_lock.md)
- メモリ Unlock API → [`../infra/apiclient_unlock.md`](../infra/apiclient_unlock.md)

---

### 前提（本仕様で固定）

| # | 方針 |
|---|------|
| L1 | 各 vault 系コマンド終了時、**プロセスメモリ上**の Identity トークン・復号鍵は `ClearSession`（または同等）で破棄する |
| L2 | Keychain の **Personal API Key** と **`vault_unlock`** はコマンド終了では削除しない（削除は `lock` / [`auth logout`](./auth_login_logout.md)） |
| L3 | 明示の **`bwsf lock`** / **`auth logout`** が Keychain 側セッションの寿命を制御する（旧「明示 lock は作らない」合意は **§3 で破棄**） |
| L4 | 次回の push/pull/list/clean は `vault_unlock` があれば restore し、無ければ MP プロンプトへ（[`vault_unlock_restore.md`](../core/vault_unlock_restore.md)） |

---

### ensureClearedSessionOnExit（または各 run* の defer）

- API クライアントを使うコマンド実行で、`defer` により **メモリ**セッションを破棄する。
- 成功時・失敗時の両方で破棄される。
- Keychain の `vault_unlock` / API Key は残る。

#### テスト：正常系

- api 経路で処理が成功終了したあと、クライアントが unlocked / authenticated でない（ClearSession 済み）。
- 同時に、事前に保存していた `hosts/<id>/vault_unlock` と API Key が Keychain 上に残っている。
- api 経路で処理がエラー終了したあとでも ClearSession が呼ばれている（Keychain は業務失敗だけでは消さない）。
- `bwsf auth login` はトークン取得＋unlock まで進むが、終了時はメモリ破棄してよい（Keychain の API Key / `vault_unlock` は残す）。

#### テスト: 異常系

- ClearSession が失敗しても、元の業務エラーを隠さない（ラップまたはログのみ。パニックしない）。

---

### エラー表示分岐（認証切れ vs 未 unlock）

- 認証切れ: `bwsf auth login` を案内し、MP プロンプトに進まない（または再試行ヘルパがプロンプトしない）。
- 未 unlock かつ有効な `vault_unlock` 無し: MP プロンプトへ進む。
- 無効な `vault_unlock`: 破棄してから MP プロンプト（[`vault_unlock_restore.md`](../core/vault_unlock_restore.md)）。

#### テスト：正常系

- 未 auth で list/push 相当の入口を叩くと、auth 案内が出る（exit non-zero）。
- auth 済み・`vault_unlock` 無しで入口を叩くと、MP プロンプト用の経路が呼ばれる（モックで検証）。
- auth 済み・有効 `vault_unlock` ありなら MP プロンプト無しで進む。

#### テスト: 異常系

- unlock / restore 成功後に vault エラーになる場合、メッセージが認証案内や MP 再入力に誤誘導しない（無効セッション判定のときだけ再プロンプト）。
- 秘密情報（API Key / MP / `vault_unlock` / notes）がエラー出力に含まれない。

---

## 対象外

- `auth login` / `logout` 本体（#174 → [`auth_login_logout.md`](./auth_login_logout.md)）
- `bwsf init`（#193）
