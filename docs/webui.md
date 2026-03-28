# Web UI

The HTMX dashboard is available on port `:8080` (configurable via `HTTP_PORT`).

## Routes

| Path | Description |
|---|---|
| `/` | Dashboard — service cards with live health status |
| `/service/{name}` | Service detail — health history timeline |
| `/clusters` | Cluster inventory overview |
| `/cluster/{name}` | Cluster detail — workloads, deployments, services |
| `/admin` | Admin panel — add, edit, delete services |

## Dashboard

The main dashboard shows all services as cards grouped by category. Each card displays:

- Service name, description, and logo/icon
- Live health status dot (green/orange/red/grey)
- Response time
- Mini health timeline (last 60 checks)
- Tags

Cards auto-refresh health status via HTMX polling.

## Admin Panel

The admin panel (`/admin`) provides:

### Add Service

Fill in the form fields:

- **Name** (required) — unique service identifier
- **URL** (required) — endpoint to monitor
- **Description** — human-readable description
- **Category** — grouping (e.g. CI/CD, Monitoring)
- **Logo URL** — service logo image URL
- **Icon** — emoji fallback if no logo
- **Tags** — comma-separated (e.g. `gitops, kubernetes`)
- **Health Check** — enable HTTP health monitoring
- **TLS Check** — monitor certificate expiry

### Inline Edit

Click **Edit** on any service row to expand an inline edit form with all fields pre-populated. Only one edit row can be open at a time. Click **Save** to persist or **Cancel** to discard.

### Delete

Click **Delete** on any service row (with confirmation prompt).

All changes are persisted immediately to the configured backend (disk YAML or Kubernetes CRD).
