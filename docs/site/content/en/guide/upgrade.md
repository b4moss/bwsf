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

Product spec: [v0.20.0-multi-host.md](https://github.com/b4moss/bwsf/blob/main/docs/specs/v0.20.0-multi-host.md)
