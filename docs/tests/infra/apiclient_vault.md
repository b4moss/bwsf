# infra: ApiBwClient 保管庫操作（Issue #53 Step 4）

対象パッケージ: `app/src/infra`  
前提: Step 2 認証 + Step 3 Unlock 済み。Community SDK は adapter の裏に閉じ込める（Q22）。  
対象外: 組織ボルト、添付・カスタムフィールドの編集、`clean`（Q26 / Q37 / Q32）。

共通前提（各メソッド）:

- 未認証なら `ErrAPINotAuthenticated`（または同等）。
- 認証済み・未 unlock なら `ErrAPINotUnlocked`（または同等）。`ErrAPINotImplemented` は返さない。
- 成功パスでは事前に sync 相当を行う（push/pull/list および folder 操作の入口。Q36 / 既存合意）。
- API Key / MP / notes 本文 / 鍵材料をログに出さない（Q35）。
- 一時的なネットワークエラーはリトライせず失敗を返す（Q34）。

---

### Sync（または Vault 操作前の sync 相当）

- サーバ上の個人ボルト状態をクライアント側に取り込む。
- folder / cipher 操作の前に呼ばれる（明示 API でも、各操作の内部でもよい。テストでは「操作前に 1 回以上 sync される」を固定）。

#### テスト：正常系

- unlocked 状態で sync が成功し、以降の folder/cipher 読み取りが可能になる。
- 連続する vault 操作のたびに sync が行われる（または各公開メソッド入口で sync が呼ばれることをモックで検証）。

#### テスト: 異常系

- 未 unlock で sync 相当を要求すると `ErrAPINotUnlocked`。
- sync がネットワーク／SDK エラーを返した場合、そのまま（またはラップして）返し、リトライしない。

---

### GetDotenvsFolderID

- 設定フォルダ名（`folder_name`、未設定時は `dotenvs`）に一致する **個人ボルト上の folder ID** を返す。
- フォルダが存在しない場合はエラー（自動作成しない。Q23）。

#### テスト：正常系

- 設定名と一致する folder が 1 件あるとき、その ID を返す。
- `folder_name` 未設定時は `dotenvs` を探す。

#### テスト: 異常系

- 一致する folder が無い場合、明確な「not found」系エラー（自動作成しない）。
- 同名 folder が複数ある場合、エラーで中止する（曖昧な ID を返さない）。
- soft-deleted のみ存在する同名 folder は「無い」扱い（アクティブのみ。Q30）。

---

### DotenvsFolderExists

- 設定フォルダ名の folder がアクティブに存在するかを bool で返す。

#### テスト：正常系

- 存在するとき `true, nil`。
- 存在しないとき `false, nil`。

#### テスト: 異常系

- sync / SDK 失敗時は `false` ではなくエラーを返す（存在判定不能を false に落とさない）。
- soft-deleted のみの場合は `false, nil`。

---

### CreateDotenvsFolder

- 設定フォルダ名で個人ボルトに folder を作成する。
- setup からの作成に使う（Q27）。業務の push/pull/list からは自動呼び出ししない（Q23）。

#### テスト：正常系

- 未存在のとき作成に成功し、以降 `DotenvsFolderExists` が true / `GetDotenvsFolderID` が成功する。

#### テスト: 異常系

- 既に同名アクティブ folder がある場合、エラー（二重作成しない）または既存仕様で安全に失敗する。
- 未 unlock では作成せず `ErrAPINotUnlocked`。

---

### ListItemsInFolder

- 指定 folder 内のアイテム一覧を返す。
- **Secure Note のみ**（Q28）。他タイプは含めない。
- soft-deleted は無視（Q30）。

#### テスト：正常系

- Secure Note のみが `[]core.Item`（ID / Name）として返る。
- 空 folder なら空スライスと `nil`。

#### テスト: 異常系

- Login アイテム等だけがある folder では空（または Secure Note 以外を除外した結果）であり、他タイプを Name 一覧に混ぜない。
- soft-deleted Secure Note を一覧に含めない。
- 未 unlock / 不正 folderID はエラー。

---

### GetItemByName

- folder 内で名前が一致する Secure Note を 1 件返す（notes 平文を含む `FullItem`）。
- 同名が複数アクティブである場合は **エラーで中止**（Q29）。0 件なら `nil, nil`（既存 bw 経路の「無い」表現に合わせる。実装がエラー型なら文書化しテストで固定）。

#### テスト：正常系

- 一意な名前の Secure Note を復号して返す（Name / Notes / ID）。
- 存在しない名前は「無し」（`nil` アイテム）として扱う（既存 `PushEnvCore` / `PullEnvCore` 前提に合わせる）。

#### テスト: 異常系

- 同名アクティブが 2 件以上ならエラー（どちらかを勝手に選ばない）。
- soft-deleted の同名だけがある場合は「無し」またはアクティブ優先の仕様をテストで固定（アクティブが無ければ無し）。
- 復号失敗・未 unlock はエラー。

---

### GetItemByID

- ID 指定で Secure Note を取得・復号する。
- soft-deleted / 別タイプはエラーまたは not found（仕様をテストで固定。推奨: アクティブな Secure Note 以外は失敗）。

#### テスト：正常系

- 有効な Secure Note ID で FullItem を返す。

#### テスト: 異常系

- 存在しない ID、または Secure Note 以外の ID はエラー／not found。
- soft-deleted ID は取得しない。

---

### CreateNoteItem

- folder 内に Secure Note を新規作成する（name / notes）。
- notes は MultiEnvData JSON 文字列をそのまま格納（暗号化は SDK 側。Q24）。

#### テスト：正常系

- 作成成功後、`GetItemByName` で同内容を読める。
- 既存同名が無いときに呼ばれる（呼び出し側契約。アダプタ単体では作成成功を確認）。

#### テスト: 異常系

- 未 unlock では作成しない。
- SDK / ネットワーク失敗はリトライせずエラー。
- （任意）同一 name が既にある状態で Create した場合の SDK 依存挙動を、実装が選ぶならエラーとしてテスト固定。

---

### UpdateNoteItem

- 既存 Secure Note の notes を更新する。
- 添付・カスタムフィールド等は触らず壊さない（Q37。モックで「notes 以外のフィールドを保持した更新ペイロード」を検証できるなら行う）。

#### テスト：正常系

- 指定 ID の notes が新内容に更新され、再取得で一致する。

#### テスト: 異常系

- 存在しない ID はエラー。
- push 競合（サーバ側 revision 不一致等）を検出できる場合はエラー停止し、自動マージしない（Q33）。検出不能なら「失敗を握りつぶさない」ことを最低限テストする。
- 未 unlock では更新しない。

---

### ErrAPINotImplemented の撤去

- Step 4 完了後、上記メソッドは `ErrAPINotImplemented` を返さない。

#### テスト：正常系

- unlocked モック vault 上で各メソッドが実装済み応答を返す。

#### テスト: 異常系

- （退行）未実装のままのメソッドが残っていないこと（コンパイル／明示テスト一覧で担保）。
