# Repository Guidelines

## Project Structure & Module Organization
- `backend/`: Go service (`cmd/shadowapi`, `internal`, `pkg`, `sdk-go`). Runtime config in `backend/config.yaml`.
- `front/`: React + Vite + TypeScript app. Tests under `front/tests`, config in `playwright.config.ts`.
- `spec/`: OpenAPI definition (`openapi.yaml`).
- `db/`: Database schema (`schema.sql`, `tg.sql`), applied with Atlas.
- `devops/`: Dockerfiles and Compose includes; root `compose.yaml` wires infra.
- Other: `docs/`, `k8s/`, `.env(.example)`, `Makefile` for common tasks.

## Build, Test, and Development Commands
- Init environment: `make init` (build images, install deps, seed/configure services).
- Run stack (dev): `docker compose watch` (frontend, backend w/ hot reload, DB, NATS, Traefik). App: `http://localtest.me`.
- Migrate DB: `make sync-db` (applies `db/schema.sql` + `db/tg.sql` via Atlas).
- Generate API types: `make api-gen` or `cd front && npm run generate-api-client`.
- E2E tests: `make playwright-run` and `make playwright-report`.

## Auth Flow
- Authentication is currently broken and needs reimplementation.
- Backend auth middleware is in `backend/internal/auth/auth.go`.
- User management is database-based via `backend/internal/auth/dbauth`.

## Coding Style & Naming Conventions
- Go: `gofmt`/`goimports` formatting; packages lowercase; exported identifiers `PascalCase`; errors wrapped (`fmt.Errorf("…: %w", err)`); context first param `ctx context.Context`.
- TypeScript/React: ESLint + Prettier (2 spaces, single quotes, no semicolons, trailing commas, 120 cols). Prefer `const`; unused args prefixed `_`; React Hooks and TanStack Query rules enabled.
- Filenames: components `PascalCase.tsx`; hooks `useX.ts`; utilities `*.ts`.

## Testing Guidelines
- Frontend: Playwright via `npx playwright test` (or `make playwright-run`). Keep tests independent; add screenshots for UI changes.
- Backend: standard Go tests (`go test ./...`) where applicable. Aim for meaningful coverage on new/changed code.

## Commit & Pull Request Guidelines
- Use Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`) as in git history.
- PRs include: clear description, linked issues, screenshots for UI, repro steps for bugs, and notes on config changes (`.env`, `backend/config.yaml`). Update docs/spec when APIs change.

## Security & Configuration Tips
- Do not commit secrets. Use the `secrets` volume and local files only.
- Copy `.env.example` → `.env` and `backend/config.example.yaml` → `backend/config.yaml` before running.
