# ShadowAPI Kubernetes Deployment

This directory contains Kubernetes manifests for deploying ShadowAPI using GitOps patterns with ArgoCD and Kustomize.

## Architecture

The deployment follows a GitOps pattern with:
- **Base configurations** in `k8s/base/` - Common manifests for all environments
- **Environment overlays** in `k8s/overlays/` - Environment-specific configurations
- **ArgoCD applications** in `k8s/argocd/` - GitOps deployment definitions

## Services Deployed

- **Frontend**: React application (port 3000 → 80)
- **Backend**: Go API server (port 8080 → 80)
- **Kratos**: Ory identity management (ports 4433, 4434)
- **NATS**: Message queue (ports 4222, 8222)
- **Migration jobs**: Database and Kratos schema migrations

## External Dependencies

- **PostgreSQL**: External database (not deployed in cluster)
- **Cert-Manager**: For TLS certificates
- **NGINX Ingress Controller**: For traffic routing

## Secrets Configuration

Before deployment, update the following secret placeholders in `k8s/base/secrets.yaml`:

### PostgreSQL Connection
```yaml
POSTGRES_USER: "your-postgres-user"
POSTGRES_PASSWORD: "your-postgres-password"  
POSTGRES_HOST: "your-postgres-host"
POSTGRES_PORT: "5432"
POSTGRES_DB: "shadowapi"
POSTGRES_DB_KRATOS: "kratos"
```

### Application Secrets
```yaml
SECRETS_DEFAULT: "your-kratos-secret-key-32-chars"
TG_APP_ID: "your-telegram-app-id"
TG_APP_HASH: "your-telegram-app-hash"
```

## Deployment

### Using ArgoCD (Recommended)

1. Update the repository URL in `k8s/argocd/shadowapi-app.yaml`
2. Apply ArgoCD application:
```bash
kubectl apply -f k8s/argocd/shadowapi-app.yaml
```

### Manual Deployment

For production:
```bash
kubectl apply -k k8s/overlays/production
```

For staging:
```bash
kubectl apply -k k8s/overlays/staging
```

## Environments

### Production
- **Domain**: shadowapi.devinlab.com
- **Namespace**: shadowapi
- **Replicas**: Frontend (3), Backend (3), Kratos (2)
- **Resources**: Higher limits for production workloads

### Staging  
- **Domain**: staging.shadowapi.devinlab.com
- **Namespace**: shadowapi-staging
- **Replicas**: Default (lower than production)
- **Resources**: Standard limits for testing

## Ingress Configuration

The ingress routes traffic as follows:
- `/api/*` → Backend service
- `/assets/*` → Backend service (static assets)
- `/auth/user/*` → Kratos public API
- `/auth/admin/*` → Kratos admin API
- `/*` → Frontend service (catch-all)

## Migration Jobs

Two migration jobs run before the main services:
1. **kratos-migrate**: Sets up Kratos database schema
2. **db-migrate**: Applies Atlas database migrations

These are Kubernetes Jobs that complete before the main deployments start.

## Monitoring

Each service includes health checks:
- **Liveness probes**: Restart unhealthy containers
- **Readiness probes**: Route traffic only to ready containers

## Customization

To customize for your environment:

1. **Fork this repository**
2. **Update secrets** in `k8s/base/secrets.yaml`
3. **Modify domains** in ingress and configs
4. **Adjust resource limits** in overlay patches
5. **Update ArgoCD app** repository URL
6. **Configure your PostgreSQL** connection details

## File Structure

```
k8s/
├── base/                    # Base Kubernetes manifests
│   ├── backend.yaml        # Backend deployment & service
│   ├── configmap.yaml      # Application configuration
│   ├── frontend.yaml       # Frontend deployment & service
│   ├── ingress.yaml        # Ingress routing rules
│   ├── kratos.yaml         # Kratos deployment & service
│   ├── kustomization.yaml  # Base kustomization
│   ├── migrations.yaml     # Migration jobs
│   ├── namespace.yaml      # Namespace definition
│   ├── nats.yaml          # NATS deployment & service
│   └── secrets.yaml       # Secret templates
├── overlays/
│   ├── production/        # Production environment
│   └── staging/          # Staging environment
└── argocd/
    └── shadowapi-app.yaml # ArgoCD applications
```