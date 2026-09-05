# config / core: `save_files` と `!` 否定（Issue #177 / v0.20.0 §2.3）

対象パッケージ:

- `app/src/core`（基盤通過後の basename フィルタ）
- `app/src/config`（グローバル／プロジェクトのスキーマ検証）
- `app/src/cmd`（push / pull / clean への実効リスト配線）

Issue: [#177](https://github.com/b4moss/bwsf/issues/177)  
製品正本: [`../specs/v0.20.0-multi-host.md`](../specs/v0.20.0-multi-host.md) §2.3

関連:

- 旧プロジェクト契約 → [`project_local.md`](./project_local.md)（**フィルタ節は本仕様が v0.20 で優先**）
- グローバル I/O → [`global_v2.md`](./global_v2.md)

---

### 前提（本仕様で固定）

| # | 方針 |
|---|------|
| F1 | キーは **`save_files` のみ**。グローバル・プロジェクトとも **`not_save_files` は廃止**。キーが存在したらスキーマエラー（空配列でもキーありならエラーにしてよい。実装は「キー検出でエラー」をテストで固定） |
| F2 | 未設定／空配列 → 基盤ルールのみ（管理対象をそのまま） |
| F3 | プロジェクトに `save_files` が（非空で）あれば、その内容で **完全オーバーライド**。グローバルとマージしない |
| F4 | プロジェクトに `save_files` が無い／空 → グローバル `settings.save_files` を使う。グローバルも無ければ基盤のみ |
| F5 | 適用は常に **基盤ルール（`isManagedFileName` + `.example` 除外）の後**。基盤外を `save_files` で新規取り込みしない |
| F6 | エントリが `!` で始まる → 否定 glob（`!` の直後がパターン）。それ以外は肯定 glob |
| F7 | 適用順: (1) 肯定が1つ以上 → それらに一致するものだけ候補 (2) 肯定無し → 基盤通過すべてが候補 (3) 否定で除外 |
| F8 | glob は basename に対し `filepath.Match` 相当 |
| F9 | プロジェクトに任意の `host`（id 参照）を許容。探索・override 名・候補選択は #133 のまま（本仕様はキーとフィルタ） |

### 解釈例

```jsonc
"save_files": [".env*", "!.env.local"]
```

→ 基盤通過のうち `.env*` に合うものから `.env.local` を除く。

```jsonc
"save_files": ["!.env.local", "!*.auto.tfvars"]
```

→ 肯定無しのため基盤通過すべてが候補 → 否定のみ除外。

---

## 1. スキーマ（プロジェクト）

#### テスト：正常系

| # | 入力 | 期待 |
|---|------|------|
| P1 | `save_files` のみ（肯定／`!` 混在可） | ロード成功 |
| P2 | `save_files` 省略 | フィルタ未設定 |
| P3 | `save_files: []` | 未設定扱い（基盤のみ） |
| P4 | `host: "default"` + `save_files` | ロード成功（解決は [`host_resolve.md`](./host_resolve.md)） |

#### テスト: 異常系

| # | 入力 | 期待 |
|---|------|------|
| P5 | `not_save_files` キーあり（値が空でも） | スキーマエラー |
| P6 | `save_files` と `not_save_files` 併記 | スキーマエラー |

---

## 2. スキーマ（グローバル）

#### テスト：正常系

| # | 入力 | 期待 |
|---|------|------|
| G1 | `settings.save_files` あり | OK |
| G2 | `settings.save_files` 省略／`[]` | フィルタ未設定 |

#### テスト: 異常系

| # | 入力 | 期待 |
|---|------|------|
| G3 | `settings.not_save_files` あり | スキーマエラー |

---

## 3. フィルタ単体（core）

共通の基盤通過候補例: `.env`, `.env.local`, `.env.production`, `foo.tfvars`, `x.auto.tfvars`  
（`.env.example` は基盤で既に除外）

| # | `save_files` | 含まれる | 含まれない |
|---|--------------|----------|------------|
| A1 | （なし／空） | 上記すべて | `.env.example` |
| A2 | `[".env", ".env.production"]` | `.env`, `.env.production` | `.env.local`, tfvars 類 |
| A3 | `[".env*", "!.env.local"]` | `.env`, `.env.production` | `.env.local`, tfvars 類 |
| A4 | `["!.env.local", "!*.auto.tfvars"]` | `.env`, `.env.production`, `foo.tfvars` | `.env.local`, `x.auto.tfvars` |
| A5 | `["secrets.yaml"]` のみ（基盤外） | （空） | `secrets.yaml` は取り込まない |
| A6 | `[".env*", "!.env*", "!.env.local"]`（肯定に全部当てはまるが否定で消える等） | 実装は F7 どおり。肯定で残ったあと否定を適用 |

---

## 4. グローバルとプロジェクトのオーバーライド（配線）

HOME にグローバル、cwd 側にプロジェクト設定。push/clean の管理対象一覧、または `resolve` + `findEnvFiles` 相当で検証。

| # | グローバル | プロジェクト | 期待 |
|---|------------|--------------|------|
| O1 | `[".env*"]` | （無し） | グローバルどおり（`.env*` のみ） |
| O2 | `[".env*"]` | `["!.env.local"]` | **プロジェクトのみ**（基盤すべてから `.env.local` 除外）。グローバルの `.env*` 制限は使わない |
| O3 | `[".env*", "!.env.local"]` | `["*.tfvars"]` | tfvars のみ（完全置換） |
| O4 | 無し | 無し | 基盤のみ |

---

## 5. 対象外

- 基盤ルール自体の変更（[`../core/managed_files_tfvars.md`](../core/managed_files_tfvars.md)）
- `bwsf init` によるプロジェクト生成（#193）
- list がプロジェクト設定を読むかどうかの変更（#133 どおり list は読まない。フィルタ配線の主対象は push / pull / clean）

---

## 実装メモ（仕様外・参考）

```text
effectiveSaveFiles =
  project.save_files if set/non-empty else global.settings.save_files
→ isManagedFileName
→ split positive / !negative
→ positives? filter : keep all → drop negatives
```
