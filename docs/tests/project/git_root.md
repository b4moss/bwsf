# 振る舞い仕様: `.git` 有無によるプロジェクト解決（Issue #134 / v0.16.0）

Issue: [#134](https://github.com/b4moss/bwsf/issues/134)  
対象コマンド: `push` / `pull` / `clean`（`list` / `setup` / `config` は対象外）  
関連先送り: [#133](https://github.com/b4moss/bwsf/issues/133)（`override_project_name` は本仕様では常に未使用）

本ドキュメントの目的は実装詳細ではなく、**これまで（cwd 基準）とこれから（git ルート基準）で何が変わるかを固定し、その差分をテストで確認できるようにすること**。

---

## 1. 観測する2値

各シナリオで、コマンドが core に渡す（＝ユーザー効果として現れる）次を固定する。

| 記号 | 意味 |
|------|------|
| **Name** | Bitwarden Note 名（プロジェクト名） |
| **Dir** | 管理対象ファイルの読み書きディレクトリ（push/clean の探索元、pull の出力先） |
| **Warn** | `.git` 無しフォールバック時の警告を 1 回出すか |

コマンドとフラグの対応:

| コマンド | Dir を決めるフラグ |
|----------|-------------------|
| `push` / `clean` | `--from` |
| `pull` | `--output` |

フラグは **明示指定されたときだけ** Dir を上書きする（Cobra デフォルト `"."` のまま未タッチなら「未指定」）。

---

## 2. 現行（変更前）の基準振る舞い

変更前は常に次だった。

| 条件 | Name | Dir（フラグ未指定） | Warn |
|------|------|---------------------|------|
| 任意の cwd | `basename(cwd)` | `cwd`（`.`） | なし |
| フラグ明示 | `basename(cwd)` のまま | フラグ値 | なし |

サブディレクトリから実行すると **Name も Dir もサブディレクトリ基準**だった。これが #134 で変える主対象。

---

## 3. 変更後の決定表（本仕様で FIX）

共通前提の例:

```text
/repo/                 ← ここに .git がある場合の「git ルート」。basename = "repo"
/repo/app/             ← サブディレクトリ
/repo/.env             ← 管理対象はルートにある想定（Dir の話で使う）
/norepo/workdir/       ← .git が祖先に無いツリー。basename(workdir) = "workdir"
```

### 3.1 `.git` がある場合

| # | cwd | フラグ | Name | Dir | Warn | 現行との差分 |
|---|-----|--------|------|-----|------|--------------|
| A1 | `/repo` | 未指定 | `repo` | `/repo` | なし | **差分なし**（ルート直下では従来と同じ） |
| A2 | `/repo/app` | 未指定 | `repo` | `/repo` | なし | **変わる**（旧: Name=`app`, Dir=`/repo/app`） |
| A3 | `/repo/app` | `--from=/repo/app` または `--output=/repo/app` | `repo` | `/repo/app` | なし | **Name だけ変わる**（旧: Name=`app`。Dir は明示どおり） |
| A4 | `/repo/app` | `--from=/tmp/other` 等（ルート外パス） | `repo` | フラグ値 | なし | Name はルート、Dir だけフラグ（ハイブリッド） |

モノレポ（`/repo/packages/foo` に cwd）も A2/A3 と同じ。**親の `.git` が正**。パッケージ名を Note 名にしたい場合は #133（本仕様外）。

`.git` がディレクトリでもファイル（worktree）でも、上表の「git ルート」の意味は同じ。

### 3.2 `.git` が無い場合

| # | cwd | フラグ | Name | Dir | Warn | 現行との差分 |
|---|-----|--------|------|-----|------|--------------|
| B1 | `/norepo/workdir` | 未指定 | `workdir` | `/norepo/workdir` | **あり** | Name/Dir は現行と同じ。**警告だけが増える** |
| B2 | `/norepo/workdir` | フラグ明示 | `workdir` | フラグ値 | **あり** | 同上（Name/Dir は現行互換＋警告） |

`.git` 無しはエラーにしない（中断しない）。

---

## 4. テストで必ず押さえるシナリオ

実装の置き場（`project` パッケージ等）は問わない。**次の振る舞いが観測できればよい**。

### 4.1 `.git` あり — 差分の中心

| ID | 確認内容 |
|----|----------|
| T-A2 | サブディレクトリ cwd・フラグ未指定 → Name=`repo`, Dir=`/repo`（旧 `app` / `/repo/app` ではない） |
| T-A1 | ルート cwd・フラグ未指定 → Name=`repo`, Dir=`/repo`（退行: 従来と同じ） |
| T-A3 | サブディレクトリ cwd・フラグでサブディレクトリを明示 → Name=`repo`, Dir=フラグ値 |

### 4.2 `.git` なし — 互換＋警告

| ID | 確認内容 |
|----|----------|
| T-B1 | フラグ未指定 → Name/Dir は cwd 基準のまま、かつ警告が 1 回出る |
| T-B2 | フラグ明示 → Dir はフラグ、Name は cwd basename、警告が 1 回出る |

### 4.3 境界

| ID | 確認内容 |
|----|----------|
| T-W | `.git` がファイルでも T-A2 と同じ結果 |
| T-N | 祖先に `.git` が無いときエラー終了しない（B1） |

---

## 5. 明示的にテストしない／先送り

- プロジェクト設定 `.bwsf/config.(json|jsonc)` / `override_project_name` / `save_files` / `not_save_files`（#133 / `docs/tests/config/project_local.md`）
- `list` / `setup` / `config` の仕様変更
- core の MultiEnvData 形式そのもの（Dir/Name の渡し方が変わっても、渡した後の push/pull/clean ロジックは現行のまま）

---

## 6. 受け入れの一文

- **`.git` あり・サブディレクトリからフラグ未指定で実行すると、Note 名もファイル位置もリポジトリルート基準になる**（T-A2）。
- **`.git` なしでは従来どおり cwd 基準で動き、警告だけが増える**（T-B1）。
- **フラグは Dir だけを上書きし、Name は上のルールのまま**（T-A3）。
