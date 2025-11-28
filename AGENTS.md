# Repository Guidelines

## Project Structure & Module Organization
- `backend/`: Go service (`cmd/shadowapi`, `internal`, `pkg`, `sdk-go`). Runtime config in `backend/config.yaml`.
- `front/`: React + Vite + TypeScript app. Tests under `front/tests`, config in `playwright.config.ts`.
- `spec/`: OpenAPI definition (`openapi.yaml`).
- `db/`: Database schema (`schema.sql`, `tg.sql`), applied with Atlas.
- `devops/`: Dockerfiles and Compose includes; root `compose.yaml` wires infra + Zitadel.
- Other: `docs/`, `k8s/`, `.env(.example)`, `Taskfile.yml` for common tasks.

## Build, Test, and Development Commands
- Init environment: `task init` (build images, install deps, seed/configure services).
- Run stack (dev): `docker compose up --build` (frontend, backend w/ hot reload, DB, NATS, Traefik, Zitadel). App: `http://localtest.me`.
- Migrate DB: `task sync-db` (applies `db/schema.sql` + `db/tg.sql` via Atlas).
- Generate API types: `task api-gen` or `cd front && npm run generate-api-client`.
- E2E tests: `task playwright-run` and `task playwright-report`.
- Optional: provision Zitadel OIDC app with `task zitadel-init` (no longer part of `task init`).

## Auth Flow (OIDC)
- Login uses OIDC redirect to the provider (Zitadel by default).
- Backend handles start and callback outside of the generated Ogen router:
  - `GET /api/v1/auth/login` — starts the flow, redirects to the IDP authorize URL.
  - `GET /api/v1/auth/callback` — exchanges `code` for tokens and stores a browser session via `sessionStorage`.
  - Alias supported: `GET /auth/callback` (useful if your ZITADEL app is already configured with this URI).
- Provider abstraction lives in `backend/internal/idp` (see `idp.Provider`).
  - Current implementation wraps Zitadel (`backend/internal/zitadel`).
  - To add another IDP (e.g., Auth0), implement `Provider` and wire it in `idp.NewProvider`.
- Frontend login page (`front/src/shauth/LoginPage.tsx`) simply redirects to `/api/v1/auth/login` and no longer performs username/password against Zitadel APIs directly. The previous UI code remains in the repo, but is not used.
- Security middleware (`backend/internal/auth/auth.go`) allows unauthenticated access to the above auth endpoints.

Notes for contributors:
- These auth endpoints are not part of the OAS yet; they are pre‑router handlers in `backend/internal/server/server.go`. If you expose them via OpenAPI, update `spec/` and regenerate with `task api-gen`.
- The in‑memory `state` store for the login flow is intentionally simple for dev; if you improve it (cookies/DB/PKCE), keep the abstraction and backward compatibility.
- Keep `sessionStorage` key as `shadowapi_auth` (used by `front/src/api/client.ts` middleware) unless you update all consumers.

## Coding Style & Naming Conventions
- Go: `gofmt`/`goimports` formatting; packages lowercase; exported identifiers `PascalCase`; errors wrapped (`fmt.Errorf("…: %w", err)`); context first param `ctx context.Context`.
- TypeScript/React: ESLint + Prettier (2 spaces, single quotes, no semicolons, trailing commas, 120 cols). Prefer `const`; unused args prefixed `_`; React Hooks and TanStack Query rules enabled.
- Filenames: components `PascalCase.tsx`; hooks `useX.ts`; utilities `*.ts`.

## Testing Guidelines
- Frontend: Playwright via `npx playwright test` (or `task playwright-run`). Keep tests independent; add screenshots for UI changes.
- Backend: standard Go tests (`go test ./...`) where applicable. Aim for meaningful coverage on new/changed code.

## Commit & Pull Request Guidelines
- Use Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`) as in git history.
- PRs include: clear description, linked issues, screenshots for UI, repro steps for bugs, and notes on config changes (`.env`, `backend/config.yaml`). Update docs/spec when APIs change.

## Security & Configuration Tips
- Do not commit secrets. Use the `secrets` volume and local files only.
- Copy `.env.example` → `.env` and `backend/config.example.yaml` → `backend/config.yaml` before running.
- Local auth console: `auth.localtest.me` (proxied via Traefik).
 - Backend config keys to check for OIDC:
   - `auth.user_manager` — set to `zitadel` to enable Zitadel user manager paths elsewhere.
   - `auth.zitadel.instance_url`, `auth.zitadel.audience`, `auth.zitadel.redirect_uri` — required for the OIDC login flow. If `redirect_uri` is empty, backend uses `{base_url}/api/v1/auth/callback`.
 - Do not remove the legacy email/password UI code; it is intentionally kept for future use.
