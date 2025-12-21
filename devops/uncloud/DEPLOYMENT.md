# ShadowAPI Uncloud Deployment

This directory contains the configuration for deploying ShadowAPI to [uncloud.run](https://uncloud.run).

## Architecture

| Service | Domain | Description |
|---------|--------|-------------|
| `mp-frontend` | `meshpump.com` | React SPA (Vite) |
| `mp-ssr` | `www.meshpump.com` | SSR Express server |
| `mp-backend` | `api.meshpump.com` | Go API server |
| `mp-hydra` | `oidc.meshpump.com` | Ory Hydra OAuth2/OIDC |

## Prerequisites

1. **Uncloud CLI** - Install from [uncloud.run](https://uncloud.run)
2. **Domain Configuration** - Point `*.meshpump.com` to your uncloud cluster
3. **External PostgreSQL** - Managed database (Supabase, Neon, AWS RDS, etc.)
4. **External NATS** - NATS JetStream instance

## Deployment Steps

### 1. Configure Environment

```bash
cd devops/uncloud
cp .env.example .env
# Edit .env with your actual values
```

### 2. Set Up External Services

#### PostgreSQL
Create two databases on your managed PostgreSQL:
- `shadowapi` - Main application database
- `hydra` - Ory Hydra OAuth2 database

Run the schema migration against the `shadowapi` database:
```bash
# From repository root
atlas schema apply \
  --url "YOUR_POSTGRES_URI" \
  --to file://db/schema.sql \
  --auto-approve
```

#### NATS
Ensure your NATS instance has JetStream enabled.

### 3. Deploy to Uncloud

```bash
# From devops/uncloud directory
uncloud deploy -f compose.yaml
```

### 4. Create OAuth2 Client

After Hydra is running, create the SPA OAuth2 client:

```bash
uncloud exec mp-hydra -- hydra create client \
  --endpoint http://localhost:4445 \
  --grant-type authorization_code,refresh_token \
  --response-type code \
  --scope openid,offline_access \
  --redirect-uri https://api.meshpump.com/api/v1/auth/oauth2/callback \
  --name "ShadowAPI SPA" \
  --token-endpoint-auth-method none \
  --format json
```

Copy the `client_id` from the output and update `.env`:
```
BE_OAUTH2_SPA_CLIENT_ID=<client_id>
```

Then redeploy to apply the change:
```bash
uncloud deploy -f compose.yaml
```

### 5. Verify Deployment

```bash
# Check service health
uncloud ps

# Check OIDC discovery
curl https://oidc.meshpump.com/.well-known/openid-configuration

# Check API health
curl https://api.meshpump.com/api/v1/health
```

## DNS Configuration

Add the following DNS records pointing to your uncloud cluster:

| Type | Name | Value |
|------|------|-------|
| A | `meshpump.com` | `<uncloud-ip>` |
| A | `www.meshpump.com` | `<uncloud-ip>` |
| A | `api.meshpump.com` | `<uncloud-ip>` |
| A | `oidc.meshpump.com` | `<uncloud-ip>` |

Or use a wildcard:

| Type | Name | Value |
|------|------|-------|
| A | `*.meshpump.com` | `<uncloud-ip>` |
| A | `meshpump.com` | `<uncloud-ip>` |

## Service Discovery

Services communicate internally using uncloud's `.internal` DNS:

- Backend → Hydra Admin: `http://mp-hydra.internal:4445`
- All services are isolated from external access except through `x-ports`

## Updating the Deployment

```bash
# Pull latest code
git pull

# Redeploy
cd devops/uncloud
uncloud deploy -f compose.yaml
```

## Logs and Debugging

```bash
# View logs for a service
uncloud logs mp-backend

# Follow logs
uncloud logs -f mp-backend

# Execute command in container
uncloud exec mp-backend -- /bin/sh
```

## Rollback

```bash
# List deployment history
uncloud history

# Rollback to previous version
uncloud rollback <deployment-id>
```

## Files

| File | Description |
|------|-------------|
| `compose.yaml` | Main uncloud compose file |
| `hydra.yaml` | Ory Hydra OAuth2/OIDC configuration |
| `.env.example` | Environment variable template |
| `DEPLOYMENT.md` | This documentation |
