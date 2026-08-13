# Key Features

`bwsf` is a helper command for securely managing .env files using [Bitwarden](https://bitwarden.com/).

## Saving .env Files

```bash
# cd /path/to/your/project_root
bwsf push
```

This command saves all .env files in your project root to Bitwarden at once.

- .env
- .env.local
- .env.staging
- .env.production

These files are uploaded to Bitwarden together.

Note that configuration files containing `.example` in their name, such as `.env.local.example`, are NOT saved to Bitwarden.

## Applying .env Files to Your Project

```bash
# cd /path/to/your/project_root
bwsf pull
```

This pulls the .env files for the current project stored in Bitwarden and saves them to your project root.

## Multi-User Sharing via Bitwarden

On the Bitwarden side, files are saved in a configurable folder (default: `dotenvs`).

When you run `bwsf` in your project root, the name of that root folder becomes the project name.

By sharing that folder with other users on Bitwarden, you can share .env files among multiple team members.

For more details, please refer to the [Bitwarden documentation](https://bitwarden.com/resources/).

