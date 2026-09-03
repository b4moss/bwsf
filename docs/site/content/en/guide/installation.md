# Installation

## Requirements

### Operating Systems

| OS | Status |
|---|---|
| macOS | ✅ Supported |
| Linux | ✅ Supported |
| Windows | 🚧 Planned |

### Dependencies

- A Bitwarden account (Cloud or self-hosted / Vaultwarden)
- A **Personal API Key** (Account Settings → Security → Keys)
- OS secret store access (**macOS Keychain** / **Linux secret service**)
- **Homebrew** (to install bwsf)

The Bitwarden CLI (`bw`) is **not required** (removed in v0.20.0).

## Install bwsf

### macOS

```bash
brew tap b4m-oss/tap && brew install bwsf
```

### Linux

::: tip
Linux requires [Homebrew on Linux](https://docs.brew.sh/Homebrew-on-Linux) to be installed first.
:::

```bash
brew tap b4m-oss/tap && brew install bwsf
```

### Install a specific version

The Homebrew tap keeps versioned formulas for every patch of the current minor
and the latest patch of the previous minor. Older releases are pruned from the
tap but remain on [GitHub Releases](https://github.com/b4moss/bwsf/releases).

```bash
brew tap b4m-oss/tap
brew install bwsf@0.18.0
```

Versioned formulas are keg-only, so you may need to adjust your `PATH` or run
`brew link --force bwsf@0.18.0`.

## Verify Installation

```bash
bwsf -v
# bwsf version 0.20.0
```

## Initial Setup

```bash
bwsf setup                 # hosts (skippable) + optional save_files / folder
bwsf auth                  # Personal API Key → OS secret store
```

`setup` does **not** log in with a master password. Authentication is `bwsf auth`. Vault unlock (master password) happens when you run `push` / `pull` / `list`. Settings are saved to `~/.config/bwsf/config.jsonc`.

## Uninstall

```bash
brew uninstall bwsf
```

## Upgrading

```bash
brew upgrade bwsf
```
