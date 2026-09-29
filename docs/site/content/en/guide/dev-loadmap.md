# Features in Development

Marketing-facing checklist. Repository plans / roadmap (OKF): [`docs/roadmap.md`](https://github.com/b4moss/bwsf/blob/main/docs/roadmap.md) · [`docs/plans/`](https://github.com/b4moss/bwsf/blob/main/docs/plans/README.md). Implemented product specs live under [`docs/specs/`](https://github.com/b4moss/bwsf/blob/main/docs/specs/README.md).

- [x] Ability to use folder names other than `dotenvs` (`bwsf setup --folder` / host `target_section`)
- [x] `bwsf clean` command: remove local managed files after verifying Bitwarden backup
- [x] Project root resolution via `.git` (#134)
- [x] `.bwsf/config.(json|jsonc)` project settings (#133 / #177)
  - `override_project_name`, optional `host`, `save_files` (with `!` exclusions; `not_save_files` removed)
- [x] Global multi-host config v2 (#177) — `~/.config/bwsf/config.jsonc`
- [x] Per-host Keychain / unlock·lock (#153)
- [x] `auth login` / `logout` (#174)
- [x] `bwsf init` (#193)

# About Versioning

`bwsf` follows semantic versioning.

Currently, the release of v1.0.0 is undetermined.

Updates to v0.x.0 may always include breaking changes.

Please use with caution in production.
