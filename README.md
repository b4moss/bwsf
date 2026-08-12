# repo-ruleset-src-repo

GitHub リポジトリ向けの **ruleset 定義**と、作成・適用用の CLI です。

Gitea への自動同期は対象外です。Organization ruleset も無料プランでは使えないため、**リポジトリ単位**で適用します。

他リポジトリへ git subtree で取り込む想定のため、Make のターゲット / 変数は `ruleset-*` / `RULESET_*` に名前空間化しています。

## 前提

- [`gh`](https://cli.github.com/)（対象リポの **admin** で `gh auth login` 済み）
- `jq`
- `make`（任意。scripts を直接実行してもよい）

### プラン制約（GitHub Free）

| 機能 | Free Organization |
|------|-------------------|
| Organization ruleset | 不可 |
| Repository ruleset | **public のみ** |
| private の ruleset / branch protection | Pro / Team 以上が必要 |

`make ruleset-create` のデフォルトは `RULESET_VISIBILITY=public` です。

## 定義ファイル

| ファイル | 対象 |
|----------|------|
| [`.github/rulesets/standard-branch-policy.json`](.github/rulesets/standard-branch-policy.json) | `main` / `develop` / `staging` / `release` / `dev-v*` |
| [`.github/rulesets/standard-tag-policy.json`](.github/rulesets/standard-tag-policy.json) | `refs/tags/v*` |

ブランチ側の要点:

- 削除禁止・force-push 禁止
- PR 必須（approve 数は 0）
- Repository admin（`RepositoryRole` id=5）は bypass 可
- `feature/*` は対象外（日常の直接 push を縛らない）

## 使い方

### 新規リポ作成 + ruleset 適用（推奨）

```bash
make ruleset-create RULESET_REPO=OWNER/new-app
# 同等:
#   ./scripts/create-repo-with-rulesets.sh OWNER/new-app --public
```

オプション例:

```bash
make ruleset-create RULESET_REPO=OWNER/new-app RULESET_VISIBILITY=public RULESET_CREATE_FLAGS="--clone --description demo"
```

subtree 配下から叩く例:

```bash
make -f path/to/repo-ruleset/Makefile ruleset-create RULESET_REPO=OWNER/new-app
```

### 既存リポへ後付け / 更新

```bash
make ruleset-apply RULESET_REPO=OWNER/existing-app
```

同名 ruleset があれば更新、なければ作成します。

### 適用確認

```bash
make ruleset-check RULESET_REPO=OWNER/existing-app RULESET_BRANCH=main
```

内部では `gh ruleset list` と `gh ruleset check` を実行します。

## 初回フロー（イメージ）

1. このリポジトリを clone する（定義とスクリプトのソース）
2. `gh auth login`（適用先リポの admin 権限があるアカウント）
3. `make ruleset-create RULESET_REPO=OWNER/new-app`
4. `make ruleset-check RULESET_REPO=OWNER/new-app` で確認

利用者の手作業は上記までです。GitHub UI での Rulesets 設定は不要です。

## 権限について

ruleset の作成・更新にはリポジトリ admin が必要です。  
Secret や bot は使いません。リポを作る人が、作成と同時に自分の `gh` 認証で適用する想定です。
