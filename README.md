# ShadowAPI

ShadowAPI is a **unified messaging API** that enables seamless integration with Gmail,
Telegram, WhatsApp, and LinkedIn in your applications.

It provides a single interface for managing both personal and team‑shared messages
across platforms, letting you tag, process, and expose communications via REST endpoints,
large language models (LLMs), or message‑centric processing (MCP) workflows.

## Development Setup

We use Make to manage the project. Make sure Make and Docker Compose are both installed.

### Prerequisites

#### SOPS and age (for secrets management)

Production secrets are encrypted with [SOPS](https://github.com/getsops/sops) and [age](https://github.com/FiloSottile/age).

**macOS (Homebrew):**
```bash
brew install sops age
```

**Ubuntu/Debian:**
```bash
# Install age
sudo apt install age

# Install SOPS (download latest release)
SOPS_VERSION=$(curl -s https://api.github.com/repos/getsops/sops/releases/latest | grep tag_name | cut -d '"' -f 4)
curl -LO "https://github.com/getsops/sops/releases/download/${SOPS_VERSION}/sops-${SOPS_VERSION}.linux.amd64"
sudo mv sops-${SOPS_VERSION}.linux.amd64 /usr/local/bin/sops
sudo chmod +x /usr/local/bin/sops
```

**Generate your age key:**
```bash
mkdir -p ~/.config/sops/age
age-keygen -o ~/.config/sops/age/keys.txt
```

Share your public key (starts with `age1...`) with the team to get access to encrypted secrets.
Base domain is `localtest.me`, which resolves to `localhost`.
Traefik routes domain based traffic to the correct service.

Run `make help` to see all available targets.

### 1. Initialize

```bash
cp .env.example .env
# make sure to adjust the values
make init

# generate .env file for local development and up docker compose
make up
```

### 2. Start Development Environment

```bash
docker compose watch
```

- Spins up all services in development mode (frontend, backend with hot reload,
  database, etc.).

### 3. Apply Database Migrations

```bash
make sync-db
```

- Applies migrations against the development database.

### 4. Access the App

- Open your browser at [http://localtest.me](http://localtest.me).

### 5. Stopping the Development Environment

```bash
docker compose down
```

### Cleaning up Volumes
```bash
docker compose down -v
```
