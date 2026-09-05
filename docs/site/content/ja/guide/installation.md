# インストール

## 動作環境

### 対応OS

| OS | 状態 |
|---|---|
| macOS | ✅ 対応 |
| Linux | ✅ 対応 |
| Windows | 🚧 計画中 |

### 依存関係

- Bitwarden アカウント（Cloud またはセルフホスト / Vaultwarden）
- **Personal API Key**（アカウント設定 → セキュリティ → キー）
- OS の秘密保管（**macOS Keychain** / **Linux secret service**）
- **Homebrew**（bwsf 本体のインストール用）

Bitwarden CLI（`bw`）は **不要** です（v0.20.0 で廃止）。

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
# bwsf version 0.20.0
```

## 初期設定

```bash
bwsf setup                 # ホスト（スキップ可）+ 任意の save_files / folder
bwsf auth login             # Personal API Key を OS 秘密保管へ + unlock
```

`setup` はマスターパスワードでのログインを行いません。認証は `bwsf auth login` です（API Key 保存と vault unlock）。その後の `push` / `pull` / `list` は `vault_unlock` の restore、または再プロンプトになります。設定は `~/.config/bwsf/config.jsonc` に保存されます。

----

以上で初期設定は終了です。お疲れ様でした！
