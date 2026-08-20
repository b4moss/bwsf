# インストール

## 動作環境

### 対応OS

| OS | 状態 |
|---|---|
| macOS | ✅ 対応 |
| Linux | ✅ 対応 |
| Windows | 🚧 計画中 |

### 依存関係

`bwsf`の利用とインストールには、以下が必要です。

- **Bitwarden CLI (`bw`)**
- **Homebrew**
- Bitwardenのアカウント

### Bitwarden CLIのインストール

```bash
# macOS
brew install bitwarden-cli

# Linux (Snap)
sudo snap install bw

# npm (クロスプラットフォーム)
npm install -g @bitwarden/cli
```

その他のオプションについては、[公式 Bitwarden CLI ドキュメント](https://bitwarden.com/help/cli/#download-and-install)を参照してください。

### Homebrewのセットアップ

詳しくは、[Homebrewの公式サイト](https://brew.sh/)をご覧ください。

- [macOS](https://brew.sh/)
- [Linux](https://docs.brew.sh/Homebrew-on-Linux)

### Bitwardenアカウントのセットアップ

Bitwardenには、2種類のホスティング形式があります

- **[Bitwarden Cloud](https://bitwarden.com/)**: Bitwardenが、公式のホスティングしているSaaSサービスです
- **Bitwardenセルフホスト**: BitwardenはOSSとして公開されており、セルフホストすることも可能です

`bwsf`は、このどちらにも対応しています。

[Bitwardenのアカウントの作り方は、こちらのドキュメント](https://bitwarden.com/help/create-bitwarden-account/)を参考にして下さい。

## bwsf のインストール

さて、いよいよ`bwsf`のインストールです。

### macOS

```bash
brew tap b4m-oss/tap && brew install bwsf
```

### Linux

```bash
brew tap b4m-oss/tap && brew install bwsf
```

## インストールの確認

```bash
bwsf -v
# bwsf version x.x.x
```

## 初期設定

インストール後、セットアップコマンドを実行して Bitwarden 接続を設定します。

```bash
bwsf setup
```

以下の入力を求められます：
1. Bitwarden サーバー URL（Bitwarden Cloud の場合は空欄）
2. Bitwarden のメールアドレス
3. マスターパスワード

----

以上で初期設定は終了です。お疲れ様でした！

