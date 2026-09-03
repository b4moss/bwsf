---
title: Frequently Asked Questions
schemaRole: FAQPage
---

# Frequently Asked Questions

::faq-list
:::faq-item{question="I don't have a Bitwarden account"}
`bwsf` is a command that saves managed files (`.env*`, `*.tfvars`, `*.tfvars.json`) to Bitwarden.

Therefore, **a Bitwarden account is required** to use it.
:::

:::faq-item{question="I'm self-hosting Bitwarden. Can I use bwsf?"}
Of course. `bwsf` supports self-hosted Bitwarden.
:::

:::faq-item{question="Is it available on Windows?"}
Unfortunately, not at this time. We're actively working on it.

Currently, only macOS/Linux is supported.
:::

:::faq-item{question="Can `bwsf` be used by multiple users?"}
`bwsf` itself is a command installed on individual development machines.

However, it uses Bitwarden as the backend for storing managed files.

Bitwarden allows you to set detailed user permissions, so by configuring users on the Bitwarden side, you can securely manage those files among multiple users.
:::

:::faq-item{question="How are files stored in Bitwarden?"}
Bitwarden has several proprietary formats for storing sensitive information.

Among these, `bwsf` uses a format called "Secure Note".

The Secure Note title is the **project name**, and the content is stored in **JSON format**.
:::

:::faq-item{question="Can I save files for multiple environments?"}
Yes, you can. For example:

- `.env`
- `.env.local`
- `.env.staging`
- `.env.production`
- `terraform.tfvars`

These managed files are saved all at once.

However, **files with `.example` in the filename** are NOT saved.
:::

:::faq-item{question="Can I exclude specific managed files?"}
For example:

- `.env` ← Save this
- `.env.production` ← Don't want to save this

Add a project-local `.bwsf/config.json` (or `.jsonc`) and list globs under `not_save_files` (do not set `save_files` at the same time):

```json
{
  "not_save_files": [".env.production", "*.auto.tfvars"]
}
```

See issue [#133](https://github.com/b4moss/bwsf/issues/133) for details.
:::

:::faq-item{question="Can I edit files on the Bitwarden host?"}
Yes, **if you can visually edit JSON files**.

However, if you're developing with multiple members, be careful as unintended pushes from other members may overwrite your changes, or incorrect pulls may occur.

The development team **does not officially recommend this**.
:::

:::faq-item{question="Does it support Terraform's `tfvars` files?"}
Yes. From v0.15.0, bwsf manages files ending in `.tfvars` or `.tfvars.json` the same way as `.env*` files (push / pull / list / clean).

Names that contain `.example` are excluded, just like `.env.example`.
:::

:::faq-item{question="How do I log out from Bitwarden?"}
Please do this on the `bw` command side. `bwsf` itself does not have login/logout functionality.
:::
::
