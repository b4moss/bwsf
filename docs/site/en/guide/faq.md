# Frequently Asked Questions

## Q: I don't have a Bitwarden account

`bwsf` is a command that saves .env files to Bitwarden.

Therefore, **a Bitwarden account is required** to use it.

## Q: I'm self-hosting Bitwarden. Can I use bwsf?

Of course. `bwsf` supports self-hosted Bitwarden.

## Q: Is it available on Windows?

Unfortunately, not at this time. We're actively working on it.

Currently, only macOS/Linux is supported.

## Q: Can `bwsf` be used by multiple users?

`bwsf` itself is a command installed on individual development machines.

However, it uses Bitwarden as the backend for storing .env files.

Bitwarden allows you to set detailed user permissions, so by configuring users on the Bitwarden side, you can securely manage .env files among multiple users.

## Q: How are files stored in Bitwarden?

Bitwarden has several proprietary formats for storing sensitive information.

Among these, `bwsf` uses a format called "Secure Note".

The Secure Note title is the **project name**, and the content is stored in **JSON format**.

## Q: Can I save .env files for multiple environments?

Yes, you can.

- .env
- .env.local
- .env.staging
- .env.production

These files are saved all at once.

However, **files with `.example` in the filename** are NOT saved.

## Q: Can I exclude specific .env files?

For example:

- .env ← Save this
- .env.production ← Don't want to save this

Unfortunately, this feature has not been implemented yet.

## Q: Can I edit .env files on the Bitwarden host?

Yes, **if you can visually edit JSON files**.

However, if you're developing with multiple members, be careful as unintended pushes from other members may overwrite your changes, or incorrect pulls may occur.

The development team **does not officially recommend this**.

## Q: Does it support Terraform's `tfvars` files?

Yes. From v0.15.0, bwsf manages files ending in `.tfvars` or `.tfvars.json` the same way as `.env*` files (push / pull / list / clean).

Names that contain `.example` are excluded, just like `.env.example`.

## Q: How do I log out from Bitwarden?

Please do this on the `bw` command side. `bwsf` itself does not have login/logout functionality.

