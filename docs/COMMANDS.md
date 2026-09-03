# bwsf コマンド一覧（v0.19.0）

実装（`app/src/cmd/`）を正とした現状コマンドの棚卸しです。  
詳細な説明・ワークフローはドキュメントサイトの [Commands](https://bwsf.oss.b4m.jp/en/guide/commands) / [コマンド](https://bwsf.oss.b4m.jp/ja/guide/commands) を参照してください。

## 本体コマンド

| コマンド | 役割 | 主なフラグ |
|---|---|---|
| `bwsf setup` | ホスト/アカウント設定（`api` 時はログインなし。認証は `auth`） | `--folder` `--host-type` `--url` `--email` `--password` `--yes` |
| `bwsf auth` | Personal API Key の保存・認証（`api` バックエンド） | `--clear` |
| `bwsf backend` | バックエンドの表示 / 切替（`api` \| `bw`） | `--set` |
| `bwsf config show` | ローカル設定（`~/.config/bwsf/config.json`）の表示 | （なし） |
| `bwsf push` | 管理対象ファイルを Bitwarden へプッシュ | `--from` |
| `bwsf pull` | Bitwarden から管理対象ファイルを取得 | `--output` |
| `bwsf list` | 設定フォルダ内のプロジェクト一覧 | （なし） |
| `bwsf clean` | リモートバックアップ確認後にローカル管理対象を削除 | `--from` |

### 管理対象ファイル

- 名前が `.env` で始まるもの
- 末尾が `.tfvars` / `.tfvars.json` のもの
- 名前に `.example` を含むものは除外

### バックエンド

- デフォルト: **`api`**（Personal API Key + プロセス内 unlock）
- `bw` CLI 経路: `bwsf backend --set bw`

## 付帯（cobra 標準）

| コマンド / フラグ | 役割 |
|---|---|
| `bwsf -v` / `--version` | バージョン表示 |
| `bwsf help` | ヘルプ |
| `bwsf completion` | シェル補完スクリプト生成 |

## サブコマンド構造

```text
bwsf
├── setup
├── auth
├── backend
├── config
│   └── show
├── push
├── pull
├── list
├── clean
├── completion   # cobra 標準
└── help         # cobra 標準
```

`config` の子は現状 `show` のみです。
