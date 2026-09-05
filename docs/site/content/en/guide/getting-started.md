# Getting Started

## What is bwsf?

**bwsf** is a CLI tool that uses [Bitwarden](https://bitwarden.com/) to manage project files such as `.env*` and Terraform `*.tfvars` / `*.tfvars.json` securely.

Instead of sharing secrets through insecure channels like email or Slack, bwsf lets you store them in your Bitwarden vault and sync them across your team.

## Prerequisites

Before using bwsf, make sure you have:

1. **Bitwarden Account** — [Bitwarden Cloud](https://bitwarden.com/) or a self-hosted / Vaultwarden server
2. **Personal API Key** — Account Settings → Security → Keys
3. **OS secret store** — macOS Keychain or Linux secret service (stores the API key)

The Bitwarden CLI (`bw`) is **not** required (removed in v0.20.0).

## How bwsf Works

bwsf stores managed files as **Note items** in a Bitwarden folder (`target_section`, default name: `dotenvs`). Here's how the structure looks:

```
Bitwarden Vault
└── dotenvs/                    # Default folder for bwsf (configurable per host)
    ├── my-web-app              # Project name = current directory name
    │   ├── .env
    │   ├── .env.staging
    │   ├── .env.production
    │   └── terraform.tfvars
    └── another-project
        └── .env
```

::: info
By default the folder name is `dotenvs`. You can change it with `bwsf setup --folder <name>`. Changing the name does not move existing notes.
:::

## Initial Setup

After installing bwsf:

```bash
bwsf setup                 # hosts (skippable) + optional save_files / folder
bwsf auth login             # store Personal API Key; unlock vault
```

`setup` writes `~/.config/bwsf/config.jsonc`. `auth` stores your Personal API Key in the OS secret store.

Check saved values any time with:

```bash
bwsf config show
```

On each `push` / `pull` / `list`, bwsf prompts for your **master password** to unlock vault keys in memory, then discards keys and tokens when the command exits.

## Next Steps

- [Install bwsf](/en/guide/installation) - Installation instructions for your platform
- [Commands](/en/guide/commands) - Learn all available commands
- [Upgrade](/en/guide/upgrade) - Breaking changes (v0.20.0 multi-host)
