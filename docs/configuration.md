# Configuration

All configuration is via environment variables.

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `LOAD_CONFIG_FROM` | `disk` | Config source: `disk` (YAML file) or `cr` (Kubernetes CRD) |
| `CONFIG_LOCATION` | `tests` | Directory path for disk mode, namespace for CRD mode |
| `CONFIG_NAME` | `services.yaml` | YAML filename or CRD resource name |
| `HTTP_PORT` | `8080` | Web UI / REST API port |
| `SERVER_PORT` | `50051` | gRPC server port |
| `KUBECONFIG` | (K8s default) | Kubernetes config path (required for `cr` mode) |

## Service Definition (YAML)

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
      interval: 30
      method: GET
      expectedStatus: 200
      expectedBody: "ok"
      tlsCheck: true
      timeout: 10
      headers:
        Authorization: "Bearer token"
      body: '{"key":"value"}'
```

### Service Fields

| Field | Required | Description |
|---|---|---|
| `name` | yes | Unique service identifier |
| `description` | no | Human-readable description |
| `category` | no | Grouping category (e.g. CI/CD, Monitoring) |
| `url` | yes | Service URL to monitor |
| `logoURL` | no | URL to service logo image |
| `icon` | no | Unicode emoji fallback if no logo |
| `tags` | no | Searchable tags |
| `healthCheck` | no | Health check configuration |

### Health Check Fields

| Field | Default | Description |
|---|---|---|
| `enabled` | `false` | Enable health checking |
| `interval` | `30` | Check interval in seconds |
| `method` | `GET` | HTTP method |
| `expectedStatus` | `200` | Expected HTTP status code |
| `expectedBody` | — | Required text in response body |
| `tlsCheck` | `false` | Check TLS certificate expiry |
| `timeout` | `10` | Request timeout in seconds |
| `headers` | — | Custom HTTP headers (map) |
| `body` | — | Request body for POST/PUT |

## Kubernetes CRD Mode

Set `LOAD_CONFIG_FROM=cr` to read services from a Kubernetes Custom Resource:

```yaml
apiVersion: github.stuttgart-things.com/v1
kind: ServicePortal
metadata:
  name: portal-labul
  namespace: run-things
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

## Example .env

```bash
LOAD_CONFIG_FROM=disk
CONFIG_LOCATION=tests
CONFIG_NAME=services.yaml
SERVER_PORT=50051
HTTP_PORT=8080
```
