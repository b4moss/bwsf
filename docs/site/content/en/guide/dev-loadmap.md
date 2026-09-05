# Features in Development

- [x] Ability to use folder names other than `dotenvs` (`bwsf setup --folder` / host `target_section`)
- [x] `bwsf clean` command: remove local managed files after verifying Bitwarden backup
- [x] Project root resolution via `.git` (#134)
- [x] `.bwsf/config.(json|jsonc)` project settings (#133 / #177)
  - `override_project_name`, optional `host`, `save_files` (with `!` exclusions; `not_save_files` removed)
- [x] Global multi-host config v2 (#177) — `~/.config/bwsf/config.jsonc`
- [ ] Per-host Keychain / unlock·lock (#153)
- [ ] `auth login` / `logout` (#174)
- [ ] `bwsf init` (#193)

# About Versioning

`bwsf` follows semantic versioning.

Currently, the release of v1.0.0 is undetermined.

Updates to v0.x.0 may always include breaking changes.

Please use with caution in production.
