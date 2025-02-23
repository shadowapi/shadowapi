# ShadowAPI

ShadowAPI is a spin-off of Reactima CRM (https://reactima.com/).

If you're interested in streaming all your communications from IMAP, Telegram, WhatsApp, and Signal into your system—tagging, processing, and making them available for LLM usage—let's develop it together.

Your feature requests will be prioritized if you can sponsor 20 hours of development per week. Project ownership, licences, and other details can be discussed.

Please contact [@reactima](https://github.com/reactima) for details.

## Development Setup

We use [Task](https://taskfile.dev/installation/) instead of traditional Makefiles to manage the project. Make sure Task and Docker Compose are both installed.

### 1. Initialize

```bash
task init
```

- Builds necessary images (including an image for SQLC).
- Installs frontend dependencies.

### 2. Start Development Environment

```bash
task dev-up
```

- Spins up all services in development mode (frontend, backend with hot reload, database, etc.).

### 3. Apply Database Migrations

```bash
task sync-db
```

- Applies migrations against the development database.

### 4. Access the App

- Open your browser at [http://localtest.me](http://localtest.me).
- You can sign up via [http://localtest.me/signup](http://localtest.me/signup).

### 5. Stopping the Development Environment

```bash
task dev-down
```

---

## Additional Commands

### Resetting the Development Environment

```bash
task clean
```

- Brings down the dev environment, removes volumes, images, and orphan containers.
- **Warning:** This permanently deletes all data in the development PostgreSQL database.

### Production Build and Run

```bash
task prod-up
```

- Builds optimized production images and spins up containers using `docker-compose.prod.yaml`.

```bash
task prod-down
```

- Stops and removes the production environment containers.

---

## Common Tasks

- **Open Shell in Backend Container:** `task shell`
- **Open Postgres Shell:** `task db-shell`
- **Regenerate SQL Queries (SQLC):** `task sqlc`
- **Generate API Specs (Backend + Frontend):** `task api-gen`
- **Run Playwright Tests (Frontend):** `task playwright-run`
