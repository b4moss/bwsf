# テスト仕様書

このドキュメントは、`bwsf` CLI のリファクタ後の構成と、それぞれの関数／メソッドの処理内容をテスト仕様書としてまとめたものです。

---

## 1. 全体構成

### main / CLI エントリポイント

- `main`
  - `cmd.Execute()` を呼び出し、Cobra ベースの CLI を起動する。
- `Execute`
  - `rootCmd.Execute()` を実行して CLI コマンド群を起動する。
  - `rootCmd.Execute()` がエラーを返した場合は `Logger` を通じてエラーメッセージを表示する。
  - エラー発生時には `os.Exit(1)` でプロセスを終了する。

#### テストシナリオ

- 正常系
  - `rootCmd.Execute()` が `nil` を返す場合に、`Logger` がエラー出力を行わず、`os.Exit` が呼ばれないことを確認する（`os.Exit` はモック／フックで検証）。
- 異常系
  - `rootCmd.Execute()` がエラーを返す場合に、`Logger` が適切なメッセージを出力し、`os.Exit(1)` が呼ばれることを確認する。

### レイヤ構造の方針

- **CLI 層**
  - Cobra コマンド (`runPush`, `runPull`, `runList`, `runSetup`, `runClean`) からなる。
  - 引数／フラグのパースと、エラー時の Exit/メッセージ表示のみを担当する「薄いラッパー」とする。
  - 実際の処理はすべて「コアロジック層」の関数に委譲する。
- **コアロジック層**
  - ビジネスロジック（Push/Pull/List/Setup/Clean）を環境依存のない形で実装する。
  - I/O は `FileSystem` / `BwClient` / `Logger` インターフェースに抽象化する。
- **インフラ層**
  - `exec.Command` による `bw` CLI 呼び出しや、`os` ベースのファイル操作など、現行実装に近い具体クラスを持つ。

---

## 2. 抽象インターフェース

### BwClient インターフェース

- `type BwClient interface`
  - Bitwarden CLI (`bw`) とのやり取りを抽象化する。
  - すべてのメソッドは **ネットワーク／CLI 側のエラー** を `error` で返す。

#### テストシナリオ（インターフェース利用側）

- 正常系
  - 各コア関数で `BwClient` のモック実装を使用し、想定どおりのメソッド呼び出し順序・引数で呼ばれていることを検証する。
- 異常系
  - モック `BwClient` の各メソッドがエラーを返すように設定し、そのエラーがコア関数からラップされて伝播することを確認する。

#### GetDotenvsFolderID

- "dotenvs" フォルダの ID を取得する。
- フォルダが存在しない場合は `"dotenvs folder not found"` に相当するエラーを返す。
- Bitwarden CLI がロックされている場合は、`ErrBitwardenLocked` を含むエラーを返す。

##### テストシナリオ

- 正常系
  - CLI 出力に "dotenvs" フォルダが含まれている場合に、該当フォルダの ID が返されることを確認する。
- 異常系
  - CLI 出力に "dotenvs" フォルダが含まれない場合に、`"dotenvs folder not found"` 相当のエラーが返ることを確認する。
  - CLI 出力に "Master password" を含めた場合に、`ErrBitwardenLocked` が返ることを確認する。

#### ListItemsInFolder

- 指定フォルダ ID 内のアイテム一覧を `[]Item` として返す。
- フォルダが空の場合でも、空スライスと `nil` エラーを返す。
- CLI ロック時は `ErrBitwardenLocked` を含むエラーを返す。

##### テストシナリオ

- 正常系
  - CLI 出力に 3 件のアイテムが含まれている場合に、`[]Item` に 3 件が正しくパースされることを確認する。
  - CLI 出力が空配列 `[]` の場合に、空スライスと `nil` エラーが返ることを確認する。
- 異常系
  - CLI 出力が JSON ではない文字列の場合に、パースエラーが返ることを確認する。
  - 出力に "Master password" を含めた場合に、`ErrBitwardenLocked` が返ることを確認する。

#### GetItemByName

- 指定フォルダ ID 内から、名前が一致するアイテムを検索して `*FullItem` を返す。
- 見つからない場合は `nil, nil` を返す。
- ロック状態の場合は `ErrBitwardenLocked` を含むエラーを返す。

##### テストシナリオ

- 正常系
  - CLI 出力に対象 `Name` を持つアイテムが含まれているときに、該当 `FullItem` が返ることを確認する（内部で `GetItemByID` が呼ばれる前提）。
  - CLI 出力が空、またはマッチする `Name` が無い場合に、`nil, nil` が返ることを確認する。
- 異常系
  - CLI 出力に "Master password" を含めた場合に、`ErrBitwardenLocked` が返ることを確認する。
  - CLI 出力が JSON ではない（`[ ...` で始まらない）場合に、JSON 形式エラーが返ることを確認する。

#### GetItemByID

- 指定 ID のアイテムを取得して `*FullItem` を返す。
- 存在しない ID の場合は適切なエラーを返す。
- ロック状態の場合は `ErrBitwardenLocked` を含むエラーを返す。

##### テストシナリオ

- 正常系
  - CLI 出力が単一のアイテム JSON の場合に、`FullItem` に正しくパースされることを確認する。
- 異常系
  - CLI 出力が空文字列の場合に、`"no output from bw get item command"` 相当のエラーを返すことを確認する。
  - CLI 出力に "Master password" を含めた場合に、`ErrBitwardenLocked` が返ることを確認する。

#### CreateNoteItem

- フォルダ ID／名前／ノート文字列を受け取り、Secure Note（タイプ 2）として新規アイテムを作成する。
- 成功時は `nil`、失敗時は CLI 出力を含んだエラーを返す。

##### テストシナリオ

- 正常系
  - `bw get template item` が成功し、そのテンプレートに値を上書きして `bw create item` に渡すケースで、エラーが返らないことを確認する（モックで呼び出し内容を検証）。
- 異常系
  - `bw get template item` が失敗し、`createItemDirectly` 経由の作成を行うパスで、`bw create item` のエラーが呼び出し元に伝播することを確認する。
  - `bw create item` が非 0 終了コードとエラーメッセージを返す場合に、そのメッセージを含んだエラーが返ることを確認する。

#### UpdateNoteItem

- アイテム ID と新しいノート文字列を受け取り、既存アイテムの `notes` フィールドを更新する。
- 成功時は `nil` を返す。
- ロック状態の場合は `ErrBitwardenLocked` を含むエラーを返す。

##### テストシナリオ

- 正常系
  - `bw get item` が既存アイテム JSON を返し、`notes` のみ書き換えたうえで `bw encode` → `bw edit item ID ENCODED` を呼び出すケースで、エラーが返らないことを確認する。
  - `bw encode` が失敗し、標準入力に JSON を流し込む `bw edit item ID` のフォールバックが成功するケースで、エラーが返らないことを確認する。
- 異常系
  - `bw get item` の JSON が壊れている場合に、パースエラーが返ることを確認する。
  - `bw edit item`（エンコードあり／なし両方）が失敗した場合に、その出力を含んだエラーが返ることを確認する。

#### Login

- メールアドレス・パスワード・サーバー URL を受け取り、`bw login` 相当の処理を行う。
- 成功時は `nil` を返す。
- 失敗時は CLI 出力を含んだエラーを返す。

##### テストシナリオ

- 正常系
  - サーバー URL が空のときに、`bw config server` が呼ばれず、`bw login email password` が呼ばれて成功するパスを確認する。
  - サーバー URL が既存設定と同一の場合に、`bw logout` が呼ばれないことを確認する。
- 異常系
  - サーバー URL 変更時に `bw config server` が "Logout required" を返し、`bw logout` → `bw config server` 再試行の流れになることを確認する。
  - `bw login` が非 0 終了コードとエラーメッセージを返す場合に、そのメッセージを含んだエラーが返ることを確認する。

#### Unlock

- マスターパスワードを受け取り、`bw unlock` 相当の処理を行う。
- セッションキーの設定／検証まで内部で行い、成功時は `nil` を返す。
- ロック解除ができなかった場合は、詳細なメッセージを含むエラーを返す。

##### テストシナリオ

- 正常系
  - 方法1（`--passwordfile --raw`）でセッションキー文字列が標準出力に出力され、`BW_SESSION` が設定されるケースで成功と判定されることを確認する。
  - 方法4（`bw unlock password`）で標準出力に長いトークンが出力され、それをセッションキーとみなして成功と判定するケースを確認する。
- 異常系
  - すべての方法（passwordfile/passwordenv/引数）が非 0 終了コードを返し、`bw status` の結果も `"status":"unlocked"` にならない場合に、まとめたエラーメッセージが返ることを確認する。
  - 出力にセッションキーらしき文字列があるが `bw status` が locked のままの場合に、失敗として扱われることを確認する。

### FileSystem インターフェース

- `type FileSystem interface`
  - 実ディスクへのアクセスを抽象化する。

#### OpenEnvFile

- `.env` ファイルのパスを受け取り、読み取り用ハンドル／リーダを返す。
- 存在しない場合は `"not found"` 相当のエラーを返す。

#### ReadFile

- 任意パスのファイル内容を文字列として読み込んで返す。

#### WriteFile

- 任意パスに文字列を保存する。
- パーミッションは呼び出し側（コアロジック）から指定できるようにする。

#### Remove

- 任意パスのファイルを削除する。
- 存在しない場合はエラーを返す（呼び出し側で必要なら事前に存在確認する）。

#### Stat

- 任意パスの存在確認／属性取得を行う。
- 存在しない場合は明示的に判定できるエラー（`os.IsNotExist` と同等の扱い）を返す。

#### MkdirAll

- ディレクトリを再帰的に作成する。
- すでに存在する場合は成功扱いとする。

##### テストシナリオ（FileSystem 実装）

- 正常系
  - 既存ディレクトリに対して `MkdirAll` を呼び出してもエラーが発生しないことを確認する。
  - `WriteFile` 後に `ReadFile` で同じ内容が取得できることを確認する。
  - `WriteFile` 後に `Remove` すると、以降の `Stat` / `ReadFile` で存在しない扱いになることを確認する。
- 異常系
  - 読み取り専用ディレクトリ配下に `WriteFile` を試みた際にエラーが返ることを確認する（実環境ではスキップ可・モックで代用可）。
  - 存在しないパスに対する `Remove` がエラーを返すことを確認する。

### Logger インターフェース

- `type Logger interface`
  - 既存の `utils.Error/Success/...` をラップするための抽象化。

#### Error / Errorf / Errorln

- エラー系メッセージを標準エラー出力に送る。
- カラー出力を有効／無効にするかは具象実装に委ねる。

#### Info / Infoln

- 情報メッセージを標準出力に送る。

#### Success / Successln

- 成功メッセージを標準出力に送る。

#### Warning / Warningln

- 警告メッセージを標準出力に送る。

##### テストシナリオ（Logger 実装）

- 正常系
  - `Error` / `Errorln` が標準エラー出力ストリーム（もしくはモック）に書き込まれることを確認する。
  - `Info` / `Success` / `Warning` 系が標準出力ストリームに書き込まれることを確認する。
- 異常系
  - 出力先が閉じられている／書き込めない状態のときに、パニックにならず内部エラーとして扱う（もしくは仕様として特に考慮しない）ことを確認する。

---

## 3. 共通ユーティリティ（ロック時リトライ）

### WithUnlockRetry

- シグネチャ（イメージ）:
  - `func WithUnlockRetry(bw BwClient, cfg *config.Config, promptPassword func() (string, error), logger Logger, fn func() error) error`
- 処理内容:
  - `fn()` を一度実行し、エラーが `ErrBitwardenLocked` またはロック関連エラーの場合のみ、アンロック／ログインを行う。
  - アンロックフロー:
    - `promptPassword` を使ってマスターパスワードを取得する。
    - `bw.Unlock` を実行する。
    - 失敗した場合、`cfg` が存在すれば `bw.Login(cfg.Email, password, cfg.SelfhostedURL)` を試みる。
    - ログイン成功後に再度 `bw.Unlock` を実行する。
  - アンロック／ログインに成功した場合、`fn()` を再実行し、その結果を返す。
  - 途中でのパスワード取得失敗やアンロック／ログイン失敗は、適切なエラーメッセージをラップして返す。

#### テストシナリオ

- 正常系
  - `fn()` が 1 回目で成功する場合に、`promptPassword`／`bw.Unlock`／`bw.Login` が一切呼ばれないことを確認する。
  - `fn()` が 1 回目で `ErrBitwardenLocked` を返し、パスワード入力 → `bw.Unlock` 成功 → 2 回目の `fn()` が成功するフローを確認する。
- 異常系
  - `promptPassword` がエラーを返した場合に、`bw.Unlock`／`bw.Login` が呼ばれず、そのエラーが `WithUnlockRetry` から返ることを確認する。
  - `bw.Unlock` と `bw.Login` の両方が失敗する場合に、その失敗理由を含んだエラーが返り、`fn()` が再実行されないことを確認する。

### ParseEnvFile / EnvDataToJSON / RestoreEnvFileFromJSON / FindEnvFile

- 既存の挙動を維持するが、テストでは `FileSystem` のモックを介して利用できるようにする（`os` 直呼び出しを内部に閉じ込める）。

---

## 4. コアロジック層の関数

### PushEnvCore

- シグネチャ（イメージ）:
  - `func PushEnvCore(fromDir, projectName string, fs FileSystem, bw BwClient, cfg *config.Config, promptPassword func() (string, error), logger Logger) error`
- 処理:
  - `fromDir` と `projectName` に基づいて対象 `.env` ファイルパスを決定する。
    - `fromDir` が `"."` または `".."` の場合、`/project-root` へのフォールバックを含む。
  - `fs.OpenEnvFile` を利用して `.env` を開き、`ParseEnvFile` で `EnvData` にパースする。
  - `EnvDataToJSON` で JSON 文字列に変換する。
  - `WithUnlockRetry` を使って `"dotenvs"` フォルダ ID 取得 (`bw.GetDotenvsFolderID`) を行う。
  - 同じく `WithUnlockRetry` を通じて、`bw.GetItemByName(folderID, projectName)` で既存アイテムを検索する。
  - 以下の分岐:
    - 既存アイテムが **ある** 場合:
      - 呼び出し側（CLI 層）で「上書き確認」済みである前提とし、`WithUnlockRetry` を通じて `bw.UpdateNoteItem(item.ID, jsonData)` を実行する。
    - 既存アイテムが **ない** 場合:
      - `WithUnlockRetry` 経由で `bw.CreateNoteItem(folderID, projectName, jsonData)` を実行し、新規作成する。
  - エラー発生時はそれぞれのエラーを呼び出し元に返す（Exit は CLI 層で行う）。

#### テストシナリオ

- 正常系
  - `.env` が存在し、`GetItemByName` で既存アイテムが **見つからない** ケースで `CreateNoteItem` が呼ばれ、エラーなく終了することを確認する。
  - `.env` が存在し、既存アイテムが **見つかる** ケースで `UpdateNoteItem` が呼ばれ、エラーなく終了することを確認する（上書き確認は CLI 側で済んでいる前提）。
  - `fromDir` が `"."` かつカレントディレクトリに `.env` がなく、`/project-root/.env` が存在するケースで、フォールバック先が正しく利用されることを確認する。
- 異常系
  - `.env` ファイルが見つからない場合に、適切な「not found」系エラーが返ることを確認する。
  - `.env` パース時に I/O エラーやスキャナエラーが発生した場合に、そのエラーがラップされて返ることを確認する。
  - `bw.GetDotenvsFolderID` や `bw.CreateNoteItem` がロック関連以外のエラーを返す場合に、そのエラーが `WithUnlockRetry` 経由で呼び出し元へ伝播することを確認する。

### PullEnvCore

- シグネチャ（イメージ）:
  - `func PullEnvCore(outputDir, projectName string, fs FileSystem, bw BwClient, cfg *config.Config, promptPassword func() (string, error), confirmOverwrite func(path string) (bool, error), logger Logger) error`
- 処理:
  - `outputDir` が `"."` または `".."` の場合、`/project-root` に正規化する。
  - `WithUnlockRetry` 経由で `bw.GetDotenvsFolderID()` を呼び出し、フォルダ ID を取得する。
  - `WithUnlockRetry` 経由で `bw.GetItemByName(folderID, projectName)` を呼び出し、対応するアイテムを取得する。
  - アイテムが存在しない場合は `"Item '%s' not found in dotenvs folder"` 相当のエラーを返す。
  - `RestoreEnvFileFromJSON(item.Notes)` で `.env` 内容文字列を復元する。
  - 出力先 `.env` パス（`outputDir/.env`）を構成し、`fs.Stat` で存在を確認する。
    - すでに存在する場合は `confirmOverwrite(envPath)` を呼び出し、`false` の場合は `"operation cancelled"` 相当のエラーを返す、または `nil` を返して何もしない。
  - 必要に応じて `fs.MkdirAll(outputDir)` を呼び出し、ディレクトリを作成する。
  - `fs.WriteFile(envPath, content, 0644)` で `.env` を書き出す。

#### テストシナリオ

- 正常系
  - 対象アイテムが存在し、出力先に `.env` が存在しない場合に、新規で `.env` が作成されることを確認する。
  - 出力ディレクトリが存在しない場合に `MkdirAll` が呼ばれ、その後 `.env` が正しい内容で書き込まれることを確認する。
  - `outputDir` が `"."` または `".."` の場合に `/project-root` へ正しく正規化されることをモックで確認する。
- 異常系
  - `bw.GetItemByName` が `nil, nil`（アイテム無し）を返すケースで、「Item not found」系エラーが返ることを確認する。
  - 既存 `.env` がある状態で `confirmOverwrite` が `false` を返す場合に、上書き処理を行わず、適切にキャンセル扱いになることを確認する。
  - `RestoreEnvFileFromJSON` が壊れた JSON でエラーを返す場合に、そのエラーがそのまま呼び出し元へ伝播することを確認する。

### CleanEnvCore

- シグネチャ（イメージ）:
  - `func CleanEnvCore(targetDir, projectName string, fs FileSystem, bw BwClient, cfg *config.Config, promptPassword func() (string, error), selectMismatchAction func(mismatchedFiles []string) (CleanMismatchAction, error), logger Logger) error`
- 処理:
  - `FindEnvFiles` と同一のルール（`.env*` かつ `.example` 除外）でローカル対象ファイルを検出する。
  - ローカルに対象ファイルが無い場合はエラーを返す。
  - `WithUnlockRetry` 経由で `bw.GetDotenvsFolderID()` / `bw.GetItemByName(folderID, projectName)` を実行し、dotenvs 配下の同名 Note アイテムを取得する。
  - アイテムが存在しない、またはアイテム内に管理対象ファイルが無い場合は、ローカルを削除せずエラーで中止する。
  - ローカルとリモートの `MultiEnvData` を比較する。1 ファイルでも差分（欠落含む）があればプロジェクト全体を差分ありとみなす。
  - 差分が無い場合は確認なしでローカル対象ファイルを `fs.Remove` する。
  - 差分がある場合は `selectMismatchAction` で単一選択させる（Abort / Overwrite remote then clean / Remove local）。
    - Abort: 削除せず中止（`ErrCleanAborted`）。
    - Overwrite remote then clean: ローカル内容でリモートを更新（push 相当）してからローカルを削除する。
    - Remove local: リモートを更新せずローカルのみ削除する（危険選択肢）。

#### テストシナリオ

- 正常系
  - リモート同名アイテムがあり内容が一致する場合に、確認コールバックを呼ばずローカル `.env*` が `Remove` されることを確認する。
  - 差分があり `Overwrite remote then clean` が選ばれた場合に、`UpdateNoteItem`（または作成）の後にローカルが削除されることを確認する。
  - 差分があり `Remove local` が選ばれた場合に、リモート更新なしでローカルのみ削除されることを確認する。
- 異常系
  - リモートに同名アイテムが無い場合に、ローカルを削除せずエラーになることを確認する。
  - リモートアイテムはあるが管理対象ファイルが空の場合に、ローカルを削除せずエラーになることを確認する。
  - 差分ありで `Abort` が選ばれた場合に、ローカルもリモートも変更されないことを確認する。

### ListDotenvsCore

- シグネチャ（イメージ）:
  - `func ListDotenvsCore(bw BwClient, cfg *config.Config, promptPassword func() (string, error), logger Logger) ([]Item, error)`
- 処理:
  - `WithUnlockRetry` 経由で `bw.GetDotenvsFolderID()` を実行し、フォルダ ID を取得する。
  - `WithUnlockRetry` 経由で `bw.ListItemsInFolder(folderID)` を実行し、アイテム一覧を取得する。
  - 取得した `[]Item` をそのまま呼び出し元に返す。
  - アイテム 0 件は正常ケースとして空スライスを返す。

#### テストシナリオ

- 正常系
  - `bw.ListItemsInFolder` が 3 件のアイテムを返す場合に、同じ 3 件の `[]Item` が戻り値となることを確認する。
  - `bw.ListItemsInFolder` が空スライスを返す場合に、空スライスと `nil` エラーが返ることを確認する。
- 異常系
  - `bw.GetDotenvsFolderID` がロック関連エラーを返し、`WithUnlockRetry` を経ても失敗した場合に、そのエラーが `ListDotenvsCore` から返ることを確認する。
  - `bw.ListItemsInFolder` がロック関連以外のエラーを返す場合に、そのエラーがラップされずに呼び出し元に伝播することを確認する。

### SetupBitwardenCore

- シグネチャ（イメージ）:
  - `func SetupBitwardenCore(fs FileSystem, bw BwClient, logger Logger, selectHostType func() (string, error), inputURL func() (string, error), inputEmail func() (string, error), inputPassword func() (string, error)) error`
- 処理:
  - 既存設定を `LoadConfig` で読み込み、存在する場合は上書きされる旨を `logger` で通知する。
  - `selectHostType` で `"cloud"` または `"selfhosted"` を選択させる。
  - `"selfhosted"` の場合は `inputURL` で Self-hosted URL を取得する。
  - `inputEmail` でメールアドレスを取得する。
  - `inputPassword` でパスワードを非エコー入力させる。
  - `bw.Login(email, password, selfhostedURL)` を実行し、ログインを試みる。
  - ログイン成功時に `SaveConfig` で設定（HostType / URL / Email）を保存する。
  - 失敗時にはエラーを呼び出し元に返す（Exit は CLI 層が行う）。

#### テストシナリオ

- 正常系
  - 既存設定が無い状態で、`cloud` が選択され、URL 入力を行わずに `Login` が成功し、`SaveConfig` が正しい内容で呼び出されることを確認する。
  - 既存設定がある状態で、`selfhosted` が選択され、新しい URL・メールアドレス・パスワードで `Login` に成功し、上書き保存されることを確認する。
- 異常系
  - `selectHostType` がエラーを返した場合に、後続の入力処理／`Login`／`SaveConfig` が呼ばれず、そのエラーが返ることを確認する。
  - `bw.Login` がエラーを返した場合に、設定ファイル保存が行われないことを確認する。

---

## 5. CLI 層（Cobra コマンド）

### runPush

- Cobra から呼ばれるエントリポイント。
- 処理:
  - `--from` フラグをパースして `fromDir` を取得する。
  - `os.Getwd()` からカレントディレクトリ名を取り出し、`projectName` として利用する。
  - 具象 `BwClient` / `FileSystem` / `Logger` / `config.Config` を生成する。
  - 上書き確認のための UI 関数（`confirmOverwrite`）を定義しておく。
  - `PushEnvCore` を呼び出し、戻り値のエラーに応じて:
    - エラーあり: `Logger` でエラーメッセージを出力し `os.Exit(1)`。
    - 正常終了: 成功メッセージを表示し、終了コード 0 で返る。

#### テストシナリオ

- 正常系
  - `PushEnvCore` が `nil` を返す場合に、`Logger` の成功メッセージ出力が呼ばれ、`os.Exit` が呼ばれないことを確認する。
- 異常系
  - `PushEnvCore` がエラーを返す場合に、その内容が `Logger` を通じて出力され、`os.Exit(1)` が呼ばれることを確認する。

### runPull

- 処理:
  - `--output` フラグをパースして `outputDir` を取得する。
  - `os.Getwd()` から `projectName` を決定する。
  - 具象 `BwClient` / `FileSystem` / `Logger` を生成する。
  - 上書き確認用 `confirmOverwrite` 関数を用意する。
  - `PullEnvCore` を呼び出し、エラー時はメッセージ表示＋`os.Exit(1)`、成功時は成功メッセージを表示する。

#### テストシナリオ

- 正常系
  - `PullEnvCore` が `nil` を返す場合に、成功メッセージのみが出力されることを確認する。
- 異常系
  - `PullEnvCore` がエラーを返す場合に、そのエラー内容が `Logger` から出力され、`os.Exit(1)` が呼ばれることを確認する。

### runList

- 処理:
  - 具象 `BwClient` / `Logger` / `config.Config` を生成する。
  - `ListDotenvsCore` を呼び出し、戻り値の `[]Item` を標準出力に 1 行ずつ出力する。
  - アイテム 0 件の場合は `"No items found in dotenvs folder"` 相当のメッセージのみを出力する。

#### テストシナリオ

- 正常系
  - `ListDotenvsCore` が 3 件のアイテムを返す場合に、その `Name` が 3 行出力されることを確認する。
  - 空スライスが返る場合に、「No items found in dotenvs folder」のみが出力され、`os.Exit` が呼ばれないことを確認する。
- 異常系
  - `ListDotenvsCore` がエラーを返す場合に、その内容が標準エラー出力に出力され、`os.Exit(1)` が呼ばれることを確認する。

### runSetup

- 処理:
  - `BwClient` / `FileSystem` / `Logger` を生成する。
  - 入力 UI (`SelectHostType` / `InputURL` / `InputEmail` / `InputPassword`) を関数として `SetupBitwardenCore` に渡す。
  - `SetupBitwardenCore` の戻り値がエラーの場合はメッセージ表示＋`os.Exit(1)`。
  - 正常終了時にはサインイン成功メッセージを表示する。

#### テストシナリオ

- 正常系
  - `SetupBitwardenCore` が `nil` を返す場合に、サインイン成功メッセージのみが出力されることを確認する。
- 異常系
  - `SetupBitwardenCore` がエラーを返す場合に、その内容がエラーとして表示され、`os.Exit(1)` が呼ばれることを確認する。

### runClean

- 処理:
  - `--from` フラグをパースして対象ディレクトリを取得する。
  - `os.Getwd()` から `projectName` を決定する。
  - 具象 `BwClient` / `FileSystem` / `Logger` を生成する。
  - 差分時の単一選択 UI (`SelectCleanMismatchAction`) を `CleanEnvCore` へ渡す。
  - `CleanEnvCore` の戻り値が `ErrCleanAborted` の場合は情報メッセージを出して終了コード 0。
  - その他のエラーはメッセージ表示＋`os.Exit(1)`、成功時は成功メッセージを表示する。

#### テストシナリオ

- 正常系
  - `clean` コマンドが root に登録され、`--from` フラグのデフォルトが `"."` であることを確認する。
- 異常系
  - （コア層で担保）リモート未整備や Abort 時にローカル削除が行われないことを確認する。

---

## 6. 既存ユーティリティの仕様（インフラ実装側）

### Config 周り

- `GetConfigPath`
  - ユーザーのホームディレクトリ配下の `.config/bwsf/config.json` へのフルパスを返す。
  - ホームディレクトリ取得に失敗した場合はエラー。
- `LoadConfig`
  - 設定ファイルが存在しない場合は `nil, nil` を返す（エラー扱いしない）。
  - 存在する場合は JSON を `Config` にアンマーシャルして返す。
- `SaveConfig`
  - 必要に応じてディレクトリを作成し、`Config` を `0600` で書き出す。

### 入力／色付き出力ユーティリティ

- `SelectHostType` / `InputURL` / `InputEmail` / `InputPassword` / `ConfirmOverwrite`
  - 既存仕様を維持しつつ、テストではモック関数に差し替えられるようにする。
- `Error` / `Errorln` / `Success` / `Successln` / `Warning` / `Warningln` / `Info` / `Infoln` / `Question` / `Questionln`
  - `Logger` の具象実装として利用される。

### `.env` パースユーティリティ

- `ParseEnvFile` / `EnvDataToJSON` / `RestoreEnvFileFromJSON` / `FindEnvFile`
  - 既存の仕様を維持する（行順／コメント／クォートをそのまま保持）。

### ロック判定ユーティリティ

- `ErrBitwardenLocked`
  - Bitwarden CLI がロックされていることを表す代表的なエラー。
- `IsLockedError`
  - 渡されたエラーがロック関連であるかを判定し、`WithUnlockRetry` などから利用される。

