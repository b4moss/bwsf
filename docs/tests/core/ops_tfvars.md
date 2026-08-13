# core: tfvars を含む push / pull / list / clean（Issue #102 / v0.15.0）

対象パッケージ: `app/src/core`（必要なら `app/src/cmd` の表示文言の薄い更新）  
方針: 管理対象に tfvars を足しても、**コマンド契約は現行 `.env*` と対称**（#102 Q3: キー=ファイル名）。  
ユニットはモック中心。

関連: [`managed_files_tfvars.md`](./managed_files_tfvars.md)

---

### PushEnvCore

- 検出された管理対象ファイルを読み、MultiEnvData（キー=ベースファイル名）として 1 Note に格納する
- `.env*` のみ / tfvars のみ / 混在、いずれも成功しうる
- 本文は行配列のまま保存（HCL/JSON を解釈しない）

#### テスト：正常系

- `terraform.tfvars` のみでも `CreateNoteItem` / `UpdateNoteItem` が呼ばれ成功する
- `.env` と `terraform.tfvars` が両方あるとき、Notes JSON に両方のキーが含まれる
- `prod.auto.tfvars` / `secret.tfvars.json` もキーとしてそのまま載る

#### テスト: 異常系

- 管理対象が 0 件（例: `.example` のみ、または無関係ファイルのみ）のときエラー（「見つからない」系。文言は実装で固定しテストする）
- 既存どおり、フォルダ取得失敗・更新失敗は伝播する（退行）

---

### PullEnvCore

- Note の MultiEnvData から、キー名のファイルを出力ディレクトリに書き戻す
- tfvars キーが含まれていれば、対応するファイル名で復元する
- 既存ファイルがある場合の上書き確認は現行どおり（ファイル単位）

#### テスト：正常系

- Notes に `terraform.tfvars` があるとき `WriteFile(.../terraform.tfvars)` が呼ばれ、内容が一致する
- `.env` と tfvars が混在する Note から両方復元できる
- 旧形式（単一 EnvData）の退行は維持（従来どおり `.env` 復元）

#### テスト: 異常系

- アイテム無しは既存どおり not found
- 上書き拒否されたファイルは書かれない（既存挙動の退行）

---

### GetPushedEnvFiles / GetPulledEnvFiles（表示用）

- 名称は歴史的に Env だが、**管理対象ファイル名一覧**を返す（tfvars を含む）
- pull 側は Note 内キー一覧（ソートポリシーは検出仕様に合わせる）

#### テスト：正常系

- push 表示一覧に `terraform.tfvars` が含まれる
- pull 表示一覧に Note 内の tfvars キーが含まれる
- `.example` は含まれない

#### テスト: 異常系

- 対象 0 件のときの挙動は現行 push/pull と整合（エラー or 空。実装で固定してテスト）

---

### CleanEnvCore

- ローカル管理対象（`.env*` + tfvars）を検出し、リモート Note 内容と突き合わせてから削除する
- 一致 / 差分 / リモート欠落の分岐は現行 clean 仕様のまま（対象集合だけ拡張）

#### テスト：正常系

- ローカルに `terraform.tfvars` があり、リモート同内容なら確認なしで `Remove` される
- `.env` と tfvars が両方一致なら両方削除される

#### テスト: 異常系

- リモートに該当キーが無い / 管理対象ファイルがリモートに無い扱いは現行どおり中止（ローカルを消さない）
- 1 ファイルでも差分があれば mismatch 分岐（現行）。tfvars だけ差分でも同様

---

### 退行（`.env*` のみの既存ケース）

#### テスト：正常系

- 既存の `.env` / `.env.staging` push・pull・clean 正常系が引き続き通る

#### テスト: 異常系

- `.env.example` 除外など既存除外が壊れない
