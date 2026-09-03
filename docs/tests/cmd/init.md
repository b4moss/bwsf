# cmd: `bwsf init`（Issue #193 / v0.20.0 §5）

対象パッケージ: `app/src/cmd` / `app/src/config`（プロジェクト設定の書き込み）  
Issue: [#193](https://github.com/b4moss/bwsf/issues/193)  
製品正本: [`../specs/v0.20.0-multi-host.md`](../specs/v0.20.0-multi-host.md) §5

関連:

- グローバル設定の存在 → [`../config/global_v2.md`](../config/global_v2.md)
- プロジェクト `host` の読み取り解決 → [`../config/host_resolve.md`](../config/host_resolve.md)
- `save_files` + `!` → [`../config/save_files_bang.md`](../config/save_files_bang.md)
- プロジェクト探索・override → [`../config/project_local.md`](../config/project_local.md)
- setup 対話パターン → [`setup_v2.md`](./setup_v2.md)（init はプロジェクト側のみ。host **追加**はしない）

---

### 前提（本仕様で固定）

| # | 方針 |
|---|------|
| I1 | 生成先は常に **カレントディレクトリ** の `.bwsf/config.jsonc`（git root へ自動上昇して書かない） |
| I2 | グローバル設定ファイルが **無い** ときはエラー（先に `bwsf setup`）。**`hosts: []` のグローバルは「有り」** |
| I3 | `host` は **任意**。グローバル `hosts[]` から選ぶかスキップ。書いた場合は id 参照のみ（接続情報はグローバル側） |
| I4 | `hosts` が空のとき、host プロンプトは出さず（または即スキップ）、プロジェクトに `host` キーを書かない |
| I5 | `save_files` は対話で「設定 / 未設定」。設定時は glob 入力（否定は `!`）。未設定ならキー省略 |
| I6 | `override_project_name` も任意（設定 / 未設定）。空文字は未設定としてキー省略 |
| I7 | 既存の `.bwsf/config.json` または `.bwsf/config.jsonc` がある場合は上書き確認。拒否なら書き込まない。**`--yes`** なら確認スキップ |
| I8 | 保存は常に `.jsonc`。既存が `.json` のみでも `.jsonc` を書き、旧 `.json` は削除（グローバル `SaveConfig` と同型） |
| I9 | Keychain / unlock / lock / auth は触らない |
| I10 | 非対話でもテスト可能にする（フラグ名は実装で固定し本表を更新してよい） |

生成例（`host` あり）:

```jsonc
{
  "host": "default",
  "override_project_name": "my-api",
  "save_files": [".env*", "!.env.local"]
}
```

`host` 無し:

```jsonc
{
  "override_project_name": "my-api",
  "save_files": [".env*", "!.env.local"]
}
```

---

## 1. グローバル設定ゲート

#### テスト：正常系

| # | 条件 | 期待 |
|---|------|------|
| G1 | グローバルあり・`hosts: []` | init 続行可（host キーは書かない） |
| G2 | グローバルあり・hosts 非空 | init 続行可 |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| G3 | グローバル設定ファイル無し（`LoadConfig` が nil） | エラー。`.bwsf/` を作らない。setup を案内 |
| G4 | グローバルが両方（`.json` と `.jsonc`）等でロード失敗 | エラー。プロジェクトを書かない |

---

## 2. host 対話

対話関数はスタブ可能。初期: グローバルに `default`（is_default）と `work`。

#### テスト：正常系

| # | 操作 | 期待 |
|---|------|------|
| H1 | 一覧から `work` を選択 | 出力 JSONC に `"host": "work"` |
| H2 | スキップ | `host` キーが無い |
| H3 | グローバル `hosts: []` | host プロンプト無し（または自動スキップ）。`host` キー無し |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| H4 | 非対話で存在しない `--host` id | エラー。ファイルを書かない |

---

## 3. `save_files` / `override_project_name`

#### テスト：正常系

| # | 操作 | 期待 |
|---|------|------|
| F1 | save_files「設定」・`[".env*", "!.env.local"]` | 当該配列が保存される |
| F2 | save_files「未設定」 | `save_files` キー無し |
| O1 | override「設定」・`my-api` | `"override_project_name": "my-api"` |
| O2 | override「未設定」 | キー無し |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| F3 | 設定を選んだが glob が空（肯定も否定も無し） | 未設定扱いかエラーかを実装で固定。空配列を意味なく書かない |

---

## 4. 書き込みパスと上書き

cwd を一時ディレクトリにして検証する。

#### テスト：正常系

| # | 条件 | 期待 |
|---|------|------|
| W1 | `.bwsf` 無し | `.bwsf/config.jsonc` が作成される |
| W2 | 既存 `.jsonc`・確認で Yes | 上書きされる |
| W3 | 既存 `.jsonc`・`--yes` | 確認無しで上書き |
| W4 | 既存 `.json` のみ・確認 Yes | `.jsonc` が書き込まれ、`.json` は削除される |
| W5 | 親に `.git` があっても | **cwd** の `.bwsf/` に書く（git root には書かない） |

#### テスト: 異常系

| # | 条件 | 期待 |
|---|------|------|
| W6 | 既存あり・確認で No | 既存ファイル不変。exit non-zero または「キャンセル」成功を実装で固定（推奨: 非ゼロ／明示キャンセル） |
| W7 | 書き込み失敗（権限等） | エラー。中途半端な破壊を最小化 |

---

## 5. コマンド登録・非対話

#### テスト：正常系

| # | 操作 | 期待 |
|---|------|------|
| R1 | root サブコマンド | `init` が含まれる |
| N1 | 非対話: グローバルあり・`--skip-host`・save_files / override 未指定・`--yes` | `host` 無しの最小 `.jsonc` が書ける |
| N2 | 非対話: `--host default` + `--save-files '.env*,!.env.local'` | 対応キーが保存される |

フラグ名は実装で固定し、確定後に本表を更新する。

---

## 対象外

- Keychain / `unlock` / `lock`（#153）
- `auth login` / `logout`（#174）
- グローバル `hosts[]` の追加・更新（それは `setup`）
- プロジェクト設定の **読み取り** 契約の再定義（既存 [`project_local.md`](../config/project_local.md) / [`save_files_bang.md`](../config/save_files_bang.md)）
