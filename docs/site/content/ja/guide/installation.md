# インストール

## 動作環境

### 対応OS

| OS | 状態 |
|---|---|
| macOS | ✅ 対応 |
| Linux | ✅ 対応 |
| Windows | 🚧 計画中 |

### 依存関係（デフォルト API バックエンド）

- Bitwarden アカウント（Cloud またはセルフホスト / Vaultwarden）
- **Personal API Key**（アカウント設定 → セキュリティ → キー）
- OS の秘密保管（**macOS Keychain** / **Linux secret service**）
- **Homebrew**（bwsf 本体のインストール用）

デフォルトの `api` バックエンドでは Bitwarden CLI（`bw`）は **不要** です。`bwsf backend --set bw` を使う場合のみインストールしてください。

[bw のインストール（任意）](https://bitwarden.com/help/cli/#download-and-install)

### Homebrewのセットアップ

詳しくは、[Homebrewの公式サイト](https://brew.sh/)をご覧ください。

- [macOS](https://brew.sh/)
- [Linux](https://docs.brew.sh/Homebrew-on-Linux)

### Bitwardenアカウントのセットアップ

Bitwardenには、2種類のホスティング形式があります

- **[Bitwarden Cloud](https://bitwarden.com/)**: Bitwardenが公式ホスティングしている SaaS
- **Bitwardenセルフホスト**: OSS としてセルフホストも可能（Vaultwarden 含む）

`bwsf`は、このどちらにも対応しています。

[Bitwardenのアカウントの作り方は、こちらのドキュメント](https://bitwarden.com/help/create-bitwarden-account/)を参考にして下さい。

## bwsf のインストール

### macOS

```bash
brew tap b4m-oss/tap && brew install bwsf
```

### Linux

```bash
brew tap b4m-oss/tap && brew install bwsf
```

### 特定バージョンのインストール

Homebrew tap の versioned formula は、「現行 minor の全 patch」と「1つ前の minor の最新 patch」のみ残します。それ以外は tap から削除されますが、[GitHub Releases](https://github.com/b4moss/bwsf/releases) からは引き続き取得できます。

```bash
brew tap b4m-oss/tap
brew install bwsf@0.18.0
```

versioned formula は keg-only のため、パスを通すか `brew link --force bwsf@0.18.0` が必要になる場合があります。

## インストールの確認

```bash
bwsf -v
# bwsf version 0.19.0
```

## 初期設定

```bash
bwsf setup                 # ホスト種別 / URL / メール（+ 任意で folder）
bwsf auth                  # Personal API Key を OS 秘密保管へ
```

デフォルトの API バックエンドでは、`setup` はマスターパスワードでのログインを行いません。認証は `bwsf auth` です。保管庫の unlock（マスターパスワード）は `push` / `pull` / `list` 実行時に行います。

任意: Bitwarden CLI バックエンドへ切り替え:

```bash
bwsf backend --set bw
```

----

以上で初期設定は終了です。お疲れ様でした！
