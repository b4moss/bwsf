# Getting Started

## What is bwsf?

**bwsf** is a CLI tool that uses [Bitwarden](https://bitwarden.com/) to manage project files such as `.env*` and Terraform `*.tfvars` / `*.tfvars.json` securely.

Instead of sharing secrets through insecure channels like email or Slack, bwsf lets you store them in your Bitwarden vault and sync them across your team.

## Prerequisites

Before using bwsf, make sure you have:

1. **Bitwarden Account** — [Bitwarden Cloud](https://bitwarden.com/) or a self-hosted / Vaultwarden server
2. **Personal API Key** — Account Settings → Security → Keys (default **API** backend)
3. **OS secret store** — macOS Keychain or Linux secret service (stores the API key)

The Bitwarden CLI (`bw`) is **optional**. Install it only if you switch with `bwsf backend --set bw`.

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

## Initial Setup (API backend)

After installing bwsf:

```bash
bwsf setup                 # host type / URL / email (+ optional folder)
bwsf auth                  # store Personal API Key; obtain Identity token
```

`setup` saves host and account email. `auth` stores your Personal API Key in the OS secret store.

Check saved values any time with:

```bash
bwsf config show
```

On each `push` / `pull` / `list`, bwsf prompts for your **master password** to unlock vault keys in memory, then discards keys and tokens when the command exits.

## Next Steps

- [Install bwsf](/en/guide/installation) - Installation instructions for your platform
- [Commands](/en/guide/commands) - Learn all available commands
