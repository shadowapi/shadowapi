# CLAUDE.md

This document guides Claude Code (claude.ai/code) when working inside the ShadowAPI repository. Follow it before you change or create files.

## Snapshot

- ShadowAPI is a unified messaging platform that normalises Gmail, Telegram, WhatsApp, and LinkedIn into a single REST + MCP surface.
- Backend: Go 1.24 service using Cobra CLI, samber/do dependency injection, ogen-generated handlers, SQLC + PostgreSQL 16, NATS JetStream, S3/host/local storage abstractions, and Zitadel-powered OAuth2.
- Frontend: React 19 + TypeScript via Vite 6, Adobe React Spectrum, TanStack Query 5, Zustand state, openapi-fetch typed clients, and Playwright for E2E.
- Infrastructure: Docker Compose stack with Traefik, backend, frontend, Postgres, NATS, and Zitadel; orchestrated by Task (`Taskfile.yml`).
- Authentication flows rely on Zitadel (OIDC) with an optional local admin fallback via `shadowapi reset-password`.

## Repository map

- `backend/` – Go service. `cmd/shadowapi` hosts the CLI (`serve`, `loader`, `reset-password`), `internal/` holds domain packages (auth, handler, worker, storages, queue, metrics, config), `pkg/api` is ogen-generated HTTP server (do not edit manually), `pkg/query` is SQLC output (do not edit), and `sdk-go/` contains a generated client example.
- `front/` – React application (`src/pages`, `src/forms`, `src/components`, `src/shauth`, `src/hooks`). `package.json` defines scripts and dependencies.
- `spec/` – OpenAPI definition (`openapi.yaml`) plus ogen configuration and shared pieces under `components/` and `paths/`.
- `db/` – Canonical SQL schema in `schema.sql` with Telegram-specific relations in `tg.sql`. Atlas migrations are applied from these files; no separate migration files exist.
- `devops/` – Dockerfiles, Atlas/zitadel bootstrap configs, sqlc builder image, and helper scripts used by Compose/Task.
- `docs/` – Product documentation and screenshots referenced by the README.
- `templates/`, `start/`, `k8s/`, `secrets/` – Supporting assets (UI templates, bootstrap scripts, Kubernetes manifests, local Zitadel keys). Leave anything under `secrets/` untouched.
- `Taskfile.yml` – Source of all task runner commands; prefer calling `task <name>` over shelling out to `docker compose` directly.

## Backend

### Stack & runtime

- Go 1.24 with modules pinned in `backend/go.mod`.
- Dependency injection via `github.com/samber/do/v2`; resolve services through the container rather than new-ing them manually.
- HTTP surface is generated from the OpenAPI spec by `ogen`, wired through `backend/internal/server`.
- Long-running jobs and pipelines live under `backend/internal/worker` and communicate through NATS (`backend/internal/queue`).
- Observability uses `log/slog`, OpenTelemetry (`go.opentelemetry.io/otel`), and Prometheus metrics (`backend/internal/metrics`).

### Key packages

- `internal/handler` – Implements ogen handlers for messages, contacts, pipelines, storages, auth callbacks, etc.
- `internal/auth` + `internal/zitadel` – Login flows, token validation, and Zitadel management API calls.
- `internal/storages` – Pluggable storage backends (Postgres, S3, host filesystem).
- `internal/worker` – Pipelines, extractors, filters, schedulers, job registry, and cancellation logic.
- `internal/imap`, `internal/whatsapp`, `internal/tg`, `internal/oauth2` – External channel integrations (IMAP/SMTP Gmail, WhatsApp via whatsmeow, Telegram via gotd, LinkedIn helpers).
- `pkg/api` – Generated server; never edit by hand. Regenerate with `task api-gen-backend`.
- `pkg/query` – SQLC-generated database accessors; regenerate with `task sqlc`.

### Config & secrets

- Default configuration lives in `backend/config.example.yaml`; copy to `backend/config.yaml` for local overrides.
- Environment variables are SA-prefixed (see `backend/internal/config/config.go`). Important ones include `SA_DB_URI`, `SA_QUEUE_URL`, `SA_ZITADEL_INSTANCE_URL`, `SA_AUTH_USER_MANAGER`, and `TG_APP_ID`/`TG_APP_HASH`.
- Sensitive Zitadel keys are read from `secrets/`; the repo should never gain new secrets in git history.

### Generation & migrations

- Run `task api-gen` after editing `spec/`. This updates the Go server and the frontend TypeScript types.
- Run `task sqlc` (and `task sqlc-vet`) after updating the SQL schema. SQL lives in `db/schema.sql` + `db/tg.sql`.
- `task sync-db` concatenates schema files and applies them via Atlas to the running Postgres instance.

### Command-line entrypoints

- `task build-api` builds `./bin/shadowapi` locally.
- `shadowapi serve` is the main API server when running outside Docker.
- `shadowapi loader` seeds demo data/jobs.
- `shadowapi reset-password` resets non-Zitadel admin passwords.

## Frontend

### Stack & conventions

- React 19 with TypeScript 5, Vite 6, and module resolution managed by bare modules.
- Adobe React Spectrum provides UI primitives; stay consistent with Spectrum design tokens.
- Data fetching relies on `openapi-fetch` + TanStack Query; global state handled by `zustand`.
- Forms predominantly use `react-hook-form`.

### Structure

- `src/pages/` – Route-level screens (messages, pipelines, storages, schedulers, etc.).
- `src/forms/` – Form abstractions bound to API entities.
- `src/shauth/` – Zitadel auth context, login/signup screens, and hooks (`useZitadelAuth`).
- `src/api/v1.d.ts` – Generated types; do not edit manually. Regenerate via `task api-gen-frontend` or `npm run generate-api-client`.
- `src/components/`, `src/layouts/`, `src/hooks/` – Shared UI/layout/state helpers.

### Scripts

- `npm install` handled by `task init`.
- `npm run dev` (inside `front/`) starts Vite when running outside Docker; use `task dev-up` for the full stack.
- `npm run lint`, `npm run build:tscheck`, and `npx playwright test` are the main QA commands.

## Specs & data model

- Primary API definition: `spec/openapi.yaml` with supporting fragments in `spec/components/` and `spec/paths/`.
- Backend and frontend contract is enforced through ogen + openapi-typescript. Keep the spec authoritative; update it before changing handlers or UI shapes.
- SQL schema consolidated in `db/schema.sql` plus Telegram-specific SQL in `db/tg.sql`. Run `task sync-db` after edits to apply them against containers.
- Atlas is used in-place; there are no separate migration files. Ensure schema changes are reflected in both SQL files and code.
- Workers operate over a queue prefix defined by `SA_QUEUE_PREFIX` (default `shadowapi`) with NATS JetStream enabled.

## Local development & tooling

### Task runner essentials

- `task init` – Build sqlc helper image and install frontend dependencies.
- `task dev-up` / `task dev-down` – Start/stop the full Docker Compose stack with live reload (Traefik, backend, frontend, Postgres, NATS, Zitadel).
- `task shell` – Open a shell inside the backend container (useful for running `go test`, `shadowapi serve`, etc.).
- `task db-shell` – Drop into Postgres with the project user.
- `task sync-db` – Apply schema to the running Postgres (uses Atlas).
- `task sqlc` / `task sqlc-vet` – Regenerate and validate SQLC output.
- `task api-gen`, `task api-gen-backend`, `task api-gen-frontend` – Sync code with the OpenAPI spec.
- `task playwright-run` / `task playwright-report` – Run frontend E2E tests and view reports.
- `task build-api` – Compile the backend binary.
- `task clean` – Tear down the dev stack and remove volumes/images (irreversible for local data).
- `task prod-up` / `task prod-down` – Build and run the production docker-compose profile.
- `task clean-zitadel-keys` – Purge generated Zitadel service keys from `secrets/`.

### Compose topology

- Traefik exposes `http://localtest.me` to the frontend and `http://localtest.me/api` to the backend.
- Zitadel is reachable at `http://auth.localtest.me` during development.
- Postgres, NATS, and supporting containers share the `shadowapi` network; Atlas (`db-migrate`) runs on startup to sync schema.

## Testing & QA expectations

- **Go:** Prefer running `go test ./...` inside `task shell`. Add focused tests under `*_test.go` when fixing bugs or adding features.
- **SQL:** After schema edits, run `task sqlc-vet` to ensure queries remain valid.
- **Frontend:** Run `npm run lint` and `npm run build:tscheck`. Execute `npx playwright test` (or `task playwright-run`) when UI flows change.
- **Integration:** When modifying auth or message pipelines, verify the running stack (`task dev-up`) and exercise flows through the UI or API.
- Never leave generated files stale—regenerate them in the same change set.

## Coding standards & gotchas

- Keep code self-documenting; only add comments for intent or non-obvious reasoning.
- Run `gofmt`/`goimports` on Go files, and rely on project ESLint/TypeScript configs for frontend files.
- Do not edit generated directories (`backend/pkg/api`, `backend/pkg/query`, `front/src/api/v1.d.ts`) manually.
- Wire new services through the DI container (`samber/do`) so they can be mocked and tested.
- Reuse existing logging helpers (`backend/internal/log`) and metric emitters rather than introducing new logging styles.
- Respect existing channel abstractions (`internal/tg`, `internal/whatsapp`, etc.) when adding integrations—keep protocols isolated from handler code.
- Keep secrets and credentials out of source control. Use `.env`, `backend/config.yaml`, or `secrets/` volumes instead.

## Contribution workflow

1. Plan the change: update the OpenAPI spec or SQL schema first if the contract changes.
2. Update backend/frontend code, regenerate artifacts (`task api-gen`, `task sqlc`) as needed, and ensure DI wiring stays consistent.
3. Run the relevant test and lint commands from the sections above.
4. Update documentation (README, docs, or inline help) when behaviour changes.
5. Use conventional commit prefixes (`feat:`, `fix:`, `docs:`, etc.). Do not mention Claude or other assistants in commit messages.

## Working style for Claude Code

- Always read the surrounding files (`README.md`, related packages) before editing; prefer `rg` for searching within the repo.
- Keep diffs tight and deliberate; do not refactor broadly unless explicitly asked.
- Ask the user when scope is unclear or when destructive changes (schema rewrites, data deletion) are required.
- Avoid running commands that need interactive input or elevated privileges unless instructed. Use the provided `task` commands whenever possible.
- If a change touches unfamiliar areas (auth, worker pipelines, storage), leave breadcrumbs in the PR description or docs instead of code comments.
- When in doubt, stop and request guidance rather than guessing.

## Authentication

- we use Zitadel as authentication IDP
- to build custom login&auth form we use this flow https://zitadel.com/docs/guides/integrate/login-ui/username-password