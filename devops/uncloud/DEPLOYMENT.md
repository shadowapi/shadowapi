# MeshPump Uncloud Deployment

This directory contains the configuration for deploying MeshPump to [uncloud.run](https://uncloud.run).

## Architecture

| Service     | Domain            | Description           |
|-------------|-------------------|-----------------------|
| `mp-ssr`      | `meshpump.com`      | Unified frontend (SSR + CSR) |
| `mp-backend`  | `api.meshpump.com`  | Go API server         |
| `mp-grpc2nats` | `rpc.meshpump.com` | gRPC to NATS bridge   |
| `mp-worker-olymp` | (internal, x-node: olymp) | Distributed worker |
| `mp-worker-shkhara` | (internal, x-node: shkhara) | Distributed worker |

The frontend uses a unified architecture where all routes are served by the SSR server:
- Public pages (landing, docs) are server-rendered for SEO
- App pages (`/w/{slug}/*`) hydrate as a client-side SPA

Authentication is handled by an external OIDC provider at `meshpump.oidc.oxoauth.com`.

## Prerequisites

1. **Uncloud CLI** - Install from [uncloud.run](https://uncloud.run)
2. **Domain Configuration** - Point `*.meshpump.com` to your uncloud cluster
3. **External PostgreSQL** - Managed database (Supabase, Neon, AWS RDS, etc.)
4. **External NATS** - NATS JetStream instance
5. **OIDC Tenant** - Configured at `meshpump.oidc.oxoauth.com`

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
- `meshpump` - Main application database
- `meshpump_dev` - Dev database for Atlas migrations

#### NATS
Ensure your NATS instance has JetStream enabled.

#### OIDC
Configure your tenant at `meshpump.oidc.oxoauth.com` with:
- A registered client (SPA, `auth_method=none`, PKCE required)
- Redirect URI: `https://api.meshpump.com/api/v1/auth/oauth2/callback`
- Callback secret for the login provider

### 3. Run Database Migrations

```bash
# From devops/uncloud directory
./migrate.py
```

This script:
- Establishes SSH tunnel to the remote database via `devinlab.com`
- Runs Atlas schema apply interactively (shows diff before applying)
- Automatically cleans up the tunnel on exit

### 4. Deploy to Uncloud

```bash
# From devops/uncloud directory
./deploy.py
```

Or manually:
```bash
uc deploy -f compose.yaml --yes
```

### 5. Verify Deployment

```bash
# Check service health
uc ps

# Check OIDC discovery
curl https://meshpump.oidc.oxoauth.com/.well-known/openid-configuration

# Check API health
curl https://api.meshpump.com/api/v1/health
```

## DNS Configuration

Add the following DNS records pointing to your uncloud cluster:

| Type | Name | Value |
|------|------|-------|
| A | `meshpump.com` | `<uncloud-ip>` |
| A | `api.meshpump.com` | `<uncloud-ip>` |
| A | `rpc.meshpump.com` | `<uncloud-ip>` |

Or use a wildcard:

| Type | Name | Value |
|------|------|-------|
| A | `*.meshpump.com` | `<uncloud-ip>` |
| A | `meshpump.com` | `<uncloud-ip>` |

## Service Discovery

Services communicate internally using uncloud's `.internal` DNS:

- Workers → gRPC bridge: `mp-grpc2nats.internal:9090`
- All services are isolated from external access except through `x-ports`

## Updating the Deployment

```bash
# Pull latest code
git pull

# Run migrations (if schema changed)
cd devops/uncloud
./migrate.py

# Redeploy
./deploy.py
```

## Logs and Debugging

```bash
# View logs for a service
uc logs mp-backend

# Follow logs
uc logs -f mp-backend

# Execute command in container
uc exec mp-backend -- /bin/sh
```

## Rollback

```bash
# List deployment history
uc history

# Rollback to previous version
uc rollback <deployment-id>
```

## Files

| File | Description |
|------|-------------|
| `compose.yaml` | Main uncloud compose file |
| `deploy.py` | Deployment script (validates, deploys, verifies) |
| `migrate.py` | Database migration script (SSH tunnel + Atlas) |
| `.env.example` | Environment variable template |
| `.env.enc` | SOPS-encrypted production secrets |
| `DEPLOYMENT.md` | This documentation |
