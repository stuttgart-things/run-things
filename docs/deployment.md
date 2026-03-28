# KCL Deployment

run-things ships with [KCL](https://kcl-lang.io/) manifests (`kcl/`) for Kubernetes deployment.

## Rendered Resources

| Resource | Kind | Conditional |
|---|---|---|
| `run-things` | Namespace | Always |
| `run-things` | ServiceAccount | Always |
| `run-things-config` | ConfigMap | Always |
| `run-things` | Role | Always |
| `run-things` | RoleBinding | Always |
| `run-things` | Deployment | Always |
| `run-things` | Service (gRPC) | Always |
| `run-things-http` | Service (HTTP) | Always |
| `run-things` | HTTPRoute | Only if `httpRouteEnabled=true` |

## Deploy

```bash
# Render and apply
cd kcl && kcl run | kubectl apply -f -

# With custom config
kcl run -D 'config.image=ghcr.io/stuttgart-things/run-things:v0.1.0' \
        -D 'config.namespace=run-things' \
  | kubectl apply -f -

# With HTTPRoute
kcl run -D 'config.httpRouteEnabled=true' \
        -D 'config.httpRouteParentRefName=my-gateway' \
        -D 'config.httpRouteHostname=run-things.example.com' \
  | kubectl apply -f -
```

## Configuration Reference

| Parameter | Default | Description |
|---|---|---|
| `config.name` | `run-things` | Resource name |
| `config.namespace` | `run-things` | Target namespace |
| `config.image` | `ghcr.io/stuttgart-things/run-things:v0.1.0` | Container image |
| `config.imagePullPolicy` | `Always` | Image pull policy |
| `config.replicas` | `1` | Replica count |
| `config.grpcPort` | `50051` | gRPC container port |
| `config.httpPort` | `8080` | HTTP container port |
| `config.loadConfigFrom` | `cr` | Config source: `disk` or `cr` |
| `config.configLocation` | `run-things` | K8s namespace or file path |
| `config.configName` | `portal-labul` | CRD resource name or filename |
| `config.httpRouteEnabled` | `false` | Enable HTTPRoute (Gateway API) |
| `config.httpRouteParentRefName` | — | Gateway name |
| `config.httpRouteHostname` | — | Hostname for HTTPRoute |

## Container Image

Built with [ko](https://ko.build/) using a distroless base:

```bash
task build-ko
```

Image: `ghcr.io/stuttgart-things/run-things`

## OCI Kustomize Artifact

The release pipeline publishes a Kustomize OCI artifact to `ghcr.io/stuttgart-things/run-things-kustomize`.
