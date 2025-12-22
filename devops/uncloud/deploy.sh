#!/bin/bash
# MeshPump Uncloud Deployment Script
set -e

# =============================================================================
# Configuration
# =============================================================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

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
# Step 3: Deploy to Uncloud
# =============================================================================
deploy_services() {
    log_info "Deploying services to Uncloud..."

    cd "$SCRIPT_DIR"

    uc deploy -f compose.yaml --yes || {
        log_error "Deployment failed"
        exit 1
    }

    log_success "Services deployed successfully"
}

# =============================================================================
# Step 4: Verify Deployment
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
    while [[ $# -gt 0 ]]; do
        case $1 in
            --help|-h)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
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

    # Step 3: Deploy services
    deploy_services

    # Step 4: Verify deployment
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
