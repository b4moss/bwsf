# Commands

## Overview

| Command | Description |
|---|---|
| `bwsf setup` | Configure Bitwarden connection |
| `bwsf push` | Push .env files to Bitwarden |
| `bwsf pull` | Pull .env files from Bitwarden |
| `bwsf list` | List all stored projects |

## bwsf setup

Configure your Bitwarden connection settings.

```bash
bwsf setup
```

Optional: use a custom Bitwarden folder instead of `dotenvs`:

```bash
bwsf setup --folder my-envs
```

The folder name is stored in `~/.config/bwsf/config.json` and used by push / pull / list. Renaming does **not** migrate existing notes.

This interactive command will prompt you for:
- **Server URL**: Your Bitwarden server URL (leave blank for Bitwarden Cloud)
- **Email**: Your Bitwarden account email
- **Master Password**: Your Bitwarden master password

## bwsf push

Push `.env` files from the current directory to your Bitwarden vault.

```bash
cd /path/to/your_project
bwsf push
```

### Options

| Option | Description |
|---|---|
| `--from <dir>` | Specify source directory (default: current directory) |

### Behavior

1. Uses the current directory name as the project name
2. Searches for `.env*` files in the directory
3. If a project with the same name exists in Bitwarden, prompts to overwrite
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

Pull `.env` files from your Bitwarden vault to the current directory.

```bash
cd /path/to/your_project
bwsf pull
```

### Options

| Option | Description |
|---|---|
| `--output <dir>` | Specify output directory (default: current directory) |

### Behavior

1. Uses the current directory name as the project name
2. Searches for a matching project in the configured folder (default: `dotenvs`)
3. If `.env` files already exist locally, prompts to overwrite
4. Downloads and creates the `.env` files

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

```
Projects in Bitwarden:
  • my-web-app
  • api-server
  • mobile-app
```

## Common Workflows

### Setting up a new project

```bash
# Create your .env file
echo "API_KEY=secret123" > .env

# Push to Bitwarden
bwsf push
```

### Syncing on a new machine

```bash
# Clone your project
git clone https://github.com/yourorg/my-web-app.git
cd my-web-app

# Pull .env from Bitwarden
bwsf pull
```

### Multi-environment setup

```bash
# Create multiple environment files
echo "API_URL=http://localhost:3000" > .env
echo "API_URL=https://staging.example.com" > .env.staging
echo "API_URL=https://api.example.com" > .env.production

# Push all files
bwsf push

# On another machine, pull all files
bwsf pull
```

