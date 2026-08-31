---
title: よくある質問
schemaRole: FAQPage
---

# よくある質問

::faq-list
:::faq-item{question="Bitwardenのアカウントを持っていないのですが"}
`bwsf` は、管理対象ファイル（`.env*` / `*.tfvars` / `*.tfvars.json`）を Bitwarden に保存するコマンドです。

したがって、利用には**Bitwardenアカウントの取得が必要**です。
:::

:::faq-item{question="Bitwardenをセルフホストしています。使用可能ですか？"}
もちろんです。`bwsf`は、セルフホスティング型のBitwardenをサポートしています。
:::

:::faq-item{question="Windowsでも利用できますか？"}
残念ながら、現在はできません。鋭意開発中です。

現在は macOS/Linux のみ対応となっています。
:::

:::faq-item{question="`bwsf`は、複数のユーザーで利用することができますか？"}
`bwsf`そのものは、個人の開発端末にインストールされるコマンドです。

しかし、管理対象ファイルを保存するバックエンドとして、Bitwardenを採用しています。

Bitwardenは、細かなユーザー権限を設定することができますので、Bitwarden側でユーザー設定を行うことで、複数ユーザーによるセキュアな管理を行うことができます。
:::

:::faq-item{question="Bitwardenには、どのような形で保存されるのでしょうか？"}
Bitwardenには、機密情報の保存形式に、いくつかの独自形式を持っています。

この中で、`bwsf`は「ノート形式」という形式を使って保存します。

ノート形式のタイトルは**プロジェクト名**、ノートの項目に**JSON形式**で保存されます。
:::

:::faq-item{question="複数環境のファイルを保存できますか？"}
はい、可能です。例えば：

- `.env`
- `.env.local`
- `.env.staging`
- `.env.production`
- `terraform.tfvars`

これらの管理対象ファイルは、一括して保存されます。

一方、**ファイル名に `.example` を含むファイル** は、保存されません。
:::

:::faq-item{question="特定の管理対象ファイルだけ除外することはできますか？"}
例えば、

- `.env`　← これは保存する
- `.env.production` ←これは保存したくない

残念ながら、このような機能はまだ実装されていません。
:::

:::faq-item{question="Bitwardenのホスト上で、ファイルを編集することは可能ですか？"}
はい、**もしあなたがJSONファイルを目視で編集できる**のだとしたら、可能です。

しかし、複数メンバーで開発している場合、意図しない他のメンバーのpushで上書きされたり、間違ったpullを引き起こすこともあるため、注意が必要です。

開発チームとしては、**公式には推奨しません**。
:::

:::faq-item{question="Terraformの`tfvars`ファイルには対応していますか？"}
はい。v0.15.0 から、末尾が `.tfvars` または `.tfvars.json` のファイルを `.env*` と同様に管理できます（push / pull / list / clean）。

名前に `.example` を含むものは、`.env.example` と同様に除外されます。
:::

:::faq-item{question="Bitwardenからのログアウトはどうしたらいいですか？"}
`bw`コマンド側で行って下さい。`bwsf`自体には、ログイン・ログアウトの機能は備わっていません。
:::
::
