# ドキュメント差分洗い出し（v0.15.0 / 正本 = ソースコード）

作業ブランチ: `enhance-docs-0.15.0`  
基準コミット: `main` @ `1a42efb`（Version `0.15.0`）

方針: **実装が正**。ドキュメント側を後続コミットで合わせる。本ファイルは洗い出し用チェックリスト（修正完了後に削除 or `_archived` へ移してよい）。

---

## A. 事実誤り（コードと矛盾・優先して直す）

| ID | 場所 | ドキュメントの主張 | コード上の事実 | 根拠 |
|----|------|-------------------|----------------|------|
| A1 | README（EN/JA）Usage — Push | 同名 Note があるとき **上書き確認する** | 既存があれば **無確認で `UpdateNoteItem`** | `app/src/core/core.go` `PushEnvCore` |
| A2 | `docs/site/{en,ja}/guide/commands.md` — push Behavior #3 | prompts to overwrite / 上書きを確認 | 同上（確認なし） | 同上 |
| A3 | `docs/site/{en,ja}/guide/commands.md` — list Output | `Projects in Bitwarden:` + `•` 付き箇条書き | **プロジェクト名を 1 行ずつ**（装飾なし）。0 件時のみ `No items found in {folder} folder` | `app/src/cmd/list.go` |
| A4 | `docs/site/{en,ja}/guide/dev-loadmap.md` | `bwsf clean` が未完了リストにある | **実装済み**（v0.14.0） | `app/src/cmd/clean.go` |
| A5 | `docs/todo.md` | Docs サイト未着手 / E2E(VW) 未着手（チェック無し） | Docs サイト運用中。ローカル VW スモーク実装済み | `docs/site/`、`docs/smoke.md`、#71/#108–#110 |
| A6 | `docs/site/.vitepress/config.ts` | GitHub / editLink / changelog が `b4m-oss/bwenv` | リポジトリは **`b4moss/bwsf`**（旧 `b4m-oss/bwsf`） | 現行 remote / README clone URL |
| A7 | `docs/site/{en,ja}/cookie-policy.md` 等 | `github.com/b4m-oss/bwenv` リンク | 同上 | 要一括置換 |

---

## B. 管理対象ファイル記述の欠落・偏り（v0.15.0）

正本: `isManagedFileName` — `.env*` **または** 末尾 `.tfvars` / `.tfvars.json`。名前に `.example` を含むものは除外（`app/src/core/core.go`）。

| ID | 場所 | 現状 | あるべき方向 |
|----|------|------|--------------|
| B1 | site commands — push/pull/clean 本文 | 概要表は managed files だが、本文・Behavior・`--from` 説明が `.env*` 中心 | 管理対象（`.env*` / `*.tfvars` / `*.tfvars.json`）に統一 |
| B2 | site features（EN/JA） | `.env` のみ列挙 | tfvars を追記。`.example` 除外は維持 |
| B3 | site getting-started / index（EN/JA） | `.env` / `dotenvs` 固定の説明 | managed files + フォルダ名は設定可能（default `dotenvs`） |
| B4 | README Overview 表 | `setup` が無い。Usage 見出しが「.env」寄り | `setup` 追加。表記を managed files に |
| B5 | FAQ（tfvars 以外の導入文） | 「.env を保存するコマンド」 | 管理対象ファイル全般、と整合 |

---

## C. 未ドキュメントの実装機能（書いてよい／書くべき）

| ID | 実装 | 未記載のドキュメント |
|----|------|----------------------|
| C1 | `bwsf setup --host-type/--url/--email/--password/--yes`（非対話） | site commands / README。スモークが利用（`scripts/run-smoke.sh`） |
| C2 | `bwsf config show` | site は記載済み。README Usage 詳細・features / getting-started は薄い |
| C3 | `bwsf clean` | commands にあり。features / getting-started / README Usage に無い or 薄い |
| C4 | pull はファイル単位の上書き確認あり | README は記載。commands EN/JA は概ね OK（管理対象表記は B1） |
| C5 | `folder_name` / `ResolveFolderName` → default `dotenvs` | 多くは記載済み。index の「専用 `dotenvs` フォルダ」は「デフォルト名」に修正余地 |

---

## D. ロードマップ系ドキュメントの陳腐化（事実更新）

| ID | 場所 | 内容 |
|----|------|------|
| D1 | `docs/todo.md` | **廃止予定**（本ブランチで Issue 転記後に削除）。下記「todo 転記」参照 |
| D2 | `dev-loadmap.md`（EN/JA） | clean を完了扱いに。未実装は `bwsf.config.json` / `.git` ルート検証のみ残すのが実態と一致 |
| D3 | `docs/todo.md` 「GitHub Pages にドキュメントサイト」 | 実質完了（#71、https://bwsf.oss.b4m.jp/）。新規 Issue は作らず転記表で閉じる |

---

## E. 意図的に「将来」のまま残してよい記述

コード未実装のため、**現状維持でよい**（別 Issue 化は任意）:

| 内容 | 主な記載場所 | 関連 |
|------|--------------|------|
| Windows 未対応 / planned | FAQ, installation, index, README | 転記 Issue（After v1.0） |
| 特定ファイル除外（ignore）未実装 | FAQ | loadmap の `bwsf.config.json` と重複。Issue 未作成なら後続で切り出し可 |
| `bwsf.config.json` / `.git` ルート検証 | loadmap | todo.md 外。本作業の todo 転記対象外 |
| logout は `bw` 側 | FAQ | #53 完了まで妥当 |
| Homebrew は自前 tap | installation / README | 本家登録は転記 Issue |

---

## F. 内部ドキュメント（任意・優先度低）

| ID | 場所 | メモ |
|----|------|------|
| F1 | `docs/TEST.md` | 単一 `.env` 時代の文言が残存。charter 新規は `docs/tests/`。整理は別タスクでも可 |
| F2 | `CONTRIBUTING.md` | 大枠は現行と一致。todo.md 参照は無し |
| F3 | philosophy（EN/JA） | 思想ドキュメント。製品コマンド表ではないため、tfvars 追記は任意 |

---

## 推奨修正順（次工程）

1. A1–A3, A6–A7（誤り・旧 URL）
2. B1–B5 + C1–C3（v0.15 の製品説明を揃える）
3. D2 + `todo.md` 削除（転記完了後）
4. F 系は余力

---

## todo.md → Issue 転記結果

転記日: 2026-08-13（ブランチ `enhance-docs-0.15.0`）  
元ファイル: `docs/todo.md`（削除はドキュメント修正コミットと同時でよい）

| todo.md 項目 | 状態 | 転記先 / 対応 |
|--------------|------|----------------|
| `[x] dotenvsディレクトリの別名設定` | 完了済 | #54（closed） |
| `[ ] GitHub Pagesにドキュメントサイトを作る` | **実質完了**（未チェックのまま残存） | #71（closed）＋運用中 https://bwsf.oss.b4m.jp/ 。**新規 Issue なし** |
| `[ ] Homebrew本家への登録` | 未着手 | **#129**（Until v1.0.0） |
| `[ ] bwコマンド依存からの脱却` | 進行中 | #53（Until v1.0.0）※ v1.0 直前着手方針 |
| `[ ] enhance: 毎回マスターパスワードを入れるのがめんどくさい` | 未着手 | **#130**（After v1.0.0） |
| `[ ] Windows版ビルド＆Chocolaty公開` | 未着手 | **#131**（After v1.0.0）※ Chocolatey 表記で正規化 |
| `[ ] E2Eテスト（Vaultwarden）` | **ローカルは完了** | #108 / #109 / #110（closed）。リモート残は #112。**新規 Issue なし** |

マイルストーン新設: **After v1.0.0**（#130 / #131 用）

### loadmap にあって todo.md に無い未実装（参考・本転記の対象外）

- `bwsf.config.json`（`ignore` / `override_project_name`）
- `.git` によるプロジェクトルート検証  

→ 必要なら別途 Issue 化（ドキュメント整理の後続で可）。
