# project: `.git` ルート解決と push/pull/clean（Issue #134 / v0.16.0）

対象パッケージ:

- 主: `app/src/project`（新規。`FindGitRoot` / `Resolve`）
- 配線: `app/src/cmd` の `push` / `pull` / `clean`
- 触らない: `app/src/core` の `PushEnvCore` / `PullEnvCore` / `CleanEnvCore` シグネチャ（`fromDir` / `projectName` を受け取る契約は維持）

Issue: [#134](https://github.com/b4moss/bwsf/issues/134)  
関連（先送り）: [#133](https://github.com/b4moss/bwsf/issues/133) — `bwsf.config.json` / `override_project_name` / `ignore` は本仕様では **実装しない**。名前解決の第1段だけ空スロットとして残す。

合意正本（#134）:

| # | 合意 |
|---|------|
| Q1 | ハイブリッド … プロジェクト名の基準は git ルート。`--from` / `--output` 指定時は **ファイル操作ディレクトリだけ**そのパス（名前はルート／将来の config 由来のまま） |
| Q2 | 祖先に `.git` が無いときは **警告を出して cwd にフォールバック**（エラーで止めない） |
| Q3 | モノレポは **親リポジトリの `.git` がルートの正**。パッケージごとの Note 名差は #133 |
| Q7 | ファイル探索のデフォルトディレクトリは **git ルート**。無ければ **cwd**（Q2 と同じ） |

### プロジェクト名の解決順（#133 と共用・本 Issue では 1 を常に空）

1. `override_project_name`（本 Issue では呼び出し側が常に空文字を渡す＝スキップ）
2. なければ git ルートの basename
3. git ルートが無ければ cwd の basename（現行互換）

### ファイル探索ディレクトリ

1. CLI の `--from` / `--output` が **明示指定**されていればそれ
2. なければ git ルート
3. git ルートが無ければ cwd

---

### 前提（本仕様で固定）

| # | 方針 |
|---|------|
| G1 | `.git` 探索は純 Go。`git` CLI に依存しない |
| G2 | `.git` が **ディレクトリ**でも **ファイル**（worktree / 一部 submodule）でも、その親ディレクトリを git ルートとする |
| G3 | `--from` / `--output` の Cobra デフォルト文字列 `"."` は残してよい。実際の既定切替は `Flags().Changed(...)` で判定する（未変更＝git ルート／cwd フォールバック、変更済み＝フラグ値） |
| G4 | フォールバック警告は既存 `utils.Warningln`（または同等の黄色警告）を **1 回**。エラー終了にしない |
| G5 | `list` / `setup` / `config` / グローバル `~/.config/bwsf` は本 Issue の対象外 |
| G6 | #133 の設定ファイル探索・対話選択・`ignore` glob は本 Issue 外。`Resolve` が `overrideProjectName string` を受け取れることだけを契約に含める |

---

### `FindGitRoot(start string) (root string, ok bool)`

`start`（絶対パス推奨。相対なら実装側で絶対化してよい）から親へ辿り、最初に見つかった `.git` エントリの親を返す。見つからなければ `ok=false`。

#### テスト：正常系

- `repo/.git/`（ディレクトリ）があり、`start=repo/subdir` のとき `root=repo`, `ok=true`
- `repo/.git` が **ファイル**（worktree 風）で、`start=repo/pkg` のときも `root=repo`, `ok=true`
- `start` がちょうど git ルート自身でも `root` はそのディレクトリ

#### テスト: 異常系 / 境界

- `.git` が祖先に無い → `ok=false`, `root` は空（または未使用）
- ファイルシステムルートまで辿ってもパニックせず終了する
- （任意）`start` が存在しないパスでも、親辿りが可能な範囲で落ちずに `ok=false` またはエラー方針を実装で固定しテストする（本仕様は「パニックしない」を最低条件とする）

---

### `Resolve(cwd, overrideProjectName string) (Context, error)`

`Context` が少なくとも持つもの:

- `GitRoot string`（無ければ空）
- `ProjectName string`
- `WorkDir string`（ファイル探索のデフォルトディレクトリ）
- `UsedCwdFallback bool`（git 無しで cwd に落としたとき true）

#### テスト：正常系 — 名前

- git ルートあり・`overrideProjectName=""` → `ProjectName` は git ルートの basename（サブディレクトリを `cwd` にしてもルート名）
- git ルートあり・`overrideProjectName="my-api"` → `ProjectName` は `my-api`（#133 予約スロット。本 Issue の cmd からは渡さないがヘルパ単体では検証する）
- git ルートなし・`overrideProjectName=""` → `ProjectName` は `cwd` の basename、`UsedCwdFallback=true`

#### テスト：正常系 — WorkDir

- git ルートあり → `WorkDir` は git ルート（`cwd` がサブディレクトリでもルート）
- git ルートなし → `WorkDir` は `cwd`、`UsedCwdFallback=true`

#### テスト: 異常系

- `cwd` 取得不能相当をヘルパに渡すケースは cmd 側で先にエラーにしてよい。`Resolve` に空 `cwd` を渡した場合の扱いは実装で固定し、パニックしないこと

---

### cmd: `push` / `clean`（`--from`）

- `Resolve(wd, "")` で `ProjectName` / デフォルト作業ディレクトリを得る
- `UsedCwdFallback` なら警告を 1 回出して続行
- `Flags().Changed("from") == false` → 作業ディレクトリは `Context.WorkDir`
- `Changed("from") == true` → 作業ディレクトリはフラグ値。**`ProjectName` は変えない**
- 得た `projectName` と作業ディレクトリを既存 core に渡す

#### テスト：正常系

- サブディレクトリを cwd にし、`--from` 未指定: core に渡る `projectName` がリポジトリ basename、`fromDir` が git ルート（モック／ヘルパ経由で検証してよい）
- `--from` に別パスを明示: `fromDir` はそのパス、`projectName` は依然 git ルート basename
- `.git` 無しツリー: 警告が出て、`projectName` / `fromDir` は cwd 基準（現行互換）

#### テスト: 異常系 / 退行

- `Getwd` 失敗は現行どおりエラー終了
- core の「管理対象 0 件」等のエラー伝播は退行として維持（ディレクトリが変わった結果 0 件になるのは仕様どおり成功扱いにしない）

---

### cmd: `pull`（`--output`）

- `push` / `clean` と同じ名前解決
- `Flags().Changed("output")` で出力ディレクトリを切替（未指定 → `WorkDir`、指定あり → フラグ値）
- 名前は常に `Context.ProjectName`

#### テスト：正常系

- サブディレクトリ cwd・`--output` 未指定: 出力先は git ルート、Note 名はリポジトリ basename
- `--output` 明示: 出力先のみフラグ値、Note 名は git ルート basename のまま
- `.git` 無し: 警告＋ cwd フォールバック

#### テスト: 異常系

- リモートに該当プロジェクトが無いときのエラーは現行どおり（名前解決後の名前で検索する）

---

### 退行

#### テスト：正常系

- git ルート直下を cwd にした `push` / `pull` / `clean` は、名前・ディレクトリともに現行（ルート basename / `.` 相当）と結果が一致する
- 既存の core 単体テスト（モックに明示的 `fromDir` / `projectName` を渡すもの）は **変更不要で通る**

#### テスト: 異常系

- `list` / `setup` / `config show` の挙動が本変更で壊れない（対象外コマンドの退行スモークでよい）

---

### 実装メモ（仕様外だがテスト配置の指針）

- ユニットは `app/src/project` を厚くする
- cmd の `Changed` 分岐は、可能なら作業ディレクトリ決定を小さい純関数／テスト可能なヘルパに寄せるとよい
- `docs/tests/*.md` 以外のプロダクト docs 更新は本 Issue の必須ではない
