# REST API

## Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/services` | List all services |
| `GET` | `/api/v1/services/{name}` | Get service details + health history |
| `POST` | `/api/v1/services` | Add a service |
| `DELETE` | `/api/v1/services/{name}` | Delete a service |
| `GET` | `/api/v1/clusters` | List cluster inventory |
| `GET` | `/api/v1/health` | Health probe |

## Examples

### List all services

```bash
curl http://localhost:8080/api/v1/services
```

### Add a service

```bash
curl -X POST http://localhost:8080/api/v1/services \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ArgoCD",
    "description": "GitOps CD tool",
    "category": "CI/CD",
    "url": "https://argocd.example.com",
    "tags": ["gitops", "kubernetes"],
    "healthCheck": {
      "enabled": true,
      "interval": 30,
      "expectedStatus": 200,
      "tlsCheck": true
    }
  }'
```

### Delete a service

```bash
curl -X DELETE http://localhost:8080/api/v1/services/ArgoCD
```

### Get service details

```bash
curl http://localhost:8080/api/v1/services/ArgoCD
```

Returns the service definition, current status, and health check history (last 60 results).

### Health probe

```bash
curl http://localhost:8080/api/v1/health
# {"status":"ok"}
```
