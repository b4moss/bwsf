# core: api 保管庫接続時の push / pull / list（Issue #53 Step 4）

対象パッケージ: `app/src/core`（必要なら `app/src/cmd` の薄い配線）  
方針: CLI 挙動は backend 間で同一（Q25）。差分は `BwClient` 実装のみ。  
ユニットはモック中心（Q31）。実機 VW は PO 手動スモーク。

---

### PushEnvCore（api クライアント注入）

- 既存ロジックのまま `GetDotenvsFolderID` → `GetItemByName` → `CreateNoteItem` / `UpdateNoteItem` を使う。
- フォルダ欠落時は adapter のエラーを伝播し、**ここで folder を自動作成しない**（Q23）。
- 同名 Note 複数で `GetItemByName` がエラーなら push を中止（Q29）。
- 更新競合で `UpdateNoteItem` がエラーなら停止し、自動マージしない（Q33）。

#### テスト：正常系

- モック api クライアントで、新規 project のとき `CreateNoteItem` が呼ばれ成功する（既存 push 正常系と同等）。
- 既存 Note があるとき `UpdateNoteItem` が呼ばれ成功する。

#### テスト: 異常系

- `GetDotenvsFolderID` が not found を返した場合、`CreateDotenvsFolder` を呼ばずエラー終了する。
- `GetItemByName` が同名複数エラーを返した場合、Create/Update を呼ばずエラー終了する。
- `UpdateNoteItem` が競合エラーを返した場合、成功扱いにせずエラー終了する。
- `ErrAPINotImplemented` は想定しない（Step 4 後）。もし返ったらそのまま伝播（再試行しない）。

---

### PullEnvCore（api クライアント注入）

- 既存どおり project 名の Secure Note を取得し MultiEnvData / 旧形式を復元する（Q24）。

#### テスト：正常系

- モックが返す MultiEnvData JSON をファイルに復元できる（既存正常系と同等）。
- 旧形式 JSON も復元できる（既存下位互換テストと同等）。

#### テスト: 異常系

- アイテム無し（`nil`）のとき既存どおり not found エラー。
- 同名複数エラーは pull を中止する。
- フォルダ欠落エラーを伝播する（自動作成しない）。

---

### ListDotenvsCore（api クライアント注入）

- `ListItemsInFolder` の結果を返す。Secure Note 以外が混ざらないことは adapter 側仕様（infra テスト）。core は一覧をそのまま扱う。

#### テスト：正常系

- モックが返す Item 名一覧を取得できる（既存と同等）。

#### テスト: 異常系

- フォルダ欠落・同名 folder 曖昧・未 unlock 等のエラーを伝播する。
- 空一覧はエラーにしない（既存どおり）。

---

### WithAuthAndUnlockRetry との組み合わせ（退行）

- api で vault が実装されたあと、未 unlock エラーは従来どおり MP → Unlock → 再実行。
- 認証切れは `bwsf auth login` 案内のまま（Step 3 / #174）。
- vault 業務エラー（not found / 競合 / 同名複数）で Unlock 再試行ループに入らない。

#### テスト：正常系

- 未 unlock → Unlock 成功 → push/list 成功（Step 3 系と同等、vault モック成功）。

#### テスト: 異常系

- `GetItemByName` の同名複数エラーでは Unlock を繰り返さない。
- フォルダ not found では Unlock 再試行に入らない（認証・鍵の問題ではない）。
