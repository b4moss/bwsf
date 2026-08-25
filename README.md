# bwsf

[![CI](https://github.com/b4moss/bwsf/actions/workflows/test.yml/badge.svg)](https://github.com/b4moss/bwsf/actions/workflows/test.yml)
[![Coverage](https://img.shields.io/codecov/c/github/b4moss/bwsf)](https://codecov.io/gh/b4moss/bwsf)
[![Go Reference](https://pkg.go.dev/badge/github.com/b4moss/bwsf.svg)](https://pkg.go.dev/github.com/b4moss/bwsf)
[![Release](https://img.shields.io/github/v/release/b4moss/bwsf)](https://github.com/b4moss/bwsf/releases)
[![License](https://img.shields.io/github/license/b4moss/bwsf)](https://github.com/b4moss/bwsf/blob/main/LICENSE)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/b4moss/bwsf/badge)](https://securityscorecards.dev/viewer/?uri=github.com/b4moss/bwsf)

bwsf (Bitwarden Secured Files) is a CLI tool that uses [Bitwarden](https://bitwarden.com/) to manage `.env*` files and Terraform `*.tfvars` / `*.tfvars.json`.

[Official site](https://bwsf.oss.b4m.jp/)

## 🚨🚨 BREAKING CHANGE 🚨🚨

### Changed CLI Name

From v0.11.0, `bwenv` is re-named as `bwsf`. This is cause some bwenv commands already existed. We decieded to change our CLI name to avoid confusing.

#### MIGRATE

Rename youre setting directory.

```bash
mv ~/.config/bwenv ~/.config/bwsf
```

Uninstall your current version, and re-install latest version.

```bash
brew uninstall bwenv
brew install bwsf
```

### Multiple `.env.enviroment` files

From v0.9.0, bwsf stores multiple enviroment .env files, like `.env | .env.staging | .env.production`.

Cause with this, stored data at Bitwarden Note item structure is changed.

Stored data before v0.8.0 is no compatiblity after v0.9.0.

We will not provide migration system.

## Overview

bwsf manages project files in Bitwarden: `.env*`, `*.tfvars`, and `*.tfvars.json` (names containing `.example` are excluded).

Simple usage below:

| command | |
|----|----|
| bwsf setup | Configure Bitwarden host and account |
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

**`bw`** command is needed to be installed your machine.

[To install bw command, please read this document.](https://bitwarden.com/help/cli/#download-and-install)

**Homebrew**: Need to install bwsf.

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

Install a past release (all published versions are on the tap):

```shell
brew install bwsf@0.15.0
```

## Verify installation

```shell
bwsf -v
# bwsf version 0.16.0
```

## Usage

### Initial setup

```shell
bwsf setup
```

Set up your Bitwarden host and your account information.

By default, notes are stored in a Bitwarden folder named `dotenvs`. To use a different folder name:

```shell
bwsf setup --folder my-envs
```

The folder name is saved in `~/.config/bwsf/config.json` and used by push / pull / list / clean. Changing the folder name does **not** move existing notes; move them manually in Bitwarden if needed.

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

bwsf stores your config data at `~/.config/bwsf/`.

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
