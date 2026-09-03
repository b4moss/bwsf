# Dev Container / Codespaces

Issue: [#111](https://github.com/b4moss/bwsf/issues/111)  
Depends on smoke foundation: [#109](https://github.com/b4moss/bwsf/issues/109), [#110](https://github.com/b4moss/bwsf/issues/110)

## What is included

| Tool | Notes |
|------|--------|
| Go 1.26 | Matches `app/go.mod` |
| Node.js + `bw` | `@bitwarden/cli` via `post-create.sh` |
| Docker + Compose | **Docker-outside-of-Docker** (host/Codespaces engine via socket) |
| Make | Used for `make test` / `make smoke*` |

Vaultwarden is **not** started at container create time. Use the existing Compose `smoke` profile when needed (`make smoke-up`).

## Open in Codespaces / VS Code

1. Open the repo in GitHub Codespaces, or VS Code → “Reopen in Container”
2. Wait for `post-create.sh` (installs `bw`, downloads Go modules)
3. Run:

```bash
make test
```

Optional real-server smoke:

```bash
make smoke-up
make smoke-ready
make smoke
```

See [`docs/smoke.md`](../docs/smoke.md).

## Acceptance (v0.13.0)

- Codespaces / Dev Container starts
- `go`, `bw`, and `docker compose` are available
- `make test` passes (required)
- `make smoke` passes once manually (recommended)
