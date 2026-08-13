# core: 管理対象ファイル検出（Issue #102 / v0.15.0）

対象パッケージ: `app/src/core`（現行 `findEnvFilesFromFS` / `isExampleFile` / 並び替えの後継を含む）  
前提合意（#102）:

| # | 合意 |
|---|------|
| Q1 | 末尾 `.tfvars` と末尾 `.tfvars.json` を対象（`*.auto.tfvars` は `.tfvars` でカバー） |
| Q2 | 名前に `.example` を含むものは除外（`.env` と同じ） |
| Q3 | MultiEnvData キーはファイル名そのまま（本ファイルでは検出のみ。キーは ops 仕様） |

内容の保存形式は現行どおり **行配列（`EnvData.Lines`）**。HCL / JSON を特別パースしない（往復でバイト列相当の行が復元できればよい）。

---

### 対象判定（検出）

ディレクトリ直下の通常ファイルについて:

1. 次のいずれかに該当する
   - 名前が `.env` で始まる（現行）
   - 名前が `.tfvars` で終わる
   - 名前が `.tfvars.json` で終わる
2. かつ、名前に `.example` を **含まない**

ディレクトリは対象外。

#### テスト：正常系

- `.env` / `.env.staging` / `terraform.tfvars` / `prod.auto.tfvars` / `vars.tfvars.json` / `prod.auto.tfvars.json` が検出される
- `.env` のみ、または `terraform.tfvars` のみでも、それぞれ 1 件以上として検出される（**`.env` 必須ではない**）

#### テスト: 異常系 / 除外

- `.env.example` / `.env.staging.example` / `terraform.tfvars.example` / `foo.tfvars.json.example` は検出されない
- `main.tf` / `variables.tf` / `README.md` は検出されない
- `.tfvars` を含んでも末尾が `.tfvars` / `.tfvars.json` でないもの（例: `notes.tfvars.bak`）は検出しない

---

### 並び順

- 現行どおり、ベース名 `.env` がある場合は **先頭**
- それ以外はベース名の辞書順（安定してテスト可能な順序）

#### テスト：正常系

- `.env` と `terraform.tfvars` と `.env.staging` が混在するとき、先頭は `.env`
- `.env` が無いとき（tfvars のみ）は辞書順

---

### 公開ヘルパ（任意の実装形）

検出ロジックは push / clean / 表示用一覧で共有する。  
関数名は実装時に `findEnvFilesFromFS` の拡張、または `findManagedFilesFromFS` への改名などとしてよい。仕様は「共有検出」を満たすこと。

#### テスト：正常系

- 同一モックディレクトリに対し、push 対象一覧と clean 対象一覧が同じ集合になる（順序も同一ポリシー）

#### テスト: 異常系

- `ReadDir` 失敗時はエラーを返し、空スライスを成功扱いにしない
