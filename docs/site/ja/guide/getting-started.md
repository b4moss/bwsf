# はじめに

## bwsfとは？

**bwsf** は、[Bitwarden](https://bitwarden.com/) を使用して `.env` ファイルを安全に管理する CLI ツールです。

メールや Slack のような安全でないチャネルで `.env` ファイルを共有する代わりに、bwsf を使えば Bitwarden ボールトに保存してチーム間で同期できます。

## 前提条件

bwsf を使用する前に、以下を準備してください：

1. **Bitwarden アカウント** - [Bitwarden Cloud](https://bitwarden.com/) またはセルフホスト Bitwarden サーバー
2. **Bitwarden CLI (`bw`)** - 公式の Bitwarden コマンドラインツール

### Bitwarden CLI のインストール

[公式 Bitwarden CLI インストールガイド](https://bitwarden.com/help/cli/#download-and-install)に従って、お使いのマシンに `bw` コマンドをインストールしてください。

インストールの確認：

```bash
bw --version
```

## bwsf の仕組み

bwsf は `.env` ファイルを Bitwarden ボールト内の `dotenvs` という特別なフォルダに保存します。構造は以下のようになります：

```
Bitwarden Vault
└── dotenvs/                    # bwsf 用の予約フォルダ
    ├── my-web-app/             # プロジェクト名（ディレクトリ名）
    │   ├── .env
    │   ├── .env.staging
    │   └── .env.production
    └── another-project/
        └── .env
```

::: info
デフォルトのフォルダ名は `dotenvs` です。`bwsf setup --folder <name>` で変更できます。名前変更では既存ノートは移動しません。
:::

## 初期設定

bwsf をインストールしたら、セットアップコマンドを実行します：

```bash
bwsf setup
```

これにより以下が設定されます：
- Bitwarden サーバー URL（セルフホストインスタンスの場合）
- Bitwarden アカウントの認証情報

## 次のステップ

- [インストール](/ja/guide/installation) - お使いのプラットフォーム向けのインストール手順
- [コマンド](/ja/guide/commands) - 利用可能なすべてのコマンドを学ぶ



