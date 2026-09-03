# Commands

## Overview

Current product commands (v0.20.0). A compact inventory also lives in [`docs/COMMANDS.md`](https://github.com/b4moss/bwsf/blob/main/docs/COMMANDS.md).

| Command | Description | Main flags |
|---|---|---|
| `bwsf setup` | Configure hosts and global `save_files` (no login; use `auth`) | `--folder` `--host-type` `--url` `--email` `--skip-host` `--save-files` `--yes` |
| `bwsf init` | Create `./.bwsf/config.jsonc` (requires global config) | `--host` `--skip-host` `--save-files` `--override-project-name` `--yes` |
| `bwsf auth` | Store Personal API Key and authenticate | `--clear` `--host` |
| `bwsf config show` | Show current local configuration | — |
| `bwsf push` | Push managed files (`.env*`, `*.tfvars`, `*.tfvars.json`) to Bitwarden | `--from` `--host` |
| `bwsf pull` | Pull managed files from Bitwarden | `--output` `--host` |
| `bwsf list` | List all stored projects | `--host` |
| `bwsf clean` | Remove local managed files after verifying Bitwarden backup | `--from` `--host` |

Built-in (cobra): `bwsf -v` / `--version`, `bwsf help`, `bwsf completion`.

Managed files are directory entries whose names start with `.env`, or end with `.tfvars` / `.tfvars.json`. Names that contain `.example` are excluded. Optional `save_files` globs (with `!` exclusions) further filter after that.

bwsf uses the **API** only. The `bw` CLI backend and `bwsf backend` are removed in v0.20.0.

Host resolution for vault commands: `--host` → project `host` → global `is_default`.

## bwsf setup

Configure Bitwarden hosts (skippable) and optional global `save_files`.

```bash
bwsf setup
```

Settings are written to `~/.config/bwsf/config.jsonc`. You may skip adding a host (`hosts: []`) and still set `save_files`.

Optional: use a custom Bitwarden folder (`target_section`) instead of `dotenvs`:

```bash
bwsf setup --folder my-envs
```

Renaming does **not** migrate existing notes.

`setup` does **not** log in with a master password. Run `bwsf auth` afterward.

### Non-interactive flags

| Flag | Description |
|---|---|
| `--host-type` | `cloud` or `selfhosted` (maps to `bitwarden-cloud` / `bitwarden-selfhost`) |
| `--url` | Self-hosted server URL (required when `--host-type=selfhosted`) |
| `--email` | Account email |
| `--skip-host` | Leave `hosts: []` |
| `--save-files` | Global `save_files` globs (`!` prefix = exclude) |
| `--yes` | Assume yes for confirmations (folder create, legacy migration) |
| `--folder` | Host `target_section` (default: `dotenvs`) |

## bwsf auth

Store a Personal API Key and obtain an Identity token.

```bash
bwsf auth
bwsf auth --clear
bwsf auth --host work
```

Prompts for `client_id` / `client_secret`, stores them in the OS secret store (**macOS Keychain** / **Linux secret service**), and obtains an Identity access token (kept in memory for the process). Create a key under Account Settings → Security → Keys.

## bwsf config show

Show the current local configuration from `~/.config/bwsf/config.jsonc`.

```bash
bwsf config show
```

Prints schema metadata, `save_files`, and hosts (id / type / url / email / target_section / default). Does not call Bitwarden. If no config file exists, the command exits with an error and suggests running `bwsf setup`.

## bwsf push

Push managed files from the current directory to your Bitwarden vault.

```bash
cd /path/to/your_project
bwsf push
bwsf push --host work
```

If a Note with the same project name already exists in the host `target_section` (default: `dotenvs`), bwsf **updates it without an overwrite prompt**.

## bwsf pull

```bash
cd /path/to/your_project
bwsf pull
bwsf pull --host work
```

## bwsf list

```bash
bwsf list
bwsf list --host work
```

## bwsf clean

```bash
cd /path/to/your_project
bwsf clean
bwsf clean --host work
```

Removes local managed files after verifying the remote Bitwarden backup.
