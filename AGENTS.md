# Repository Guidelines

## Project Structure & Module Organization
- `backend/` contains Go services; the entrypoint is `cmd/shadowapi`, domain logic sits in `internal`, and generated clients live in `sdk-go`.
- The React frontend is under `front/`, with routed views in `src`, static assets in `public`, and Playwright specs in `tests`.
- API contracts reside in `spec/openapi.yaml`; regeneration writes to `backend/pkg/api` and `front/src/api`.
- Ops artefacts live in `devops/`, `k8s/`, and `db/` (SQL migrations); reference docs sit in `docs/`.

## Build, Test, and Development Commands
- `task init` copies `.env` and runs the bootstrap profile (it wipes local volumes, so use intentionally).
- `docker compose watch` starts the full stack with hot reload.
- `task sync-db` concatenates `db/schema.sql` and `db/tg.sql` and applies them via Atlas to Postgres.
- `task api-gen` updates both Go stubs and the TypeScript client after spec edits.
- `npm run dev` inside `front/` launches Vite against the existing backend.

## Coding Style & Naming Conventions
- Go code must remain `gofmt`-clean (tabs, grouped imports); prefer interfaces named `FooService` and files matching the feature (`message_store.go`).
- TypeScript uses 2-space indentation, `camelCase` for functions, and `PascalCase` for React components; colocate modules under `front/src/features/<name>/`.
- Run `npm run lint` before submitting UI changes; ESLint + Prettier enforce spacing and import order.

## Testing Guidelines
- Backend unit suites run with `cd backend && go test ./...`; add table-driven tests near the code they cover and mock external APIs via helpers in `internal/test`.
- UI end-to-end coverage lives in `front/tests`; execute `task playwright-run` or `npx playwright test` for the default mocked Zitadel workflow.
- Preserve failing artefacts in `front/test-results/` (gitignored) and attach screenshots to PRs when behaviour changes.

## Commit & Pull Request Guidelines
- Follow Conventional Commit headers (`feat:`, `fix:`, `chore:`) as seen in `git log`; keep subjects under 72 characters.
- Squash locally before opening a PR and include a summary, linked issue, migration notes, and verification commands.
- PRs touching auth or schema must mention required `.env` changes and include a screenshot or short clip of the updated flow.

## Security & Configuration Tips
- Never commit secrets; rely on `.env` (copied from `.env.example`) and, when needed, `compose.override.yaml` for overrides.
- Regenerate local TLS certificates via `task local-certs` before testing HTTPS flows.
