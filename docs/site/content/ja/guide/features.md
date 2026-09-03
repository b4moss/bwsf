# 主な機能

`bwsf` は、[Bitwarden](https://bitwarden.com/) を用いて `.env*` や Terraform の `*.tfvars` / `*.tfvars.json` といったプロジェクトファイルをセキュアに管理するためのヘルパーコマンドです。

## 管理対象ファイルの保存

```bash
# cd /path/to/your/project_root
bwsf push
```

このコマンドは、プロジェクトルートにある管理対象ファイルをまとめて Bitwarden に保存します。例:

- `.env`
- `.env.local`
- `.env.staging`
- `.env.production`
- `terraform.tfvars`
- `prod.auto.tfvars`
- `secret.tfvars.json`

名前に `.example` を含むファイル（例: `.env.local.example`、`terraform.tfvars.example`）は **保存されません**。

## 管理対象ファイルのプロジェクトへの適用

```bash
# cd /path/to/your/project_root
bwsf pull
```

Bitwarden 側に保存されている当該プロジェクトの管理対象ファイルを、プロジェクトルートへ書き戻します。ローカルに同名ファイルがある場合は、ファイル単位で上書き確認します。

## API バックエンド（デフォルト）

v0.19.0 からデフォルトバックエンドは **`api`**（Personal API Key）です。典型フロー:

```bash
bwsf setup
bwsf auth
bwsf push   # マスターパスワードで unlock
```

必要なら `bwsf backend --set bw` で Bitwarden CLI に戻せます。

## ローカル設定の確認

```bash
bwsf config show
```

`~/.config/bwsf/config.json` の値（ホスト種別、URL、メール、実効フォルダ名）を表示します。Bitwarden にはアクセスしません。

## ローカル管理対象ファイルの削除

```bash
bwsf clean
```

Bitwarden 側に一致するバックアップがあることを確認したうえで、ローカルの管理対象ファイルを削除します。

## Bitwarden を使ったマルチユーザー共有

Bitwarden 側では、設定したフォルダ（デフォルト: `dotenvs`）にノートとして保存されます。

`bwsf` をプロジェクトルートで実行したとき、そのルートフォルダ名がプロジェクト名になります。

そのフォルダを Bitwarden 上で他のユーザーと共有することで、管理対象ファイルを複数メンバーで共有できます。

詳しくは、[Bitwarden のドキュメント](https://bitwarden.com/resources/)をご覧ください。
