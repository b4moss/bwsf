# bwsf

[![CI](https://github.com/b4moss/bwsf/actions/workflows/test.yml/badge.svg)](https://github.com/b4moss/bwsf/actions/workflows/test.yml)
[![Coverage](https://img.shields.io/codecov/c/github/b4moss/bwsf)](https://codecov.io/gh/b4moss/bwsf)
[![Go Reference](https://pkg.go.dev/badge/github.com/b4moss/bwsf.svg)](https://pkg.go.dev/github.com/b4moss/bwsf)
[![Release](https://img.shields.io/github/v/release/b4moss/bwsf)](https://github.com/b4moss/bwsf/releases)
[![License](https://img.shields.io/github/license/b4moss/bwsf)](https://github.com/b4moss/bwsf/blob/main/LICENSE)
[![OpenSSF Scorecard](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fapi.scorecard.dev%2Fprojects%2Fgithub.com%2Fb4moss%2Fbwsf&query=%24.score&label=OpenSSF%20Scorecard&suffix=%2F10)](https://scorecard.dev/viewer/?uri=github.com/b4moss/bwsf)

bwsf (Bitwarden Secured Files) is a CLI tool that uses [Bitwarden](https://bitwarden.com/) to manage `.env*` files and Terraform `*.tfvars` / `*.tfvars.json`.

[Official site](https://bwsf.oss.b4m.jp/)

## 🚨🚨 Important notice 🚨🚨

**v0.20.0 breaking changes:** global config moves to `~/.config/bwsf/config.jsonc` (multi-host schema). Legacy flat `config.json` is migrated on first run (confirm or `--yes`). `not_save_files` is removed (use `save_files` with `!` prefixes). The `bw` CLI backend and `bwsf backend` are removed (API only).

Login to Bitwarden may fail on v0.17.0 and v0.17.1. If you are on the v0.17 line, please upgrade to v0.17.2 or later first.

## Overview

bwsf manages project files in Bitwarden: `.env*`, `*.tfvars`, and `*.tfvars.json` (names containing `.example` are excluded).

Simple usage below:

| command | |
|----|----|
| bwsf setup | Configure hosts and global `save_files` (API only) |
| bwsf auth | Store Personal API Key and authenticate |
| bwsf config show | Show current local configuration |
| bwsf push | Push managed files to your Bitwarden host |
| bwsf pull | Pull managed files from your Bitwarden host |
| bwsf list | List stored projects at your Bitwarden host |
| bwsf clean | Remove local managed files after verifying Bitwarden backup |

## Motivation

We use Bitwarden as our password manager long time ago.
Also, we store .env files our Bitwarden host, manage them as shell scripts.
This project migrates our hand-maded shell scripts to modern CLI command with Go.

## Requirements

No Bitwarden CLI (`bw`) install is required. bwsf talks to Bitwarden over the **API**.

You need:

- A Bitwarden account (Cloud or self-hosted / Vaultwarden)
- A **Personal API Key** (Account Settings → Security → Keys)
- OS secret store access (**macOS Keychain** / **Linux secret service**) for storing the API key

### Homebrew

Homebrew is used to install bwsf itself.

### Machine OS

- macOS
- Linux
- [Is planning] Windows

## Installation

| OS | command |
|----|----|
| macOS | brew tap b4m-oss/tap && brew install bwsf |
| Linux | brew tap b4m-oss/tap && brew install bwsf |

> Note: Linux requires [Homebrew on Linux](https://docs.brew.sh/Homebrew-on-Linux) to be installed first.

Install a past release from the tap. Versioned formulas cover every patch of the
current minor plus the latest patch of the previous minor (older formulas are
pruned). Other releases remain on [GitHub Releases](https://github.com/b4moss/bwsf/releases).

```shell
brew install bwsf@0.18.0
```

## Verify installation

```shell
bwsf -v
# bwsf version 0.20.0
```

## Usage

### Initial setup

```shell
bwsf setup
```

Configure Bitwarden hosts (skippable) and optional global `save_files`. Settings are saved to `~/.config/bwsf/config.jsonc`.

By default, notes are stored in a Bitwarden folder named `dotenvs` (`target_section` on each host). To use a different folder name:

```shell
bwsf setup --folder my-envs
```

Changing the folder name does **not** move existing notes; move them manually in Bitwarden if needed.

You can pass `--host <id>` on `push` / `pull` / `list` / `clean` when multiple hosts are configured.

Check saved values with:

```shell
bwsf config show
```

### Pull managed files from Bitwarden host

```shell
cd /path/to/your_project
bwsf pull
```

bwsf searches Bitwarden for a Note matching the current directory name.
If it exists, managed files are written to the current directory.
If a target file already exists locally, bwsf asks whether to overwrite it (per file).

### Push managed files to Bitwarden host

```shell
cd /path/to/your_project
bwsf push
```

bwsf pushes managed files from the current directory to your Bitwarden host.
If a Note with the same project name already exists in the configured folder (default: `dotenvs`), bwsf **updates it without an overwrite prompt**.

### List projects in Bitwarden host

```shell
bwsf list
```

Prints project names from Bitwarden, one per line on stdout.

### Clean local managed files

```shell
cd /path/to/your_project
bwsf clean
```

Removes local managed files after verifying the remote Bitwarden backup.

### Typical flow

```shell
bwsf setup                 # hosts + optional save_files (+ optional folder)
bwsf auth                  # store Personal API Key; obtain Identity token
bwsf push                  # prompts master password to unlock, then syncs
bwsf pull
bwsf list
```

`bwsf auth` prompts for `client_id` / `client_secret`, stores them in the OS secret store (**macOS Keychain** / **Linux secret service**), and obtains an Identity access token (kept in memory for the process). Use `bwsf auth --clear` to remove the stored key.

Create a Personal API Key in the Bitwarden web vault under Account Settings → Security → Keys.

On each `push` / `pull` / `list`, bwsf prompts for your **master password** to unlock vault keys in memory, then discards keys and tokens when the command exits.

Caveat: unlock uses the Community SDK password-login path with config email + master password for key material. Identity Personal API Key tokens remain separate.

### Upgrading from v0.19

On first run, legacy `~/.config/bwsf/config.json` (flat schema) is detected and migrated after confirmation (or with `--yes`). A backup is written next to the original file. Project configs that still use `not_save_files` must be rewritten to `save_files` with `!` prefixes (load error otherwise).

## Uninstall

```shell
brew uninstall bwsf
```

## FAQ

<details>
<summary>Q. I don't have Bitwarden account.</summary>

To use bwsf, you need a Bitwarden account.

You can access to [Bitwarden Cloud](https://bitwarden.com/), sign up a account.

No fee, No credit card.

</details>

<details>
<summary>Q. I'm Bitwarden self hosted user.</summary>

Ofcourse, bwsf is available for Bitwarden self hosted users.

You can input your self hosted URL when initial setup.

</details>

<details>
<summary>Q. How does my .env file store at Bitwarden host?</summary>

Your managed files are converted to JSON syntax. bwsf creates a Bitwarden Note item and puts that JSON in the Note section.

</details>

<details>
<summary>Q. Where are my Bitwarden account info</summary>

bwsf stores your config data at `~/.config/bwsf/` (formal file: `config.jsonc`).

But, secure information (ex. password) is never stored.

</details>

## Development

### Requirement

**Docker** is needed to be installed your development machine.

### Start up to dev

```
git clone https://github.com/b4moss/bwsf.git
cd bwsf
make run
```

## License

[MIT License](./LICENSE)
