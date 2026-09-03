# infra: ApiBwClient Unlock（Issue #53 Step 3）

対象パッケージ: `app/src/infra`  
前提: Personal API Key 認証（Step 2）済み。vault CRUD は未実装のまま。

v0.20.0 §3（[#153](https://github.com/b4moss/bwsf/issues/153)）との関係:

- Keychain キーの host 単位化・`vault_unlock` → [`secretstore_hosts.md`](./secretstore_hosts.md)
- Export / Restore / 自動 restore → [`../core/vault_unlock_restore.md`](../core/vault_unlock_restore.md)
- 本文書は **プロセスメモリ上の Unlock / ClearSession / IsUnlocked** の契約を維持する

---

### Unlock

- マスターパスワードを受け取り、保管庫復号に必要な鍵をプロセスメモリ上に復元する。
- 未認証の場合は先に `Authenticate`（Keychain の API Key → Identity トークン）を試みる。
- 成功後、以降の vault 操作（Step 4）が使える「unlocked」状態になる。
- マスターパスワードは Keychain に保存しない（都度引数／プロンプト）。
- API Key / トークン / 鍵材料をログに出さない。
- （§3）Unlock 成功後、実装は当該 host の `vault_unlock` を Keychain に保存してよい（詳細は [`vault_unlock_restore.md`](../core/vault_unlock_restore.md)）。

#### テスト：正常系

- 認証済み（有効な access_token あり）で正しい MP を渡すと、`Unlock` が `nil` を返し、unlocked 状態になる。
- 未認証でも Keychain に API Key がある場合、`Unlock` 内で Authenticate 成功後に鍵復元まで完了する。
- `Unlock` 成功後、`ClearSession` 前であれば unlocked と判定できる（`IsUnlocked` 相当）。

#### テスト: 異常系

- Keychain に API Key が無く Authenticate できない場合、`ErrAPINotAuthenticated`（または同等）を返し、鍵は載らない。
- 認証は成功するが MP が不正な場合、unlock 失敗エラーを返し、鍵は載らない。
- Scenario C / SDK が鍵復元不能な状態をモックした場合、明確なエラーを返し `ErrAPIUnlockNotImplemented` のまま放置しない（実装後は実装エラーとして返す）。
- MP が空文字の場合、鍵復元を行わずエラーを返す。

---

### ClearSession

- メモリ上の access_token（および refresh_token）と復号鍵を破棄する。
- OS Keychain 上の Personal API Key は削除しない（それは `bwsf auth --clear` / 将来の `auth logout`）。
- （§3）OS Keychain 上の **`vault_unlock` も削除しない**（それは `bwsf lock` / 将来の `auth logout`）。
- 複数回呼んでも安全（冪等）。

#### テスト：正常系

- `Unlock` 成功後に `ClearSession` すると、トークン無効かつ unlocked でなくなる。
- 認証のみ（未 unlock）の状態で `ClearSession` すると、トークンが破棄される。
- 何も保持していない状態で `ClearSession` してもパニックせず `nil`（または何もしない）。
- （§3）`vault_unlock` 保存後に `ClearSession` しても、Keychain の `vault_unlock` が残る。

#### テスト: 異常系

- （副作用の検証）`ClearSession` 後に Keychain の API Key が残っていること（削除されていないこと）。
- `ClearSession` 後の vault 的操作入口が「未認証」または「未 unlock」として失敗すること（Step 3 では stub でも状態判定で確認可。§3 では別インスタンスからの restore 成功は [`vault_unlock_restore.md`](../core/vault_unlock_restore.md)）。

---

### IsUnlocked（または同等の状態照会）

- プロセスメモリ上に有効な復号鍵があるかを返す。
- トークンだけで鍵が無い場合は false。

#### テスト：正常系

- `Unlock` 成功直後は true。
- `ClearSession` 後は false。
- 認証のみ（未 unlock）は false。

#### テスト: 異常系

- （境界）期限切れトークンのみ保持し鍵が無い場合は false（認証と unlock を混同しない）。
