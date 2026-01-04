# CLAUDE.md

Guide for Claude Code working in the MeshPump (ShadowAPI) repository.

## Quick Start

```bash
make up                    # Bootstrap: secrets, db, migrations, OAuth client, test user
docker compose watch       # Start dev environment with hot reload
```

- **Login:** `admin@example.com` / `Admin123!`
- **App:** `http://app.localtest.me`
- **API:** `http://api.localtest.me/api/v1`

## What Is This

MeshPump is a unified messaging platform (Gmail, Telegram, WhatsApp, LinkedIn) with REST + MCP surface.

| Component | Stack |
|-----------|-------|
| Backend | Go 1.24, ogen, SQLC, NATS JetStream, samber/do DI |
| Frontend | React 19, Vite, Ant Design v6, React Router v7 |
| Auth | Ory Hydra (OAuth2/OIDC), backend handles login/consent |
| Database | PostgreSQL 16, Atlas migrations |

## Code Generation (Critical)

**Never edit generated files manually.** Regenerate after changing sources:

| Change | Command |
|--------|---------|
| `spec/*.yaml` (OpenAPI) | `make api-gen` |
| `backend/proto/*.proto` | `make proto-gen` |
| `db/schema.sql` or `db/tg.sql` | `make sqlc-gen && make sync-db` |

Generated directories (do not edit):
- `backend/pkg/api/` - ogen HTTP server
- `backend/pkg/query/` - SQLC database accessors
- `backend/pkg/proto/` - buf gRPC code
- `front/src/api/v1.d.ts` - TypeScript API types

## Architecture

### Subdomains (Traefik routing)

| Subdomain | Service | Purpose |
|-----------|---------|---------|
| `localtest.me` | SSR (3000) | Public pages (landing, docs) |
| `app.localtest.me` | Frontend (5173) | React SPA |
| `api.localtest.me` | Backend (8080) | REST API |
| `rpc.localtest.me` | Backend (9090) | gRPC for workers |
| `oidc.localtest.me` | Hydra (4444) | OAuth2/OIDC |

### Workspaces

Path-based multi-tenancy: `/w/{slug}/*`

- JWT tokens contain `workspace_id` and `workspace_slug` in `ext` field
- Switching workspaces requires OAuth2 flow (`POST /api/v1/auth/workspace/switch`)
- Roles: `owner`, `admin`, `member`
- Default workspaces: `internal`, `demo`

### RBAC

Casbin-based with workspace-scoped policies. Middleware maps API operations to permissions.

- Key file: `backend/internal/rbac/middleware.go` (OperationPermissionMap)
- Predefined roles: `super_admin`, `workspace_owner`, `workspace_admin`, `workspace_member`

## Repository Layout

```
backend/
  cmd/shadowapi/     # Main server binary
  cmd/worker/        # Distributed worker binary
  internal/          # Domain packages (handler, auth, rbac, worker, grpc, storages)
  pkg/               # Generated code (api, query, proto)
  proto/             # Protobuf definitions
spec/                # OpenAPI definition
db/
  schema.sql         # Main schema
  tg.sql             # Telegram-specific
  sql/               # SQLC query files
front/
  src/app/           # CSR pages (workspace-scoped)
  src/pages/         # SSR pages (public)
  src/lib/           # Shared utilities (auth, workspace, SmartLink)
  src/layouts/       # Layout components
  src/theme.ts       # Ant Design theme config
devops/              # Docker, compose files, Ory config
```

## Adding Features

### New API Endpoint

1. Add to `spec/paths/` and reference in `spec/openapi.yaml`
2. Run `make api-gen`
3. Implement handler in `backend/internal/handler/`
4. Add to RBAC `OperationPermissionMap` if protected

### New Database Table

1. Add to `db/schema.sql`
2. Add queries in `db/sql/*.sql`
3. Run `make sqlc-gen && make sync-db`

### New Frontend Page

**CSR (app pages):**
1. Create in `front/src/app/`
2. Add route to `front/src/app/WorkspaceRouter.tsx`
3. Add menu item to `front/src/layouts/AppLayout.tsx`

**SSR (public pages):**
1. Create in `front/src/pages/`
2. Add route to `front/src/routes.tsx` with `ssr: true`

### Ant Design

Use LLMs.txt for best practices: https://ant.design/llms.txt

## Key Commands

```bash
make help              # All available targets
make up                # Bootstrap everything
make init              # Reset and reinitialize (destructive)
make sync-db           # Apply schema changes
make api-gen           # Regenerate from OpenAPI
go test ./...          # Run Go tests (from backend/)
```

## Guidelines

- **Ask first** before creating predefined objects (datasources, OAuth2 clients, pipelines)
- **Spec first**: Update OpenAPI/SQL schema before implementing
- **DI**: Wire services through `samber/do`, don't instantiate directly
- **Secrets**: Never commit to git; use `.env`, `config.yaml`, or `secrets/`
- Keep diffs tight; don't refactor beyond scope
- Use conventional commits: `feat:`, `fix:`, `docs:`, etc.
