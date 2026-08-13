# Getting Started

## What is bwsf?

**bwsf** is a CLI tool that uses [Bitwarden](https://bitwarden.com/) to manage project files such as `.env*` and Terraform `*.tfvars` / `*.tfvars.json` securely.

Instead of sharing secrets through insecure channels like email or Slack, bwsf lets you store them in your Bitwarden vault and sync them across your team.

## Prerequisites

Before using bwsf, make sure you have:

1. **Bitwarden Account** - Either [Bitwarden Cloud](https://bitwarden.com/) or a self-hosted Bitwarden server
2. **Bitwarden CLI (`bw`)** - The official Bitwarden command-line tool

### Installing Bitwarden CLI

Follow the [official Bitwarden CLI installation guide](https://bitwarden.com/help/cli/#download-and-install) to install the `bw` command on your machine.

Verify the installation:

```bash
bw --version
```

## How bwsf Works

bwsf stores managed files as **Note items** in a Bitwarden folder (default name: `dotenvs`). Here's how the structure looks:

```
Bitwarden Vault
└── dotenvs/                    # Default folder for bwsf (configurable)
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

After installing bwsf, run the setup command:

```bash
bwsf setup
```

This will configure:
- Your Bitwarden server URL (for self-hosted instances)
- Your Bitwarden account credentials

Check the saved values any time with:

```bash
bwsf config show
```

## Next Steps

- [Install bwsf](/en/guide/installation) - Installation instructions for your platform
- [Commands](/en/guide/commands) - Learn all available commands
