# Upgrade

Currently, `bwsf` is in the development stage and is actively updated.

If you installed via brew, you need to update manually:

```bash
brew upgrade bwsf
```

## v0.20.0 breaking changes

- Global config path: `~/.config/bwsf/config.jsonc` (schemaVersion 1, multi-host `settings.hosts`)
- Legacy flat `config.json` is migrated on first load after confirmation (or `--yes`). A `.bak-<timestamp>` backup is written beside the original file
- `not_save_files` removed (global and project). Use `save_files` with `!` prefixes instead
- `backend` field and `bwsf backend` removed — API only (`bw` CLI path abolished)
- Vault commands accept `--host <id>` (resolution: CLI → project `host` → `is_default`)
- Flat `bwsf auth` / `auth --clear` removed — use `bwsf auth login` / `auth logout` (`logout` also clears `vault_unlock`)

## Product specs (OKF)

- OKF index: [`docs/index.md`](https://github.com/b4moss/bwsf/blob/main/docs/index.md)
- Pillar: [`docs/README.md`](https://github.com/b4moss/bwsf/blob/main/docs/README.md)
- Specs index: [`docs/specs/`](https://github.com/b4moss/bwsf/blob/main/docs/specs/README.md)
- Multi-host domains: [`config/global-v2`](https://github.com/b4moss/bwsf/blob/main/docs/specs/config/global-v2.md), [`config/host-resolve`](https://github.com/b4moss/bwsf/blob/main/docs/specs/config/host-resolve.md), [`config/save-files`](https://github.com/b4moss/bwsf/blob/main/docs/specs/config/save-files.md), [`cmd/auth`](https://github.com/b4moss/bwsf/blob/main/docs/specs/cmd/auth.md), [`cmd/unlock-lock`](https://github.com/b4moss/bwsf/blob/main/docs/specs/cmd/unlock-lock.md), [`cmd/init`](https://github.com/b4moss/bwsf/blob/main/docs/specs/cmd/init.md)
