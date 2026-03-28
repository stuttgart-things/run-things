# Contributing

## Development Setup

```bash
git clone https://github.com/stuttgart-things/run-things.git
cd run-things
go build ./...
```

Requires **Go 1.26+** and [Task](https://taskfile.dev).

## Available Tasks

| Task | Description |
|---|---|
| `task run-web` | Run web UI with disk config on :8080 |
| `task build` | Lint + test + install |
| `task test` | Run all tests |
| `task lint` | Run golangci-lint |
| `task build-ko` | Build + push + scan container image |

## Running Tests

```bash
task test
# or directly:
go test ./... -v
```

## Project Structure

```
run-things/
├── main.go              # Entry point: env wiring, gRPC + HTTP servers
├── internal/
│   ├── web.go           # HTTP/HTMX handlers, REST API, admin panel
│   ├── models.go        # Service, HealthCheck, ClusterInfo types
│   ├── monitor.go       # Health check scheduler + state management
│   ├── load.go          # Config loading (disk YAML / K8s CRD)
│   ├── save.go          # Config persistence (disk YAML / K8s CRD)
│   ├── k8s.go           # Kubernetes dynamic client
│   ├── tls.go           # TLS certificate checker
│   ├── collector.go     # gRPC cluster collector
│   └── version.go       # Banner + build info
├── kcl/                 # KCL Kubernetes manifests
├── tests/               # Test configs
├── Taskfile.yaml        # Dev tasks
└── .ko.yaml             # Ko build config
```

## Code Conventions

- Logger: `pterm.DefaultLogger.WithLevel(pterm.LogLevelTrace)`
- Env vars: read only in `main.go`, passed as constructor args
- Config persistence: `SaveServices()` handles both disk and CRD backends
- All HTMX handlers redirect to the originating page after mutations

## PR Workflow

1. Create a feature branch: `git checkout -b feat/my-feature`
2. Implement + test
3. `go build ./... && go test ./...`
4. Push and open a PR targeting `main`
5. CI runs build + scan automatically
6. Semantic-release creates a version on merge (use conventional commits)

## Commit Convention

- `feat:` — new feature (minor version bump)
- `fix:` — bug fix (patch version bump)
- `chore:` — maintenance (no release)
- `test:` — test additions (no release)
