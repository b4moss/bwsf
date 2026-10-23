# はじめに

## bwsfとは？

**bwsf** は、[Bitwarden](https://bitwarden.com/) を使用して `.env*` や Terraform の `*.tfvars` / `*.tfvars.json` といったプロジェクトファイルを安全に管理する CLI ツールです。

メールや Slack のような安全でないチャネルで秘密情報を共有する代わりに、bwsf を使えば Bitwarden ボールトに保存してチーム間で同期できます。

## 前提条件

bwsf を使用する前に、以下を準備してください：

1. **Bitwarden アカウント** — [Bitwarden Cloud](https://bitwarden.com/) またはセルフホスト / Vaultwarden
2. **Personal API Key** — アカウント設定 → セキュリティ → キー
3. **OS の秘密保管** — macOS Keychain または Linux secret service（API Key の保存先）

Bitwarden CLI（`bw`）は **不要** です（v0.20.0 で削除）。

## bwsf の仕組み

bwsf は管理対象ファイルを Bitwarden フォルダ（`target_section`、デフォルト名: `dotenvs`）内の **Secure Note** として保存します。構造のイメージは以下のとおりです：

```
Bitwarden Vault
└── dotenvs/                    # bwsf 用デフォルトフォルダ（ホストごとに変更可能）
    ├── my-web-app              # プロジェクト名 = git ルートのベース名（無ければ cwd）。override_project_name で上書き可
    │   ├── .env
    │   ├── .env.staging
    │   ├── .env.production
    │   └── terraform.tfvars
    └── another-project
        └── .env
```

::: info
デフォルトのフォルダ名は `dotenvs` です。`bwsf setup --folder <name>` で変更できます。名前変更では既存ノートは移動しません。
:::

## 初期設定

bwsf をインストールしたら:

```bash
bwsf setup                 # ホスト（スキップ可）+ 任意の save_files / folder
bwsf auth login             # Personal API Key を保存し vault を unlock
```

`setup` は `~/.config/bwsf/config.jsonc` に書き込みます。`auth` は Personal API Key を OS の秘密保管に保存します。

保存内容はいつでも次で確認できます：

```bash
bwsf config show
```

`push` / `pull` / `list` / `clean` では、まず OS 秘密保管の `vault_unlock` を restore します。失敗時のみ **マスターパスワード** を求め、プロセスメモリ上で unlock し、必要なら `vault_unlock` を更新します。コマンド終了時にメモリ上の鍵は破棄されますが、Keychain の `vault_unlock` は `lock` / `auth logout` まで残ります。

## 次のステップ

- [インストール](/ja/guide/installation) - お使いのプラットフォーム向けのインストール手順
- [コマンド](/ja/guide/commands) - 利用可能なすべてのコマンドを学ぶ
- [アップグレード](/ja/guide/upgrade) - 破壊的変更（v0.20.0 マルチホスト）
- [OKF docs](https://github.com/b4moss/bwsf/blob/main/docs/index.md) - リポジトリの知識ハブ（pillar: `docs/README.md` · 製品仕様: `docs/specs/`）
