# config: プロジェクトローカル `.bwsf/config.(json|jsonc)`（Issue #133 / v0.18.0）

対象パッケージ:

- `app/src/config`（探索・スキーマ・JSONC パース・検証）
- `app/src/cmd`（`resolveProjectAndFileDir` への配線。対象コマンドは push / pull / clean）
- `app/src/core`（管理対象検出後の `save_files` / `not_save_files` フィルタ）
- `app/src/utils`（複数候補時の `promptui.Select` 相当）
- `app/src/project`（`Resolve` 第2引数 `overrideProjectName` — 既存スロット）

Issue: [#133](https://github.com/b4moss/bwsf/issues/133)

関連:

- [#134](https://github.com/b4moss/bwsf/issues/134) / [`project/git_root.md`](../project/git_root.md) — git ルート・Name/Dir
- [#155](https://github.com/b4moss/bwsf/issues/155) / [`config/jsonc_load.md`](./jsonc_load.md) — JSONC 読み込み
- [#177](https://github.com/b4moss/bwsf/issues/177) — グローバル同系スキーマ（**本仕様の対象外**）

---

### 前提（本仕様で固定）

| # | 方針 |
|---|------|
| P1 | 設定ファイルは **`.bwsf/config.json`** または **`.bwsf/config.jsonc`**。ルート直下の `bwsf.config.json` は探索しない |
| P2 | 探索は **cwd から親方向**（祖先に `.git` がある場合、そのルート上の候補も含む） |
| P3 | 候補 **0** … プロジェクト設定なし（Name は #134、ファイル選択フィルタなし） |
| P4 | 候補 **1** … それを採用（対話なし） |
| P5 | 候補 **2 以上** … パス一覧を対話選択（`promptui.Select` 相当）。非 TTY / 選択不可時は **エラー**（ハングしない） |
| P6 | 同一ディレクトリに `config.json` と `config.jsonc` が **両方**ある → **エラー** |
| P7 | キーは `override_project_name` と、`save_files` / `not_save_files`（**どちらか一方のみ**。両方非空はロードエラー） |
| P8 | 空文字の `override_project_name`、および空配列のみの `save_files`/`not_save_files` は **未設定**扱い |
| P9 | ファイル選択は **基盤ルール**（`isManagedFileName` + `.example` 除外）の **後**に basename glob を適用する。基盤外ファイルを `save_files` で新規取り込みしない |
| P10 | 読み込みは JSONC 可（#155 の hujson 経路）。本ファイル自体は管理対象に含めない（基盤パターン外） |
| P11 | グローバル `~/.config/bwsf/config.json` の置き換えではない。グローバルの `save_files` 等は #177 |
| P12 | 対象コマンドは **push / pull / clean**。`list` / `setup` / `config` はプロジェクト設定を読まない |

### スキーマ（初期）

```json
{
  "override_project_name": "my-api",
  "not_save_files": [".env.local", "*.auto.tfvars"]
}
```

または肯定側のみ:

```json
{
  "override_project_name": "my-api",
  "save_files": [".env", ".env.production"]
}
```

### ファイル選択の意味

| 状態 | 結果 |
|------|------|
| 両キー未設定（なし／空配列） | 基盤ルールのみ（現行） |
| `not_save_files` のみ | 基盤通過後、glob 一致を **除外** |
| `save_files` のみ | 基盤通過後、glob 一致だけ **残す** |
| 両方非空 | ロードエラー |

glob は basename に対し `filepath.Match` 相当（`*` / `?` 等）。

### プロジェクト名

#134 の解決に加え、本設定の `override_project_name`（非空）が **最優先**。Dir（`--from` / `--output` / git ルート）の規則は #134 のまま変更しない。

---

## 1. 探索（`FindProjectConfigPaths` 相当）

例ツリー:

```text
/repo/                 ← .git あり。ここに .bwsf/config.json がある場合あり
/repo/app/             ← cwd になりうる。ここに別の .bwsf/config.jsonc がある場合あり
/norepo/workdir/       ← .git 無し
```

#### テスト：正常系

| # | 配置 | cwd | 期待候補（順序） |
|---|------|-----|------------------|
| D1 | なし | `/repo/app` | 空 |
| D2 | `/repo/.bwsf/config.json` のみ | `/repo/app` | `[/repo/.bwsf/config.json]`（親方向で発見） |
| D3 | `/repo/app/.bwsf/config.jsonc` のみ | `/repo/app` | その 1 件 |
| D4 | `/repo/.bwsf/config.json` と `/repo/app/.bwsf/config.json` | `/repo/app` | **2 件以上**（cwd 側が先など、順序は仕様で固定してテストする） |
| D5 | `.git` 無し `/norepo/workdir/.bwsf/config.json` | 同 | 1 件（親を辿れる範囲で） |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| D6 | 同一 dir に `config.json` と `config.jsonc` | エラー（両方ある旨） |

---

## 2. ロード・検証（`ProjectConfig`）

#### テスト：正常系

| # | 入力 | 期待 |
|---|------|------|
| L1 | 通常 JSON | フィールドが期待どおり |
| L2 | JSONC（コメント・末尾カンマ） | 同様に成功（#155） |
| L3 | `override_project_name` のみ | override あり、フィルタ未設定 |
| L4 | `not_save_files` のみ | 否定フィルタ |
| L5 | `save_files` のみ | 肯定フィルタ |
| L6 | `override_project_name: ""` | override 未設定 |
| L7 | `save_files: []` かつ `not_save_files` なし | フィルタ未設定 |

#### テスト: 異常系

| # | 入力 | 期待 |
|---|------|------|
| L8 | `save_files` と `not_save_files` が両方非空 | エラー |
| L9 | 壊れた JSON/JSONC | エラー |

---

## 3. 候補数と選択

| # | 候補数 | 期待 |
|---|--------|------|
| S1 | 0 | 設定なし。`Resolve(wd, "")`、フィルタなし |
| S2 | 1 | 対話なしでロード |
| S3 | 2 以上・TTY | Select で選んだパスをロード（単体は選択関数をスタブ） |
| S4 | 2 以上・非 TTY / 選択失敗 | 非ゼロ相当のエラー（メッセージに複数候補・選択不能の旨） |

---

## 4. プロジェクト名配線（push / pull / clean）

#134 の Name 決定に override を乗せる。

| # | 条件 | 期待 Name |
|---|------|-----------|
| N1 | 設定なし、cwd=`/repo/app`、`.git` は `/repo` | `repo`（#134 退行） |
| N2 | `/repo/.bwsf/config.json` に `override_project_name: "my-api"`、cwd=`/repo/app` | `my-api` |
| N3 | override 空文字の設定ファイルのみ | #134 どおり（未設定） |

Dir / Warn は [`project/git_root.md`](../project/git_root.md) から退行しない。

---

## 5. ファイル選択フィルタ（core）

基盤通過後の basename フィルタ。push / clean /（管理対象一覧表示があれば）同一規則。

共通の Dir 内ファイル例: `.env`, `.env.local`, `.env.example`, `foo.tfvars`, `x.auto.tfvars`

| # | フィルタ | 期待に含まれる | 含まれない |
|---|----------|----------------|------------|
| F1 | なし | `.env`, `.env.local`, `foo.tfvars`, `x.auto.tfvars` | `.env.example` |
| F2 | `not_save_files: [".env.local", "*.auto.tfvars"]` | `.env`, `foo.tfvars` | `.env.local`, `x.auto.tfvars`, `.env.example` |
| F3 | `save_files: [".env", ".env.production"]` で実在が `.env` と `.env.local` | `.env` のみ | `.env.local` ほか |
| F4 | `save_files: ["secrets.yaml"]` のみ（基盤外） | （基盤通過ゼロなら）空 | `secrets.yaml` は取り込まない |

---

## 6. コマンド境界（退行）

| # | 期待 |
|---|------|
| C1 | `list` / `setup` / `config` は `.bwsf/config.*` を読まない（探索・Select を呼ばない） |
| C2 | 設定 0 件時、push/pull/clean の Name/Dir/Warn は #134 仕様のまま |

---

## 7. 対象外（本仕様でテストしない）

- グローバル設定への `save_files` / `not_save_files`（#177）
- 基盤ルール（`isManagedFileName`）自体の設定ファイル化
- Vault Secure Note の JSONC 化
- フック配置（#151）の実装（`.bwsf/` 集約は配置方針のみ）

---

## 実装メモ（仕様外・参考）

```text
cwd → FindProjectConfigPaths
  → 0/1/N (Select if N)
  → Load ProjectConfig (JSONC) + validate exclusive lists
  → project.Resolve(wd, override)
  → findEnvFiles: isManagedFileName → save_files | not_save_files glob
```
