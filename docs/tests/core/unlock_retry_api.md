# core: 認証切れ / 未 unlock 再試行（Issue #53 Step 3）

対象パッケージ: `app/src/core`  
既存の `WithUnlockRetry` / `IsLockedError` を、api backend でも使えるよう拡張する（bw 経路の退行なし）。

---

### IsNotAuthenticatedError

- エラーが「API / セッション未認証」（`auth` が必要）を表すかを判定する。
- CLI の master-password ロック文言だけに依存しない。

#### テスト：正常系

- `ErrAPINotAuthenticated`（またはラップ）を渡すと true。
- 無関係な I/O エラーでは false。

#### テスト: 異常系

- `nil` エラーでは false。
- 未 unlock エラー（鍵無し）を認証切れと誤判定しない（false）。

---

### IsNotUnlockedError

- エラーが「復号鍵が無い / unlock が必要」を表すかを判定する。
- bw のロック相当（既存 `IsLockedError`）と整合するか、またはここに集約する。

#### テスト：正常系

- 未 unlock を表すセンチネル／ラップエラーで true。
- 従来の Bitwarden CLI locked 文言エラーでも true（bw 経路互換）、または `IsLockedError` 経由で同等に扱える。

#### テスト: 異常系

- 認証切れエラーを未 unlock と誤判定しない。
- `ErrAPINotImplemented`（Step 4 stub）を未 unlock と誤判定しない。
- `nil` では false。

---

### WithAuthAndUnlockRetry

- 業務関数 `fn` を実行し、失敗理由に応じて回復を試みる。
- **認証切れ**: パスワードプロンプトでは回復しない。呼び出し元へ認証エラーを返す（cmd が `bwsf auth` を案内）。
- **未 unlock**: `promptPassword` → `bw.Unlock` → `fn` 再実行。
- bw 経路: 既存の Login→Unlock フォールバックを維持してよい（退行テストで固定）。

#### テスト：正常系

- `fn` が一度で成功する場合、プロンプトも Unlock も呼ばれない。
- `fn` が未 unlock エラーを返し、MP 入力 → Unlock 成功 → `fn` 再実行が成功する。
- bw 経路で従来どおり「locked → Unlock 成功 → 再実行」が通る（退行）。

#### テスト: 異常系

- `fn` が認証切れエラーを返す場合、`promptPassword` を呼ばず、認証エラーがそのまま（または案内可能な形で）返る。
- 未 unlock で `promptPassword` が失敗した場合、Unlock を呼ばずそのエラーを返す。
- Unlock が失敗した場合、`fn` を再実行せず Unlock 失敗を返す。
- `fn` が `ErrAPINotImplemented` を返す場合、再試行ループに入らずそのまま返す。
