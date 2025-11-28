# CLAUDE.md

This document guides Claude Code (claude.ai/code) when working inside the ShadowAPI repository. Follow it before you change or create files.

## Snapshot

- ShadowAPI is a unified messaging platform that normalises Gmail, Telegram, WhatsApp, and LinkedIn into a single REST + MCP surface.
- Backend: Go 1.24 service using Cobra CLI, samber/do dependency injection, ogen-generated handlers, SQLC + PostgreSQL 16, NATS JetStream, S3/host/local storage abstractions.
- Frontend: React 19 + TypeScript via Vite 6, Adobe React Spectrum, TanStack Query 5, Zustand state, openapi-fetch typed clients, and Playwright for E2E.
- Infrastructure: Docker Compose stack with Traefik, backend, frontend, Postgres, and NATS; orchestrated by Make (`Makefile`).

## Repository map

- `backend/` – Go service. `cmd/shadowapi` hosts the CLI (`serve`, `loader`, `reset-password`), `internal/` holds domain packages (auth, handler, worker, storages, queue, metrics, config), `pkg/api` is ogen-generated HTTP server (do not edit manually), `pkg/query` is SQLC output (do not edit), and `sdk-go/` contains a generated client example.
- `front/` – React application (`src/pages`, `src/forms`, `src/components`, `src/hooks`). `package.json` defines scripts and dependencies.
- `spec/` – OpenAPI definition (`openapi.yaml`) plus ogen configuration and shared pieces under `components/` and `paths/`.
- `db/` – Canonical SQL schema in `schema.sql` with Telegram-specific relations in `tg.sql`. Atlas migrations are applied from these files; no separate migration files exist.
- `devops/` – Dockerfiles, Atlas configs, sqlc builder image, and helper scripts used by Compose/Make.
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

- Run `make api-gen` after editing `spec/`. This updates the Go server and the frontend TypeScript types.
- Run `sqlc generate` (and `sqlc vet`) in the backend directory after updating the SQL schema. SQL lives in `db/schema.sql` + `db/tg.sql`.
- `make sync-db` concatenates schema files and applies them via Atlas to the running Postgres instance.

### Command-line entrypoints

- `go build -o ./bin/shadowapi ./cmd/shadowapi` builds the binary locally.
- `shadowapi serve` is the main API server when running outside Docker.

## Frontend

### Stack & conventions

- React 19 with TypeScript 5, Vite 6, and module resolution managed by bare modules.
- Adobe React Spectrum provides UI primitives; stay consistent with Spectrum design tokens.
- Data fetching relies on `openapi-fetch` + TanStack Query; global state handled by `zustand`.
- Forms predominantly use `react-hook-form`.

### Structure

- `src/pages/` – Route-level screens (messages, pipelines, storages, schedulers, etc.).
- `src/forms/` – Form abstractions bound to API entities.
- `src/api/v1.d.ts` – Generated types; do not edit manually. Regenerate via `make api-gen-frontend` or `npm run generate-api-client`.
- `src/components/`, `src/layouts/`, `src/hooks/` – Shared UI/layout/state helpers.

### Scripts

- `npm install` handled by `make init`.
- `npm run dev` (inside `front/`) starts Vite when running outside Docker; use `docker compose watch` for the full stack.
- `npm run lint`, `npm run build:tscheck`, and `npx playwright test` are the main QA commands.

## Specs & data model

- Primary API definition: `spec/openapi.yaml` with supporting fragments in `spec/components/` and `spec/paths/`.
- Backend and frontend contract is enforced through ogen + openapi-typescript. Keep the spec authoritative; update it before changing handlers or UI shapes.
- SQL schema consolidated in `db/schema.sql` plus Telegram-specific SQL in `db/tg.sql`. Run `make sync-db` after edits to apply them against containers.
- Atlas is used in-place; there are no separate migration files. Ensure schema changes are reflected in both SQL files and code.
- Workers operate over a queue prefix defined by `BE_QUEUE_PREFIX` (default `shadowapi`) with NATS JetStream enabled.

## Local development & tooling

### Getting started

1. Run `make init` first to initialize the project (resets containers, copies env, starts db, runs migrations).
2. Start the development environment with `docker compose watch`.
   - Backend runs in the container via `air` auto-rebuild.
   - Frontend changes are handled by Docker watch mechanism.

### Make targets

Run `make help` to see all available targets. Key ones:

- `make init` – Initialize the project (reset containers, copy env, start db, migrate). **Run this first before starting development.**
- `make sync-db` – Apply schema to the running Postgres (uses Atlas).
- `make api-gen`, `make api-gen-backend`, `make api-gen-frontend` – Sync code with the OpenAPI spec.

### Compose topology

- Traefik exposes `http://localtest.me` to the frontend and `http://localtest.me/api` to the backend.
- Postgres, NATS, and supporting containers share the `shadowapi` network; Atlas (`db-migrate`) runs on startup to sync schema.

## Testing & QA expectations

- **Go:** Prefer running `go test ./...` in the backend directory. Add focused tests under `*_test.go` when fixing bugs or adding features.
- **SQL:** After schema edits, run `sqlc vet` in the backend directory to ensure queries remain valid.
- **Frontend:** Run `npm run lint` and `npm run build:tscheck`. Execute `npx playwright test` (or `make playwright-run`) when UI flows change.
- **Integration:** When modifying auth or message pipelines, verify the running stack (`docker compose watch`) and exercise flows through the UI or API.
- Never leave generated files stale—regenerate them in the same change set.

## Coding standards & gotchas

- Keep code self-documenting; only add comments for intent or non-obvious reasoning.
- Run `gofmt`/`goimports` on Go files, and rely on project ESLint/TypeScript configs for frontend files.
- Do not edit generated directories (`backend/pkg/api`, `backend/pkg/query`, `front/src/api/v1.d.ts`) manually.
- Wire new services through the DI container (`samber/do`) so they can be mocked and tested.
- Reuse existing logging helpers (`backend/internal/log`) and metric emitters rather than introducing new logging styles.
- Respect existing channel abstractions (`internal/tg`, `internal/whatsapp`, etc.) when adding integrations—keep protocols isolated from handler code.
- Keep secrets and credentials out of source control. Use `.env`, `backend/config.yaml`, or `secrets/` volumes instead.

## Known issues

- **Token validation:** The backend currently does not properly validate authentication tokens. This is a known security issue that needs to be addressed before production deployment.
- **Frontend authentication:** The frontend authentication is currently broken and needs to be reimplemented.

## Contribution workflow

1. Plan the change: update the OpenAPI spec or SQL schema first if the contract changes.
2. Update backend/frontend code, regenerate artifacts (`make api-gen`, `sqlc generate`) as needed, and ensure DI wiring stays consistent.
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
