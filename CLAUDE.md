# CLAUDE.md

Guide for Claude Code working in the MeshPump (ShadowAPI) repository.

## Quick Start

```bash
make up                    # Bootstrap: secrets, db, migrations, OAuth client, test user
docker compose watch       # Start dev environment with hot reload
```

- **Login:** `admin@example.com` / `Admin123!`
- **App:** `http://localtest.me`
- **API:** `http://api.localtest.me/api/v1`
- **API Docs:** `http://spec.localtest.me`

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
| `localtest.me` | Frontend (3000) | All pages (SSR + CSR unified) |
| `api.localtest.me` | Backend (8080) | REST API |
| `rpc.localtest.me` | Backend (9090) | gRPC for workers |
| `oidc.localtest.me` | Hydra (4444) | OAuth2/OIDC |
| `mail.localtest.me` | Mailpit (8025) | Email testing UI |
| `spec.localtest.me` | Nginx (80) | API documentation (RapiDoc) |

### Workspaces

Path-based multi-tenancy: `/w/{slug}/*`

- JWT tokens contain `workspace_id` and `workspace_slug` in `ext` field
- Switching workspaces requires OAuth2 flow (`POST /api/v1/auth/workspace/switch`)
- Roles: `owner`, `admin`, `member`
- Default workspaces: `internal`, `demo`

### Access Control

Ladon-based policy system with workspace-scoped policies. Middleware maps API operations to permissions.

- API endpoints: `/api/v1/access/*` (policy-set, permission, user policy assignments)
- Key file: `backend/internal/rbac/middleware.go` (OperationPermissionMap)
- Database tables: `policy_set`, `permission`, `user_policy_set`
- Predefined policy sets: `super_admin`, `workspace_owner`, `workspace_admin`, `workspace_member`
- Policy sets are assigned directly to users via `user_policy_set` table (UUID primary key)
- Subjects use format `policy_set:<name>` in Ladon policies
- Frontend: `front/src/app/access/` (PolicySets.tsx, PolicySetEdit.tsx)

### Usage Limits

Configurable limits on message processing with two dimensions:
- **User limits** (per workspace): Policy set defaults + per-user overrides
- **Worker limits** (per workspace): Per-worker limits

Features:
- Reset periods: `daily`, `weekly`, `monthly`, `rolling_24h`, `rolling_7d`, `rolling_30d`
- `NULL` limit value means unlimited
- Effective limit is the minimum of user and worker limits
- Partial processing: processes up to remaining quota

API endpoints:
- `GET/POST /api/v1/access/usage-limits` - Policy set default limits
- `GET/PUT/DELETE /api/v1/access/usage-limits/{uuid}`
- `GET/POST /api/v1/access/user/{user_uuid}/usage-limits` - User overrides
- `PUT/DELETE /api/v1/access/user/{user_uuid}/usage-limits/{uuid}`
- `GET/POST /api/v1/access/worker/{worker_uuid}/usage-limits` - Worker limits
- `PUT/DELETE /api/v1/access/worker/{worker_uuid}/usage-limits/{uuid}`
- `GET /api/v1/access/usage-status` - Combined usage status

Database tables: `usage_limit`, `user_usage_limit_override`, `worker_usage_limit`, `user_usage_tracking`, `worker_usage_tracking`

Key file: `backend/internal/usagelimits/manager.go`

### Email

| Environment | Service | Configuration |
|-------------|---------|---------------|
| Development | Mailpit | `mailpit:1025` (SMTP), `mail.localtest.me` (Web UI) |
| Production | Amazon SES | Configure via `BE_SMTP_*` environment variables |

- **Dev:** All emails are captured by Mailpit and viewable at `http://mail.localtest.me`
- **Prod:** Must use Amazon SES. Never use Mailpit in production.

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
  src/layouts/       # Layout components (AppLayout, AuthLayout, LandingLayout)
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

**CSR (app pages with sidebar):**
1. Create in `front/src/app/`
2. Add route to `front/src/app/WorkspaceRouter.tsx`
3. Add menu item to `front/src/layouts/AppLayout.tsx`

**SSR (public pages without sidebar):**
1. Create in `front/src/pages/`
2. Add route to `front/src/routes.tsx` with `ssr: true`, `layout: 'app'`, `showSidebar: false`

All pages use `AppLayout` - use `showSidebar: false` to hide the sidebar for public/documentation pages.

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

## Secrets Management

Production secrets are encrypted with [SOPS](https://github.com/getsops/sops) + [age](https://github.com/FiloSottile/age).

### Files

| File | Purpose |
|------|---------|
| `.sops.yaml` | SOPS config with team age public keys |
| `devops/uncloud/.env.enc` | Encrypted production secrets (committed) |
| `devops/uncloud/.env` | Decrypted secrets (gitignored) |
| `~/.config/sops/age/keys.txt` | Your private key (never share!) |

### Commands

```bash
make secrets-decrypt   # Decrypt .env.enc → .env
make secrets-encrypt   # Encrypt .env → .env.enc
make secrets-edit      # Edit secrets in-place (auto re-encrypts)
make secrets-rotate    # Re-encrypt after adding/removing team members
```

### Setup (New Team Member)

1. Install tools:
   ```bash
   brew install sops age   # macOS
   ```

2. Generate your age key:
   ```bash
   age-keygen -o ~/.config/sops/age/keys.txt
   ```

3. Share your **public key** (starts with `age1...`) with the team

4. An existing member adds your key to `.sops.yaml` and runs:
   ```bash
   make secrets-rotate
   ```

5. Pull the updated `.sops.yaml` and `.env.enc`, then:
   ```bash
   make secrets-decrypt
   ```

### Adding a Team Member

1. Get their age public key (starts with `age1...`)

2. Edit `.sops.yaml`, append their key (comma-separated):
   ```yaml
   age: >-
     age1existing...,
     age1newmember...
   ```

3. Re-encrypt secrets:
   ```bash
   make secrets-rotate
   ```

4. Commit both `.sops.yaml` and `devops/uncloud/.env.enc`

### Removing a Team Member

1. Remove their key from `.sops.yaml`
2. Run `make secrets-rotate`
3. **Rotate all secrets** (they had access to plaintext values)
4. Commit changes

## Production Deployment
Production deployment uses Uncloud platform (https://uncloud.run/docs/).
We use separate compose file for it: `devops/uncloud/compose.yaml`.
The secrets are loaded from decrypted `.env` file in `devops/uncloud/`.
To do full deploy (with migrations), run:
```bash
make uncloud-deploy
```

Check `devops/uncloud/DEPLOYMENT.md` for full instructions.

## Guidelines

- **Ask first** before creating predefined objects (datasources, OAuth2 clients, pipelines)
- **Spec first**: Update OpenAPI/SQL schema before implementing
- **DI**: Wire services through `samber/do`, don't instantiate directly
- **Secrets**: Use SOPS encryption (see above); never commit plaintext secrets
- Keep diffs tight; don't refactor beyond scope
- Use conventional commits: `feat:`, `fix:`, `docs:`, etc.
