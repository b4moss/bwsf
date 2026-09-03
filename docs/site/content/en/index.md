---
title: Home
description: Manage .env* and Terraform tfvars with Bitwarden CLI
---

# bwsf

Secure File Sync — manage `.env*` and Terraform tfvars with Bitwarden CLI.

- [Get Started](/en/guide/getting-started)
- [View on GitHub](https://github.com/b4moss/bwsf)

## Key Features

- **Secure Storage** — Store managed files (`.env*`, `*.tfvars`, `*.tfvars.json`) securely in your Bitwarden vault.
- **Easy Sync** — Push and pull managed files between your local machine and Bitwarden with simple commands.
- **Multi-Environment** — Manage multiple files (`.env`, `.env.staging`, `terraform.tfvars`, and more) in a single project.
- **Cross-Platform** — Works on macOS and Linux. Windows support is planned.

## Quick Start

```bash
# Install via Homebrew
brew tap b4m-oss/tap && brew install bwsf

# Initial setup
bwsf setup

# Pull managed files from Bitwarden
cd /path/to/your_project
bwsf pull

# Push managed files to Bitwarden
bwsf push
```

## How It Works

bwsf uses the official Bitwarden CLI (`bw`) to securely store and retrieve managed files. They are stored as **Note items** in a Bitwarden folder (default name: `dotenvs`, configurable via setup).

Each project's files are identified by the current directory name, making it easy to organize and manage multiple projects.
