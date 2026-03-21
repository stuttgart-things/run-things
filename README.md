# weRunThings

Service portal & health monitor for infrastructure services. Provides a real-time dashboard with health checks, TLS certificate monitoring, and cluster inventory tracking.

<img src="https://raw.githubusercontent.com/stuttgart-things/docs/main/hugo/sthings-cinema%20wide.png" alt="weRunThings" width="200">

## Features

- HTMX-powered dashboard with dark theme
- HTTP health checks with configurable intervals, expected status codes, and body matching
- TLS certificate expiry monitoring
- Kubernetes cluster inventory via gRPC collectors
- REST API for service management
- Dual config source: YAML files or Kubernetes CRDs
- Admin panel for adding/editing/deleting services

## Quick Start

```bash
# run locally with default test config
task run-web

# or directly
LOAD_CONFIG_FROM=disk CONFIG_LOCATION=tests CONFIG_NAME=services.yaml go run .
```

Open [http://localhost:8080](http://localhost:8080)

## Configuration

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `LOAD_CONFIG_FROM` | `disk` | Config source: `disk` (YAML file) or `cr` (Kubernetes CRD) |
| `CONFIG_LOCATION` | `tests` | Directory path for disk mode, namespace for CRD mode |
| `CONFIG_NAME` | `services.yaml` | YAML filename or CRD resource name |
| `HTTP_PORT` | `8080` | Web UI / REST API port |
| `SERVER_PORT` | `50051` | gRPC server port |
| `KUBECONFIG` | (K8s default) | Kubernetes config path (required for `cr` mode) |

### Service Definition (YAML)

```yaml
services:
  - name: ArgoCD
    description: Declarative GitOps continuous delivery tool
    category: CI/CD
    url: https://argocd.example.com
    logoURL: https://argo-cd.readthedocs.io/en/stable/assets/logo.png
    tags:
      - gitops
      - kubernetes
    healthCheck:
      enabled: true
      interval: 30           # check interval in seconds (default: 30)
      method: GET             # HTTP method: GET, POST, HEAD (default: GET)
      expectedStatus: 200     # expected HTTP status code (default: 200)
      expectedBody: "ok"      # text that must appear in response body
      tlsCheck: true          # monitor TLS certificate expiry
      timeout: 10             # request timeout in seconds (default: 10)
      headers:                # custom HTTP headers
        Authorization: "Bearer token"
      body: '{"key":"value"}' # request body for POST/PUT
```

### Service Fields

| Field | Required | Description |
|---|---|---|
| `name` | yes | Unique service identifier |
| `description` | no | Human-readable description |
| `category` | no | Grouping category (e.g. CI/CD, Monitoring, Security) |
| `url` | yes | Service URL to monitor |
| `logoURL` | no | URL to service logo image |
| `icon` | no | Unicode emoji fallback if no logo |
| `tags` | no | Searchable tags |
| `healthCheck` | no | Health check configuration (see below) |

### Health Check Fields

| Field | Default | Description |
|---|---|---|
| `enabled` | `false` | Enable health checking |
| `interval` | `30` | Check interval in seconds |
| `method` | `GET` | HTTP method |
| `expectedStatus` | `200` | Expected HTTP status code |
| `expectedBody` | - | Required text in response body |
| `tlsCheck` | `false` | Check TLS certificate expiry |
| `timeout` | `10` | Request timeout in seconds |
| `headers` | - | Custom HTTP headers (map) |
| `body` | - | Request body for POST/PUT |

### Status Logic

| Status | Condition |
|---|---|
| `UP` | Status code matches, body matches, response < 5s, TLS > 14 days |
| `DEGRADED` | Slow response (> 5s) or TLS cert expiring within 14 days |
| `DOWN` | Request failed or unexpected status code |

### Kubernetes CRD Mode

Set `LOAD_CONFIG_FROM=cr` to read services from a Kubernetes Custom Resource:

```yaml
apiVersion: github.stuttgart-things.com/v1
kind: ServicePortal
metadata:
  name: portal-labul        # matches CONFIG_NAME
  namespace: werunthings    # matches CONFIG_LOCATION
spec:
  services:
    - name: ArgoCD
      description: GitOps CD
      category: CI/CD
      url: https://argocd.example.com
      healthCheck:
        enabled: true
        interval: 30
        expectedStatus: 200
        tlsCheck: true
```

## API

### REST Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/services` | List all services |
| `GET` | `/api/v1/services/{name}` | Get service details |
| `POST` | `/api/v1/services` | Add a service |
| `DELETE` | `/api/v1/services/{name}` | Delete a service |
| `GET` | `/api/v1/clusters` | List cluster inventory |
| `GET` | `/api/v1/health` | Health probe |

### Web UI Routes

| Path | Description |
|---|---|
| `/` | Dashboard |
| `/service/{name}` | Service detail page |
| `/clusters` | Cluster inventory |
| `/admin` | Admin panel |

## Build

```bash
# build + install with version info
task build

# build container image with ko
task build-ko

# run tests
task test

# lint
task lint
```

### Build-Time Variables

Version, commit, and build date are injected via ldflags:

```bash
go install -ldflags="-X github.com/stuttgart-things/weRunThings/internal.version=v1.0.0 \
  -X github.com/stuttgart-things/weRunThings/internal.date=$(date -Ih) \
  -X github.com/stuttgart-things/weRunThings/internal.commit=$(git log -n1 --format=%h)"
```

## License

Copyright 2026 Patrick Hermann patrick.hermann@sva.de

a [stuttgart-things](https://github.com/stuttgart-things) project
