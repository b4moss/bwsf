# Key Features

`bwsf` is a helper command for securely managing project files such as `.env*` and Terraform `*.tfvars` / `*.tfvars.json` using [Bitwarden](https://bitwarden.com/).

## Saving Managed Files

```bash
# cd /path/to/your/project_root
bwsf push
```

This command saves managed files in your project root to Bitwarden at once. Examples:

- `.env`
- `.env.local`
- `.env.staging`
- `.env.production`
- `terraform.tfvars`
- `prod.auto.tfvars`
- `secret.tfvars.json`

Files whose names contain `.example` (for example `.env.local.example` or `terraform.tfvars.example`) are **not** saved.

## Applying Managed Files to Your Project

```bash
# cd /path/to/your/project_root
bwsf pull
```

This pulls the managed files for the current project stored in Bitwarden and writes them to your project root. Existing local files are overwritten only after confirmation (per file).

## Inspecting Local Configuration

```bash
bwsf config show
```

Shows the values in `~/.config/bwsf/config.json` (host type, URL, email, effective folder name) without calling Bitwarden.

## Cleaning Local Managed Files

```bash
bwsf clean
```

Removes local managed files after verifying that Bitwarden already has a matching backup.

## Multi-User Sharing via Bitwarden

On the Bitwarden side, notes are saved in a configurable folder (default: `dotenvs`).

When you run `bwsf` in your project root, the name of that root folder becomes the project name.

By sharing that folder with other users on Bitwarden, you can share managed files among multiple team members.

For more details, please refer to the [Bitwarden documentation](https://bitwarden.com/resources/).
