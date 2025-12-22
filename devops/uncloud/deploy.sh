#!/bin/bash
# MeshPump Uncloud Deployment Script
# Handles database migrations with approval and service deployment
set -e

# =============================================================================
# Configuration
# =============================================================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# =============================================================================
# Helper Functions
# =============================================================================
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# =============================================================================
# Step 1: Validate Prerequisites
# =============================================================================
validate_prerequisites() {
    log_info "Validating prerequisites..."

    # Check uc CLI
    if ! command -v uc &> /dev/null; then
        log_error "uc (uncloud CLI) not found. Install from https://uncloud.run"
        exit 1
    fi

    # Check .env file exists
    if [ ! -f "$SCRIPT_DIR/.env" ]; then
        log_error ".env file not found at $SCRIPT_DIR/.env"
        log_error "Copy .env.example to .env and configure it before deploying"
        exit 1
    fi

    # Check schema files exist
    if [ ! -f "$PROJECT_ROOT/db/schema.sql" ]; then
        log_error "Schema file not found: $PROJECT_ROOT/db/schema.sql"
        exit 1
    fi

    if [ ! -f "$PROJECT_ROOT/db/tg.sql" ]; then
        log_error "Schema file not found: $PROJECT_ROOT/db/tg.sql"
        exit 1
    fi

    log_success "Prerequisites validated"
}

# =============================================================================
# Step 2: Load Environment Variables
# =============================================================================
load_environment() {
    log_info "Loading environment configuration..."

    set -a
    source "$SCRIPT_DIR/.env"
    set +a

    # Validate required variables
    if [ -z "$BE_DB_URI" ]; then
        log_error "BE_DB_URI not set in .env"
        exit 1
    fi

    log_success "Environment loaded"
}

# =============================================================================
# Step 3: Check and Show Pending Migrations (Remote via Uncloud)
# =============================================================================
check_migrations() {
    log_info "Checking for pending database migrations (remote)..."

    # Test database connection first
    log_info "Testing database connection..."
    CONNECTION_TEST=$(uc exec mp-migrate sh -c 'atlas schema inspect --url "$DB_URI" 2>&1 | head -5') || {
        if echo "$CONNECTION_TEST" | grep -qi "connection\|connect\|refused\|timeout\|dial"; then
            log_error "Cannot connect to database. Please check BE_DB_URI in .env"
            log_error "Details: $CONNECTION_TEST"
            exit 1
        fi
        log_error "Failed to inspect database schema"
        log_error "$CONNECTION_TEST"
        exit 1
    }
    log_success "Database connection successful"

    # Note: Atlas requires a clean dev database to compute diffs
    # Since we don't have one in production, we skip the preview
    # The schema will be applied declaratively - Atlas only applies necessary changes
    echo ""
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}     DATABASE MIGRATION${NC}"
    echo -e "${YELLOW}========================================${NC}"
    echo ""
    echo "Atlas will synchronize the database schema with:"
    echo "  - db/schema.sql"
    echo "  - db/tg.sql"
    echo ""
    echo "Atlas uses declarative migrations - only necessary changes"
    echo "will be applied to match the desired schema state."
    echo ""
    echo -e "${YELLOW}========================================${NC}"
    echo ""

    return 0  # Always return 0 to prompt for confirmation
}

# =============================================================================
# Step 4: Prompt for Migration Approval
# =============================================================================
prompt_migration_approval() {
    # Check if running in non-interactive mode
    if [ "$AUTO_APPROVE" = "true" ]; then
        log_warning "Auto-approve enabled - proceeding with migrations"
        return 0
    fi

    echo -e "${RED}WARNING: These migrations will be applied to your PRODUCTION database!${NC}"
    echo ""
    read -p "Do you want to apply these migrations? [y/N]: " -r REPLY
    echo ""

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_warning "Migration aborted by user"
        exit 0
    fi

    log_info "Migration approved by user"
}

# =============================================================================
# Step 5: Apply Migrations (Remote via Uncloud)
# =============================================================================
apply_migrations() {
    log_info "Applying database migrations (remote)..."

    # Apply migrations remotely via uc exec on mp-migrate service
    # Schema files are already in the container at /schema/combined.sql
    # Use sh -c to properly expand $DB_URI inside the container
    uc exec mp-migrate sh -c 'atlas schema apply \
        --url "$DB_URI" \
        --to file:///schema/combined.sql \
        --auto-approve' || {
        log_error "Failed to apply migrations"
        exit 1
    }

    log_success "Database migrations applied successfully"
}

# =============================================================================
# Step 6: Deploy to Uncloud
# =============================================================================
deploy_services() {
    log_info "Deploying services to Uncloud..."

    cd "$SCRIPT_DIR"

    # Deploy using uc (always --yes since user already ran make deploy)
    uc deploy -f compose.yaml --yes || {
        log_error "Deployment failed"
        exit 1
    }

    log_success "Services deployed successfully"
}

# =============================================================================
# Step 7: Verify Deployment
# =============================================================================
verify_deployment() {
    log_info "Verifying deployment..."

    # Wait a moment for services to start
    sleep 5

    # Check service status
    echo ""
    echo "Service Status:"
    uc ps
    echo ""

    # Try to reach the health endpoints
    log_info "Checking service health..."

    # Check API health
    if curl -sf "https://api.meshpump.com/api/v1/health" > /dev/null 2>&1; then
        log_success "API is healthy"
    else
        log_warning "API health check failed (may still be starting)"
    fi

    # Check OIDC discovery
    if curl -sf "https://oidc.meshpump.com/.well-known/openid-configuration" > /dev/null 2>&1; then
        log_success "OIDC is healthy"
    else
        log_warning "OIDC health check failed (may still be starting)"
    fi

    log_success "Deployment verification complete"
}

# =============================================================================
# Main Execution
# =============================================================================
main() {
    echo ""
    echo -e "${BLUE}============================================${NC}"
    echo -e "${BLUE}    MeshPump Uncloud Deployment${NC}"
    echo -e "${BLUE}============================================${NC}"
    echo ""

    # Parse arguments
    SKIP_MIGRATIONS=false
    AUTO_APPROVE=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-migrations)
                SKIP_MIGRATIONS=true
                shift
                ;;
            --yes|-y)
                AUTO_APPROVE=true
                shift
                ;;
            --help|-h)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --skip-migrations    Skip database migrations"
                echo "  --yes, -y           Auto-approve all prompts (for CI/CD)"
                echo "  --help, -h          Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    # Step 1: Validate prerequisites
    validate_prerequisites

    # Step 2: Load environment
    load_environment

    # Step 3: Deploy services first (mp-migrate needs to be running for exec)
    deploy_services

    # Step 4-6: Handle migrations (unless skipped)
    if [ "$SKIP_MIGRATIONS" = false ]; then
        if check_migrations; then
            # Migrations are pending
            prompt_migration_approval
            apply_migrations
        fi
    else
        log_warning "Skipping database migrations as requested"
    fi

    # Step 7: Verify deployment
    verify_deployment

    echo ""
    echo -e "${GREEN}============================================${NC}"
    echo -e "${GREEN}    Deployment Complete!${NC}"
    echo -e "${GREEN}============================================${NC}"
    echo ""
    echo "Services:"
    echo "  - Frontend: https://app.meshpump.com"
    echo "  - SSR:      https://meshpump.com"
    echo "  - API:      https://api.meshpump.com"
    echo "  - OIDC:     https://oidc.meshpump.com"
    echo ""
}

# Run main with all arguments
main "$@"
