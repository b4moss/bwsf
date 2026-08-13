# cmd: api セッション寿命（Issue #53 Step 3）

対象パッケージ: `app/src/cmd`  
合意: 明示の `lock` / `logout` は作らない。各コマンド終了時にメモリ上のトークン・鍵を破棄する。

---

### ensureClearedSessionOnExit（または各 run* の defer）

- `backend=api` のコマンド実行で `ApiBwClient`（または ClearSession 可能なクライアント）を使う場合、`defer` でセッションを破棄する。
- 成功時・失敗時の両方で破棄される。
- `backend=bw` では従来どおり（不要なら no-op）。

#### テスト：正常系

- api 経路で処理が成功終了したあと、クライアントが unlocked / authenticated でない（ClearSession 済み）。
- api 経路で処理がエラー終了したあとでも ClearSession が呼ばれている。
- `bwsf auth` 自体はトークン取得が目的のため、終了時破棄ポリシーを文書どおりに固定する（取得確認後に破棄してよい／プロセス終了で捨ててよい。実装に合わせテストする）。

#### テスト: 異常系

- ClearSession が失敗しても、元の業務エラーを隠さない（ラップまたはログのみ。パニックしない）。
- `backend=bw` で ClearSession 相当が誤って必須化され、bw 経路が壊れない。

---

### エラー表示分岐（認証切れ vs 未 unlock）

- 認証切れ: `bwsf auth` を案内し、MP プロンプトに進まない（または再試行ヘルパがプロンプトしない）。
- 未 unlock: MP プロンプトへ進む。
- Step 4 stub（`ErrAPINotImplemented`）: Step 3 では「未実装」として表示し、unlock 成功後でも CRUD は失敗してよい。

#### テスト：正常系

- 未 auth で list/push 相当の入口を叩くと、auth 案内が出る（exit non-zero）。
- auth 済み・未 unlock で入口を叩くと、MP プロンプト用の経路が呼ばれる（モックで検証）。

#### テスト: 異常系

- unlock 成功後に vault stub エラーになる場合、メッセージが認証案内や MP 再入力に誤誘導しない。
- 秘密情報（API Key / MP / notes）がエラー出力に含まれない。
