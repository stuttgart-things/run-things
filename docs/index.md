# run-things

**Service portal & health monitor for infrastructure services.**

run-things provides a real-time dashboard with health checks, TLS certificate monitoring, and cluster inventory tracking — backed by Kubernetes CRDs or YAML files.

## Features

- **HTMX dashboard** — dark-themed service portal with live status updates
- **Health checks** — HTTP probes with configurable intervals, expected status codes, and body matching
- **TLS monitoring** — certificate expiry tracking with degraded status warnings
- **Cluster inventory** — Kubernetes workload tracking via gRPC collectors
- **REST API** — JSON endpoints for service CRUD and cluster data
- **Admin panel** — inline editing, tags, add/delete services via web UI
- **Dual storage** — YAML files (disk) or Kubernetes CRDs (ServicePortal)
- **KCL deployment** — type-safe Kubernetes manifests

## Quickstart

```bash
# Run locally with test config
LOAD_CONFIG_FROM=disk CONFIG_LOCATION=tests CONFIG_NAME=services.yaml go run .

# Or via Taskfile
task run-web
```

Open [http://localhost:8080](http://localhost:8080)

## Status Logic

| Status | Condition |
|---|---|
| `UP` | Status code matches, body matches, response < 5s, TLS > 14 days |
| `DEGRADED` | Slow response (> 5s) or TLS cert expiring within 14 days |
| `DOWN` | Request failed or unexpected status code |

## Task Shortcuts

```bash
task run-web                    # run with disk config on :8080
task build                      # lint + test + install
task test                       # go test ./...
task lint                       # golangci-lint
task build-ko                   # build + push container image
```
