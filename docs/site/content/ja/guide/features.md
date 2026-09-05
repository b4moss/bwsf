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

任意のフィルタは、グローバル（`~/.config/bwsf/config.jsonc`）またはプロジェクト（`.bwsf/config.jsonc`）設定の `save_files` で指定します。glob に `!` 接頭辞を付けると除外になります。プロジェクトの `save_files` はグローバルを完全に上書きします。

## 管理対象ファイルのプロジェクトへの適用

```bash
# cd /path/to/your/project_root
bwsf pull
```

Bitwarden 側に保存されている当該プロジェクトの管理対象ファイルを、プロジェクトルートへ書き戻します。ローカルに同名ファイルがある場合は、ファイル単位で上書き確認します。

## API（のみ）

v0.20.0 から、bwsf は Bitwarden **API** のみを使用します（Personal API Key）。典型フロー:

```bash
bwsf setup
bwsf auth
bwsf push   # マスターパスワードで unlock
```

## マルチホスト

グローバル設定の `settings.hosts` に複数ホストを登録できます。`--host <id>`、プロジェクトの `host`、または `is_default` が付いたホストで選択します。

## ローカル設定の確認

```bash
bwsf config show
```

`~/.config/bwsf/config.jsonc` の値（hosts、`save_files`、メタデータ）を表示します。Bitwarden にはアクセスしません。

## ローカル管理対象ファイルの削除

```bash
bwsf clean
```

Bitwarden 側に一致するバックアップがあることを確認したうえで、ローカルの管理対象ファイルを削除します。

## Bitwarden を使ったマルチユーザー共有

Bitwarden 側では、ホストごとに設定可能なフォルダ（`target_section`、デフォルト: `dotenvs`）にノートとして保存されます。

`bwsf` をプロジェクトルートで実行したとき、そのルートフォルダ名がプロジェクト名になります。

そのフォルダを Bitwarden 上で他のユーザーと共有することで、管理対象ファイルを複数メンバーで共有できます。

詳しくは、[Bitwarden のドキュメント](https://bitwarden.com/resources/)をご覧ください。
