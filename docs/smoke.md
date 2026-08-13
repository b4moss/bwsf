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
| `make smoke` | Vaultwarden | Full command walk (#110; stub until then) |

CI: PR runs unit + mock e2e only. `Smoke Ready` workflow is `workflow_dispatch` / nightly and is **not** required on PRs.

## HTTPS / certificates

Vaultwarden is started with **self-signed TLS** (no `-k` / insecure skip by default).

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

## Local usage

```bash
make build          # once (image includes curl)
make smoke-up       # Vaultwarden on compose profile "smoke"
make smoke-ready    # wait until https://vaultwarden/alive succeeds
make smoke-down
```

Host port mapping (optional browser / host tools): `https://127.0.0.1:8443`  
From containers on the compose network: `https://vaultwarden` (port 80 inside the container, TLS-enabled).

Account creation / `bw login` / running `bwsf` subcommands belong to **#110**, not this foundation.
