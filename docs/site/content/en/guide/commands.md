# Commands

## Overview

| Command | Description |
|---|---|
| `bwsf setup` | Configure Bitwarden connection |
| `bwsf config show` | Show current local configuration |
| `bwsf push` | Push managed files (`.env*`, `*.tfvars`, `*.tfvars.json`) to Bitwarden |
| `bwsf pull` | Pull managed files from Bitwarden |
| `bwsf list` | List all stored projects |
| `bwsf clean` | Remove local managed files after verifying Bitwarden backup |

Managed files are directory entries whose names start with `.env`, or end with `.tfvars` / `.tfvars.json`. Names that contain `.example` are excluded.

## bwsf setup

Configure your Bitwarden connection settings.

```bash
bwsf setup
```

Optional: use a custom Bitwarden folder instead of `dotenvs`:

```bash
bwsf setup --folder my-envs
```

The folder name is stored in `~/.config/bwsf/config.json` and used by push / pull / list / clean. Renaming does **not** migrate existing notes.

### Interactive prompts

- **Server URL**: Your Bitwarden server URL (leave blank for Bitwarden Cloud)
- **Email**: Your Bitwarden account email
- **Master Password**: Your Bitwarden master password

### Non-interactive flags

For automation (for example smoke tests):

| Flag | Description |
|---|---|
| `--host-type` | `cloud` or `selfhosted` |
| `--url` | Self-hosted server URL (required when `--host-type=selfhosted`) |
| `--email` | Account email |
| `--password` | Master password |
| `--yes` | Assume yes for confirmations (for example create folder) |
| `--folder` | Bitwarden folder name (default: `dotenvs`) |

## bwsf config show

Show the current local configuration from `~/.config/bwsf/config.json`.

```bash
bwsf config show
```

Prints host type, self-hosted URL, email, and the effective folder name. Does not call Bitwarden. If no config file exists, the command exits with an error and suggests running `bwsf setup`.

## bwsf push

Push managed files from the current directory to your Bitwarden vault.

```bash
cd /path/to/your_project
bwsf push
```

### Options

| Option | Description |
|---|---|
| `--from <dir>` | Directory containing managed files (default: current directory) |

### Behavior

1. Uses the current directory name as the project name
2. Detects managed files (`.env*`, `*.tfvars`, `*.tfvars.json`; excludes names containing `.example`)
3. If a Note with the same project name already exists, **updates it without an overwrite prompt**
4. Stores the files as a Note item in the configured folder (default: `dotenvs`)

### Example

```bash
# Push from current directory
cd my-web-app
bwsf push

# Push from a specific directory
bwsf push --from ./config
```

## bwsf pull

Pull managed files from your Bitwarden vault to the current directory.

```bash
cd /path/to/your_project
bwsf pull
```

### Options

| Option | Description |
|---|---|
| `--output <dir>` | Output directory (default: current directory) |

### Behavior

1. Uses the current directory name as the project name
2. Searches for a matching project in the configured folder (default: `dotenvs`)
3. If a target file already exists locally, prompts to overwrite (per file)
4. Writes the managed files from the Note

### Example

```bash
# Pull to current directory
cd my-web-app
bwsf pull

# Pull to a specific directory
bwsf pull --output ./config
```

## bwsf list

List all projects stored in your Bitwarden vault.

```bash
bwsf list
```

### Output

Prints one project name per line to stdout. When the configured folder has no items:

```
No items found in dotenvs folder
```

Example when items exist:

```
my-web-app
api-server
mobile-app
```

## bwsf clean

Remove local managed files after verifying the Bitwarden backup.

```bash
cd /path/to/your_project
bwsf clean
```

### Options

| Option | Description |
|---|---|
| `--from <dir>` | Directory containing managed files to clean (default: current directory) |

### Behavior

1. Uses the current directory name as the project name
2. Aborts if the remote note item is missing or contains no managed files
3. Deletes local files without prompting when contents match
4. On any file mismatch, prompts with a single-select action (Abort / Overwrite remote then clean / Remove local)

## Common Workflows

### Setting up a new project

```bash
# Create managed files
echo "API_KEY=secret123" > .env
echo 'region = "ap-northeast-1"' > terraform.tfvars

# Push to Bitwarden
bwsf push
```

### Syncing on a new machine

```bash
# Clone your project
git clone https://github.com/yourorg/my-web-app.git
cd my-web-app

# Pull managed files from Bitwarden
bwsf pull
```

### Multi-environment setup

```bash
# Create multiple environment files
echo "API_URL=http://localhost:3000" > .env
echo "API_URL=https://staging.example.com" > .env.staging
echo "API_URL=https://api.example.com" > .env.production

# Push all managed files
bwsf push

# On another machine, pull all files
bwsf pull
```
