# Installation

## Requirements

### Operating Systems

| OS | Status |
|---|---|
| macOS | ✅ Supported |
| Linux | ✅ Supported |
| Windows | 🚧 Planned |

### Dependencies

**Bitwarden CLI (`bw`)** is required. Install it first:

```bash
# macOS
brew install bitwarden-cli

# Linux (Snap)
sudo snap install bw

# npm (cross-platform)
npm install -g @bitwarden/cli
```

See the [official Bitwarden CLI docs](https://bitwarden.com/help/cli/#download-and-install) for more options.

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

Past releases are also published on the Homebrew tap:

```bash
brew tap b4m-oss/tap
brew install bwsf@0.15.0
```

See [GitHub Releases](https://github.com/b4moss/bwsf/releases) for available versions. Versioned formulas are keg-only, so you may need to adjust your `PATH` or run `brew link --force bwsf@0.15.0`.

## Verify Installation

```bash
bwsf -v
# bwsf version x.x.x
```

## Initial Setup

After installation, run the setup command to configure your Bitwarden connection:

```bash
bwsf setup
```

You'll be prompted to enter:
1. Your Bitwarden server URL (leave blank for Bitwarden Cloud)
2. Your Bitwarden email
3. Your master password

## Uninstall

```bash
brew uninstall bwsf
```

## Upgrading

```bash
brew upgrade bwsf
```

