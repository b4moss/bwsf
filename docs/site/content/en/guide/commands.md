# Commands

## Overview

Current product commands (v0.20.0). A compact inventory also lives in [`docs/COMMANDS.md`](https://github.com/b4moss/bwsf/blob/main/docs/COMMANDS.md).

| Command | Description | Main flags |
|---|---|---|
| `bwsf setup` | Configure hosts and global `save_files` (no login; use `auth login`) | `--folder` `--host-type` `--url` `--email` `--skip-host` `--save-files` `--yes` |
| `bwsf auth login` | Store API Key, verify Identity, unlock vault | `--host` |
| `bwsf auth logout` | Remove API Key and `vault_unlock` | `--host` `--all` |
| `bwsf unlock` | Unlock vault session and persist `vault_unlock` | `--host` |
| `bwsf lock` | Clear `vault_unlock` (keeps API Key) | `--host` `--all` |
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

`setup` does **not** log in with a master password. Run `bwsf auth login` afterward.

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

## bwsf init

Create `./.bwsf/config.jsonc` in the **current directory** (does not walk up to a git root).

```bash
bwsf init
bwsf init --skip-host --yes
bwsf init --host default --save-files '.env*,!.env.local' --override-project-name my-api
```

Requires an existing global config (`bwsf setup`; `hosts: []` is OK). Optional project `host` is an id reference into global `hosts[]`. Existing project config prompts for overwrite unless `--yes`.

### Non-interactive flags

| Flag | Description |
|---|---|
| `--host <id>` | Write project `host` (must exist in global config) |
| `--skip-host` | Omit project `host` |
| `--save-files` | Project `save_files` globs (`!` = exclude) |
| `--override-project-name` | Project name override (empty omits the key) |
| `--yes` | Skip overwrite confirmation |


## bwsf auth

Manage Personal API Key authentication. Bare `bwsf auth` prints help only.

```bash
bwsf auth login
bwsf auth login --host work
bwsf auth logout
bwsf auth logout --all
```

`auth login` prompts for `client_id` / `client_secret` (or reuses a stored key), verifies Identity, then unlocks the vault and persists `vault_unlock` (same outcome as `bwsf unlock`). Create a key under Account Settings → Security → Keys.

`auth logout` removes the API Key **and** `vault_unlock` for the resolved host. `bwsf lock` clears only the vault session.

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
