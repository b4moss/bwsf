# Smoke environment (Vaultwarden)

Development / CI helpers for **real-server** checks against Vaultwarden.

- Issue: [#109](https://github.com/b4moss/bwsf/issues/109) (foundation), [#110](https://github.com/b4moss/bwsf/issues/110) (`make smoke` commands)
- Parent tracker: [#108](https://github.com/b4moss/bwsf/issues/108)

## Test target split

| Command | Needs Docker server? | Purpose |
|---------|----------------------|---------|
| `make test-unit` | No | Unit / integration |
| `make test-e2e` | No | Mock E2E (`./src/e2e/...`) |
| `make test` | No | All of the above (`go test ./...`; does **not** include smoke) |
| `make smoke-up` / `smoke-down` | Vaultwarden | Start / stop smoke profile |
| `make smoke-ready` | Vaultwarden | HTTPS reachability from the golang container |
| `make smoke` | Vaultwarden | Real command walk (`setup` → `push` → `pull` → `list`) |

CI: PR runs unit + mock e2e only. `Smoke` workflow is `workflow_dispatch` / nightly and is **not** required on PRs.

## HTTPS / certificates

Vaultwarden is started with **self-signed TLS** (no `-k` / insecure skip by default for `bw` / `bwsf`).

| Path | Role |
|------|------|
| `app/smoke/certs/ca.crt` | Smoke CA (mounted into the golang container) |
| `app/smoke/certs/vaultwarden.crt` / `.key` | Server cert (SAN: `vaultwarden`, `localhost`, `127.0.0.1`) |
| `scripts/generate-smoke-certs.sh` | Regenerate the above |

Trust wiring:

- `NODE_EXTRA_CA_CERTS=/smoke-certs/ca.crt` — Bitwarden CLI (`bw`) inside the golang image
- `BWSF_SMOKE_CA` / `curl --cacert` — `make smoke-ready` (`scripts/smoke-ready.sh`)
- Do **not** point `SSL_CERT_FILE` at the smoke CA alone; that would replace the system trust store and break normal HTTPS (e.g. `go mod`)

Regenerate:

```bash
./scripts/generate-smoke-certs.sh
```

Host port mapping (optional browser / host tools): `https://127.0.0.1:8443`  
From containers on the compose network: `https://vaultwarden:80` (ROCKET_TLS listens on 80 inside the container, not 443).

## `make smoke` (#110)

Runs inside the golang container against Vaultwarden:

1. Ensures VW is up / HTTPS ready  
2. Creates `.smoke-tmp/<run-id>/` with isolated `HOME` and project dir (`bwsf-smoke`)  
3. Registers a fixed smoke user (idempotent)  
4. Runs non-interactive `bwsf setup` then `push` / `pull` / `list`  

```bash
make build
make smoke                      # full sequence
make smoke CMD=setup
make smoke CMD=push
make smoke TARGET=vaultwarden BACKEND=bw
```

| Variable | Default | Notes |
|----------|---------|-------|
| `CMD` | `all` | `setup` / `push` / `pull` / `list` / `all` |
| `TARGET` | `vaultwarden` | OSS reserved for later |
| `BACKEND` | `bw` | `api` reserved (errors for now) |
| `BWSF_SMOKE_EMAIL` | `smoke@bwsf.local` | Fixed smoke account |
| `BWSF_SMOKE_PASSWORD` | `SmokePassw0rd!` | Smoke-only weak secret (not production) |
| `BWSF_SMOKE_KEEP_TMP` | `0` | Set `1` to keep tmp on success |

### tmp policy

| Result | Behavior |
|--------|----------|
| Success | Delete `.smoke-tmp/<run-id>/` (unless `BWSF_SMOKE_KEEP_TMP=1`) |
| Failure | Keep the directory and print its path for inspection |

`.smoke-tmp/` is gitignored.

### Non-interactive setup

```bash
bwsf setup \
  --host-type selfhosted \
  --url https://vaultwarden:80 \
  --email smoke@bwsf.local \
  --password 'SmokePassw0rd!' \
  --yes
```

Unlock prompts elsewhere honor `BWSF_PASSWORD` when set (used by the smoke runner).
