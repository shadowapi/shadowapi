#!/usr/bin/env python3
"""MeshPump Uncloud Deployment Script"""

import argparse
import os
import shutil
import subprocess
import sys
import time
import urllib.request
import urllib.error
from pathlib import Path

# Colors for output
RED = "\033[0;31m"
GREEN = "\033[0;32m"
YELLOW = "\033[1;33m"
BLUE = "\033[0;34m"
NC = "\033[0m"  # No Color

SCRIPT_DIR = Path(__file__).parent.resolve()


def log_info(msg: str) -> None:
    print(f"{BLUE}[INFO]{NC} {msg}")


def log_success(msg: str) -> None:
    print(f"{GREEN}[SUCCESS]{NC} {msg}")


def log_warning(msg: str) -> None:
    print(f"{YELLOW}[WARNING]{NC} {msg}")


def log_error(msg: str) -> None:
    print(f"{RED}[ERROR]{NC} {msg}", file=sys.stderr)


def validate_prerequisites() -> None:
    """Step 1: Validate Prerequisites"""
    log_info("Validating prerequisites...")

    # Check uc CLI
    if not shutil.which("uc"):
        log_error("uc (uncloud CLI) not found. Install from https://uncloud.run")
        sys.exit(1)

    # Check sops CLI (required for secrets decryption)
    if not shutil.which("sops"):
        log_error("sops not found. Install with: brew install sops")
        sys.exit(1)

    env_enc_path = SCRIPT_DIR / ".env.enc"
    env_path = SCRIPT_DIR / ".env"

    # Auto-decrypt .env.enc if .env doesn't exist
    if env_enc_path.exists() and not env_path.exists():
        log_info("Decrypting secrets from .env.enc...")
        sops_key_file = os.environ.get(
            "SOPS_AGE_KEY_FILE",
            Path.home() / ".config/sops/age/keys.txt"
        )
        env = os.environ.copy()
        env["SOPS_AGE_KEY_FILE"] = str(sops_key_file)

        try:
            result = subprocess.run(
                [
                    "sops", "--decrypt",
                    "--input-type", "dotenv",
                    "--output-type", "dotenv",
                    str(env_enc_path)
                ],
                capture_output=True,
                text=True,
                env=env
            )
            if result.returncode == 0:
                env_path.write_text(result.stdout)
                log_success("Secrets decrypted")
            else:
                log_error(f"Failed to decrypt .env.enc. Check your age key at {sops_key_file}")
                sys.exit(1)
        except Exception as e:
            log_error(f"Failed to decrypt .env.enc: {e}")
            sys.exit(1)

    # Check .env file exists
    if not env_path.exists():
        log_error(f".env file not found at {env_path}")
        log_error("Either decrypt .env.enc or copy .env.example and configure it")
        sys.exit(1)

    log_success("Prerequisites validated")


def load_environment() -> None:
    """Step 2: Load Environment Variables"""
    log_info("Loading environment configuration...")

    env_path = SCRIPT_DIR / ".env"
    with open(env_path) as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            if "=" in line:
                key, _, value = line.partition("=")
                # Remove surrounding quotes if present
                value = value.strip()
                if (value.startswith('"') and value.endswith('"')) or \
                   (value.startswith("'") and value.endswith("'")):
                    value = value[1:-1]
                os.environ[key.strip()] = value

    # Validate required variables
    if not os.environ.get("BE_DB_URI"):
        log_error("BE_DB_URI not set in .env")
        sys.exit(1)

    log_success("Environment loaded")


def deploy_services() -> None:
    """Step 3: Deploy to Uncloud"""
    log_info("Deploying services to Uncloud...")

    result = subprocess.run(
        ["uc", "deploy", "-f", "compose.yaml", "--yes"],
        cwd=SCRIPT_DIR
    )
    if result.returncode != 0:
        log_error("Deployment failed")
        sys.exit(1)

    log_success("Services deployed successfully")


def check_url(url: str) -> bool:
    """Check if a URL is reachable and returns success status"""
    try:
        req = urllib.request.Request(url, method="GET")
        with urllib.request.urlopen(req, timeout=5) as response:
            return response.status == 200
    except (urllib.error.URLError, urllib.error.HTTPError, TimeoutError):
        return False


def verify_deployment() -> None:
    """Step 4: Verify Deployment"""
    log_info("Verifying deployment...")

    # Wait a moment for services to start
    time.sleep(5)

    # Check service status
    print()
    print("Service Status:")
    subprocess.run(["uc", "ps"], cwd=SCRIPT_DIR)
    print()

    # Try to reach the health endpoints
    log_info("Checking service health...")

    # Check API health
    if check_url("https://api.meshpump.com/api/v1/health"):
        log_success("API is healthy")
    else:
        log_warning("API health check failed (may still be starting)")

    # Check OIDC discovery
    if check_url("https://oidc.meshpump.com/.well-known/openid-configuration"):
        log_success("OIDC is healthy")
    else:
        log_warning("OIDC health check failed (may still be starting)")

    log_success("Deployment verification complete")


def main() -> None:
    print()
    print(f"{BLUE}============================================{NC}")
    print(f"{BLUE}    MeshPump Uncloud Deployment{NC}")
    print(f"{BLUE}============================================{NC}")
    print()

    parser = argparse.ArgumentParser(description="MeshPump Uncloud Deployment Script")
    parser.parse_args()

    # Step 1: Validate prerequisites
    validate_prerequisites()

    # Step 2: Load environment
    load_environment()

    # Step 3: Deploy services
    deploy_services()

    # Step 4: Verify deployment
    verify_deployment()

    print()
    print(f"{GREEN}============================================{NC}")
    print(f"{GREEN}    Deployment Complete!{NC}")
    print(f"{GREEN}============================================{NC}")
    print()
    print("Services:")
    print("  - Frontend: https://app.meshpump.com")
    print("  - SSR:      https://meshpump.com")
    print("  - API:      https://api.meshpump.com")
    print("  - OIDC:     https://oidc.meshpump.com")
    print()


if __name__ == "__main__":
    main()
