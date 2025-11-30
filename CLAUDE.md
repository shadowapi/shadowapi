# CLAUDE.md

This document guides Claude Code (claude.ai/code) when working inside the ShadowAPI repository. Follow it before you change or create files.

## Snapshot

- ShadowAPI is a unified messaging platform that normalises Gmail, Telegram, WhatsApp, and LinkedIn into a single REST + MCP surface.
- Backend: Go 1.24 service using Cobra CLI, samber/do dependency injection, ogen-generated handlers, SQLC + PostgreSQL 16, NATS JetStream, S3/host/local storage abstractions.
- Infrastructure: Docker Compose stack with Traefik, backend, Postgres, and NATS; orchestrated by Make (`Makefile`).

## Repository map

- `backend/` – Go service. `cmd/shadowapi` hosts the CLI (`serve`, `loader`, `reset-password`), `internal/` holds domain packages (auth, handler, worker, storages, queue, metrics, config), `pkg/api` is ogen-generated HTTP server (do not edit manually), `pkg/query` is SQLC output (do not edit), and `sdk-go/` contains a generated client example.
- `spec/` – OpenAPI definition (`openapi.yaml`) plus ogen configuration and shared pieces under `components/` and `paths/`.
- `db/` – Canonical SQL schema in `schema.sql` with Telegram-specific relations in `tg.sql`. Atlas migrations are applied from these files; no separate migration files exist.
- `devops/` – Dockerfiles, Atlas configs, sqlc builder image, helper scripts, and Ory configuration files used by Compose/Make.
- `devops/ory/` – Ory Kratos and Hydra configuration files (`kratos/kratos.yaml`, `kratos/identity.schema.json`, `hydra/hydra.yaml`).
- `devops/compose/auth/` – Docker Compose for Ory authentication services (Kratos, Hydra, hydra-consent, mailhog).
- `docs/` – Product documentation and screenshots referenced by the README.
- `templates/`, `start/`, `k8s/`, `secrets/` – Supporting assets (UI templates, bootstrap scripts, Kubernetes manifests, local keys). Leave anything under `secrets/` untouched.
- `Makefile` – Source of all make targets; prefer calling `make <target>` over shelling out to `docker compose` directly.

## Backend

### Stack & runtime

- Go 1.24 with modules pinned in `backend/go.mod`.
- Dependency injection via `github.com/samber/do/v2`; resolve services through the container rather than new-ing them manually.
- HTTP surface is generated from the OpenAPI spec by `ogen`, wired through `backend/internal/server`.
- Long-running jobs and pipelines live under `backend/internal/worker` and communicate through NATS (`backend/internal/queue`).
- Observability uses `log/slog`, OpenTelemetry (`go.opentelemetry.io/otel`), and Prometheus metrics (`backend/internal/metrics`).

### Key packages

- `internal/handler` – Implements ogen handlers for messages, contacts, pipelines, storages, auth callbacks, etc.
- `internal/auth` – Authentication middleware and token validation.
- `internal/auth/dbauth` – Database-based user manager implementation.
- `internal/storages` – Pluggable storage backends (Postgres, S3, host filesystem).
- `internal/worker` – Pipelines, extractors, filters, schedulers, job registry, and cancellation logic.
- `internal/imap`, `internal/whatsapp`, `internal/tg`, `internal/oauth2` – External channel integrations (IMAP/SMTP Gmail, WhatsApp via whatsmeow, Telegram via gotd, LinkedIn helpers).
- `pkg/api` – Generated server; never edit by hand. Regenerate with `make api-gen-backend`.
- `pkg/query` – SQLC-generated database accessors; regenerate with `sqlc generate` in the backend directory.

### Config & secrets

- Default configuration lives in `backend/config.example.yaml`; copy to `backend/config.yaml` for local overrides.
- Environment variables are BE-prefixed (see `backend/internal/config/config.go`). Important ones include `BE_DB_URI`, `BE_QUEUE_URL`, and `TG_APP_ID`/`TG_APP_HASH`.
- Sensitive keys are read from `secrets/`; the repo should never gain new secrets in git history.

### Generation & migrations

- Run `make api-gen` after editing `spec/`. This updates the Go server.
- Run `sqlc generate` (and `sqlc vet`) in the backend directory after updating the SQL schema. SQL lives in `db/schema.sql` + `db/tg.sql`.
- `make sync-db` concatenates schema files and applies them via Atlas to the running Postgres instance.

### Command-line entrypoints

- `go build -o ./bin/shadowapi ./cmd/shadowapi` builds the binary locally.
- `shadowapi serve` is the main API server when running outside Docker.

## Frontend

### Stack

- Vite and React 19 with TypeScript.
- Ant Design component library (v6).
- React Router v7 for client-side routing.
- Use LLMs.txt for Ant Design best practices and guidelines from this URL https://ant.design/llms.txt

### Hybrid SSR Architecture

The frontend uses a hybrid rendering approach where public pages (`/page/*`) are server-side rendered for SEO, while the app (`/`, `/app/*`) uses client-side rendering.

**Two containers, one codebase:**
- **Frontend container** (port 5173): Vite dev server for CSR routes
- **SSR container** (port 3000): Express + Vite middleware for SSR routes
- Both containers use the same `front/` directory

**Routing behavior:**
- Direct URL access to `/page/*` → SSR container renders full HTML, then client hydrates
- SPA navigation within `/page/*` → React Router (no page reload)
- Navigation from `/page/*` to `/` → Full page reload (crosses container boundary)
- Direct URL access to `/` → Frontend container serves index.html, client renders

### Key frontend files

- `front/server.ts` – Express SSR server with Vite middleware, handles `/page/*` routes
- `front/src/entry-client.tsx` – Client entry point; uses `hydrateRoot` for SSR pages, `createRoot` for CSR
- `front/src/entry-server.tsx` – SSR render function with Ant Design CSS-in-JS extraction
- `front/src/routes.tsx` – Centralized route configuration with `ssr` and `protected` flags per route
- `front/src/lib/SmartLink.tsx` – Navigation component that decides between SPA navigation and full reload
- `front/src/lib/ssr-context.tsx` – SSR data provider for passing server-fetched data to client
- `front/src/lib/data-fetching.ts` – Route-based data loaders for SSR
- `front/src/lib/auth/` – Authentication module (Kratos client, context, hooks, protected route)
- `front/src/layouts/` – Layout components (BaseLayout for shared, AppLayout for CSR, PageLayout for SSR, AuthLayout for login)
- `front/src/pages/auth/` – Authentication pages (LoginPage)

### Development scripts

- `npm run dev` – Start Vite dev server (CSR only, used by frontend container)
- `npm run dev:ssr` – Start Express SSR server with Vite middleware (used by SSR container)
- `npm run build` – Build both client and server bundles for production

### Adding new pages

1. Create the page component in `front/src/pages/` or `front/src/app/`
2. Add route to `front/src/routes.tsx` with appropriate flags:
   - `ssr: true` for public/SEO pages under `/page/*`
   - `ssr: false` for app pages under `/` or `/app/*`
   - `protected: false` for public routes that don't require authentication (default is protected for `app` layout)
3. If the page needs server-side data, add a loader in `front/src/lib/data-fetching.ts`

### Frontend Authentication

The frontend uses Ory Kratos for authentication with a custom login form.

**Key files:**
- `front/src/lib/auth/kratos-client.ts` – Kratos API client (login flow, session, logout)
- `front/src/lib/auth/AuthProvider.tsx` – React context provider managing auth state
- `front/src/lib/auth/useAuth.ts` – Hook to access auth state and functions
- `front/src/lib/auth/ProtectedRoute.tsx` – Route wrapper that redirects to `/login` if unauthenticated
- `front/src/pages/auth/LoginPage.tsx` – Login form component

**Authentication flow:**
1. On app load, `AuthProvider` checks for existing Kratos session via `/auth/kratos/sessions/whoami`
2. Protected routes (`/` and `/app/*`) redirect to `/login` if no valid session
3. Login form creates a Kratos browser flow, submits credentials, receives session cookie
4. Session is stored in `ory_kratos_session` cookie on `localtest.me` domain
5. Logout clears the session via Kratos logout flow

**Using auth in components:**
```typescript
import { useAuth } from '../lib/auth';

function MyComponent() {
  const { user, isAuthenticated, login, logout } = useAuth();
  // user.traits.email, user.traits.name.first, etc.
}
```

**Adding new protected routes:**
Routes with `layout: 'app'` are protected by default. To make a route public, set `protected: false`.

## Specs & data model

- Primary API definition: `spec/openapi.yaml` with supporting fragments in `spec/components/` and `spec/paths/`.
- Backend contract is enforced through ogen. Keep the spec authoritative; update it before changing handlers.
- SQL schema consolidated in `db/schema.sql` plus Telegram-specific SQL in `db/tg.sql`. Run `make sync-db` after edits to apply them against containers.
- Atlas is used in-place; there are no separate migration files. Ensure schema changes are reflected in both SQL files and code.
- Workers operate over a queue prefix defined by `BE_QUEUE_PREFIX` (default `shadowapi`) with NATS JetStream enabled.

## Local development & tooling

### Getting started

1. Run `make up` first to bootstrap the project (generates secrets, starts db, runs migrations, creates OAuth client and test user).
2. Start the development environment with `docker compose watch`.
   - Backend runs in the container via `air` auto-rebuild.
   - Test login: `admin@example.com` / `Admin123!`

### Make targets

Run `make help` to see all available targets. Key ones:

- `make up` – Bootstrap and start the full stack (generates secrets, creates OAuth client, test user). **Run this first before starting development.**
- `make init` – Reset containers and reinitialize database (destructive).
- `make sync-db` – Apply schema to the running Postgres (uses Atlas).
- `make api-gen`, `make api-gen-backend` – Sync code with the OpenAPI spec.

### Compose topology

- Traefik routes requests based on path prefix with priority:
  - `/.well-known/*` (priority 40) → Hydra OAuth2/OIDC discovery (port 4444)
  - `/auth/kratos/*` (priority 35) → Kratos identity API (port 4433)
  - `/oauth2/*` (priority 35) → Hydra OAuth2 endpoints (port 4444)
  - `/api/`, `/assets/`, `/auth/` (priority 30) → Backend container (port 8080)
  - `/page/*` (priority 20) → SSR container (port 3000)
  - Everything else (priority 1) → Frontend container (port 5173)
- Access the app at `http://localtest.me`
- Postgres, NATS, and supporting containers share the `shadowapi` network; Atlas (`db-migrate`) runs on startup to sync schema.

### Ory Authentication Stack

The project includes Ory Kratos (identity management) and Ory Hydra (OAuth2/OIDC) for authentication:

**Services:**
- **kratos** (v1.3.1) – Identity management with password, TOTP, WebAuthn, and code-based authentication
- **hydra** (v2.2.0) – OAuth2/OIDC provider for token issuance
- **hydra-consent** – Bridges Kratos identity to Hydra OAuth2 consent flow
- **mailhog** – Email testing server (web UI at `http://localhost:8025`)

**Configuration files:**
- `devops/ory/kratos/kratos.yaml` – Kratos configuration with self-service flows and OAuth2 provider integration
- `devops/ory/kratos/identity.schema.json` – Identity schema (email required, optional name)
- `devops/ory/hydra/hydra.yaml` – Hydra OAuth2/OIDC configuration

**Databases:**
- Kratos uses `kratos` database (created by `devops/compose/infra/db-init.sh`)
- Hydra uses `hydra` database (created by `devops/compose/infra/db-init.sh`)

**Environment variables (in `.env`):**
- `KRATOS_DSN`, `KRATOS_SECRETS_DEFAULT`, `KRATOS_SECRETS_COOKIE` – Kratos database and secrets
- `HYDRA_DSN`, `HYDRA_SECRETS_SYSTEM`, `OIDC_PAIRWISE_SALT` – Hydra database and secrets
- `HYDRA_URLS_*` – Hydra URL configuration for login, consent, logout flows

**Testing the setup:**
```bash
# Check Kratos health
curl http://localtest.me/auth/kratos/health/alive

# Check Hydra OIDC discovery
curl http://localtest.me/.well-known/openid-configuration

# Create a test user via Kratos Admin API
curl -X POST http://localhost:4434/admin/identities \
  -H "Content-Type: application/json" \
  -d '{
    "schema_id": "default",
    "traits": {"email": "test@example.com"},
    "credentials": {"password": {"config": {"password": "testpassword123"}}}
  }'

# Create a test OAuth2 client
docker compose exec hydra hydra create client \
  --endpoint http://localhost:4445 \
  --grant-type authorization_code,refresh_token \
  --response-type code \
  --scope openid,offline_access \
  --redirect-uri http://localtest.me/callback \
  --name "Test Client" \
  --format json
```

**Frontend login page:** `http://localtest.me/login`

## Testing & QA expectations

- **Go:** Prefer running `go test ./...` in the backend directory. Add focused tests under `*_test.go` when fixing bugs or adding features.
- **SQL:** After schema edits, run `sqlc vet` in the backend directory to ensure queries remain valid.
- **Integration:** When modifying auth or message pipelines, verify the running stack (`docker compose watch`) and exercise flows through the API.
- Never leave generated files stale—regenerate them in the same change set.

## Coding standards & gotchas

- Keep code self-documenting; only add comments for intent or non-obvious reasoning.
- Run `gofmt`/`goimports` on Go files.
- Do not edit generated directories (`backend/pkg/api`, `backend/pkg/query`) manually.
- Wire new services through the DI container (`samber/do`) so they can be mocked and tested.
- Reuse existing logging helpers (`backend/internal/log`) and metric emitters rather than introducing new logging styles.
- Respect existing channel abstractions (`internal/tg`, `internal/whatsapp`, etc.) when adding integrations—keep protocols isolated from handler code.
- Keep secrets and credentials out of source control. Use `.env`, `backend/config.yaml`, or `secrets/` volumes instead.

## Known issues

- **Token validation:** The backend currently does not properly validate authentication tokens. This is a known security issue that needs to be addressed before production deployment.

## Contribution workflow

1. Plan the change: update the OpenAPI spec or SQL schema first if the contract changes.
2. Update backend code, regenerate artifacts (`make api-gen`, `sqlc generate`) as needed, and ensure DI wiring stays consistent.
3. Run the relevant test and lint commands from the sections above.
4. Update documentation (README, docs, or inline help) when behaviour changes.
5. Use conventional commit prefixes (`feat:`, `fix:`, `docs:`, etc.). Do not mention Claude or other assistants in commit messages.

## Working style for Claude Code

- Always read the surrounding files (`README.md`, related packages) before editing; prefer `rg` for searching within the repo.
- Keep diffs tight and deliberate; do not refactor broadly unless explicitly asked.
- Ask the user when scope is unclear or when destructive changes (schema rewrites, data deletion) are required.
- Avoid running commands that need interactive input or elevated privileges unless instructed. Use the provided `make` targets whenever possible.
- If a change touches unfamiliar areas (auth, worker pipelines, storage), leave breadcrumbs in the PR description or docs instead of code comments.
- When in doubt, stop and request guidance rather than guessing.
