---
title: Home
description: Manage .env* and Terraform tfvars with the Bitwarden API
---

# bwsf

Secure File Sync — manage `.env*` and Terraform tfvars via the Bitwarden **API** (Personal API Key).

- [Get Started](/en/guide/getting-started)
- [View on GitHub](https://github.com/b4moss/bwsf)
- [OKF docs hub](https://github.com/b4moss/bwsf/blob/main/docs/index.md) · [pillar](https://github.com/b4moss/bwsf/blob/main/docs/README.md) · [specs](https://github.com/b4moss/bwsf/blob/main/docs/specs/README.md)

## Key Features

- **Secure Storage** — Store managed files (`.env*`, `*.tfvars`, `*.tfvars.json`) securely in your Bitwarden vault.
- **Easy Sync** — Push and pull managed files between your local machine and Bitwarden with simple commands.
- **Multi-Environment** — Manage multiple files (`.env`, `.env.staging`, `terraform.tfvars`, and more) in a single project.
- **Cross-Platform** — Works on macOS and Linux. Windows support is planned.

## Quick Start

```bash
# Install via Homebrew
brew tap b4m-oss/tap && brew install bwsf

# Initial setup + auth
bwsf setup
bwsf auth login

# Pull managed files from Bitwarden
cd /path/to/your_project
bwsf pull

# Push managed files to Bitwarden
bwsf push
```

## How It Works

bwsf talks to Bitwarden over the **API** (no `bw` CLI). Managed files are stored as **Secure Note** items in a Bitwarden folder (`target_section`, default name: `dotenvs`, configurable via setup).

Each project's files are identified by the project name: git root basename (or cwd basename without `.git`), optionally overridden by `.bwsf/config.jsonc` → `override_project_name`.
