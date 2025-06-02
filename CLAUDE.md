# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ShadowAPI is a unified messaging API that enables seamless integration with Gmail, Telegram, WhatsApp, and LinkedIn. It provides a single interface for managing both personal and team-shared messages across platforms with REST endpoints, LLM integration, and message-centric processing workflows.

## Technology Stack

**Backend:**
- Go 1.24 with standard project layout
- PostgreSQL database with SQLC for type-safe queries
- NATS.io for message queuing and event-driven architecture
- Ogen for OpenAPI code generation
- IMAP/SMTP for email integration
- OAuth2 for external service authentication
- Ory Kratos for identity management

**Frontend:**
- React 19 with TypeScript
- Vite for development and building
- Adobe React Spectrum for UI components
- TanStack Query for data fetching
- React Router for navigation
- Playwright for end-to-end testing

**Infrastructure:**
- Docker Compose for development environment
- Traefik for reverse proxy and routing
- Atlas for database migrations
- Task runner instead of Make

## Essential Development Commands

### Environment Setup
```bash
# Copy environment template and configure
cp .env.example .env
# Initialize project (builds images, installs deps)
task init
```

### Development Workflow
```bash
# Start development environment with hot reload
task dev-up
# Apply database migrations
task sync-db
# Access app at http://localtest.me
# Stop development environment
task dev-down
```

### Code Generation
```bash
# Regenerate SQL queries from schema
task sqlc
# Generate API specs (backend OpenAPI + frontend TypeScript client)
task api-gen
# Generate backend API only
task api-gen-backend
# Generate frontend TypeScript client only
task api-gen-frontend
```

### Testing and Quality
```bash
# Run frontend tests
task playwright-run
# View test reports
task playwright-report
# Frontend linting
cd front && npm run lint
# Frontend build with type checking
cd front && npm run build:tscheck
```

### Development Tools
```bash
# Open shell in backend container
task shell
# Open PostgreSQL shell
task db-shell
# Build API backend binary locally
task build-api
# Clean environment (removes all data)
task clean
```

### Production
```bash
# Build and run production environment
task prod-up
task prod-down
```

## Architecture Overview

### Core Components
- **Handlers**: HTTP request handlers in `backend/internal/handler/` for each data source (email, telegram, whatsapp, linkedin)
- **Workers**: Background job processing in `backend/internal/worker/` with broker pattern and job scheduling
- **Storage**: Abstraction layer supporting PostgreSQL, S3, and host files in `backend/internal/storages/`
- **Queue**: NATS-based message queuing system for async processing
- **Auth**: OAuth2 integration and Ory Kratos identity management

### Data Sources
Each messaging platform is implemented as a separate data source:
- **Email**: IMAP/SMTP with OAuth2 (Gmail integration)
- **Telegram**: Official Telegram client library with session management
- **WhatsApp**: WhatsApp Business API integration
- **LinkedIn**: LinkedIn messaging API integration

### Database Architecture
- SQL schema in `db/schema.sql` and `db/tg.sql`
- SQLC generates type-safe Go code in `backend/pkg/query/`
- Migrations handled by Atlas
- Supports multiple storage backends (PostgreSQL, S3, local files)

### API Design
- OpenAPI specification in `spec/openapi.yaml`
- Component definitions in `spec/components/`
- Path definitions in `spec/paths/`
- Ogen generates server and client code automatically

### Frontend Architecture
- Component-based React application with TypeScript
- Forms in `front/src/forms/` for each entity type
- Pages in `front/src/pages/` for different views
- Shared authentication in `front/src/shauth/`
- Adobe React Spectrum design system

## Configuration

Environment variables are defined in `backend/internal/config/config.go` with SA_ prefix:
- `SA_BASE_URL`: Base URL for the application
- `SA_DB_URI`: PostgreSQL connection string
- `SA_LOG_LEVEL`: Logging level (debug, info, warn, error)
- `SA_HOST`/`SA_PORT`: Server bind address
- `TG_APP_ID`/`TG_APP_HASH`: Telegram API credentials
- `SA_QUEUE_URL`: NATS server URL

## Development Notes

- Backend uses dependency injection with `github.com/samber/do/v2`
- Database queries are written in SQL and generated to Go with SQLC
- All API endpoints follow OpenAPI specification
- Frontend uses openapi-fetch for type-safe API calls
- Background jobs are processed through NATS queue system
- OAuth2 tokens are managed per data source
- Telegram integration includes session state management
- File uploads support multiple storage backends