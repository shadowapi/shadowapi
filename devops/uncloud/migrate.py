#!/usr/bin/env python3
"""MeshPump Database Migration Script

Establishes SSH tunnel to remote database and runs Atlas migrations.
"""

import atexit
import shutil
import subprocess
import sys
import time
from pathlib import Path
from urllib.parse import urlparse, urlunparse

# Colors for output
RED = "\033[0;31m"
GREEN = "\033[0;32m"
YELLOW = "\033[1;33m"
BLUE = "\033[0;34m"
NC = "\033[0m"

SCRIPT_DIR = Path(__file__).parent.resolve()
REPO_ROOT = SCRIPT_DIR.parent.parent

# SSH tunnel configuration
SSH_HOST = "devinlab.com"
REMOTE_DB_PORT = 10432
LOCAL_DB_PORT = 15432

# Global reference to tunnel process for cleanup
_tunnel_process = None


def log_info(msg: str) -> None:
    print(f"{BLUE}[INFO]{NC} {msg}")


def log_success(msg: str) -> None:
    print(f"{GREEN}[SUCCESS]{NC} {msg}")


def log_warning(msg: str) -> None:
    print(f"{YELLOW}[WARNING]{NC} {msg}")


def log_error(msg: str) -> None:
    print(f"{RED}[ERROR]{NC} {msg}", file=sys.stderr)


def cleanup_tunnel() -> None:
    """Terminate SSH tunnel process on exit."""
    global _tunnel_process
    if _tunnel_process is not None:
        log_info("Closing SSH tunnel...")
        _tunnel_process.terminate()
        try:
            _tunnel_process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            _tunnel_process.kill()
        _tunnel_process = None


def load_env() -> dict[str, str]:
    """Load environment variables from .env file."""
    env_path = SCRIPT_DIR / ".env"
    if not env_path.exists():
        log_error(f".env file not found at {env_path}")
        sys.exit(1)

    env_vars = {}
    with open(env_path) as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            if "=" in line:
                key, _, value = line.partition("=")
                value = value.strip()
                if (value.startswith('"') and value.endswith('"')) or \
                   (value.startswith("'") and value.endswith("'")):
                    value = value[1:-1]
                env_vars[key.strip()] = value
    return env_vars


def rewrite_db_url(url: str, new_host: str, new_port: int, db_suffix: str = "") -> str:
    """Rewrite database URL with new host/port and optional db name suffix."""
    parsed = urlparse(url)

    # Build new netloc with new host:port
    if parsed.username and parsed.password:
        new_netloc = f"{parsed.username}:{parsed.password}@{new_host}:{new_port}"
    elif parsed.username:
        new_netloc = f"{parsed.username}@{new_host}:{new_port}"
    else:
        new_netloc = f"{new_host}:{new_port}"

    # Modify database name if suffix provided
    new_path = parsed.path
    if db_suffix and new_path:
        new_path = new_path + db_suffix

    return urlunparse((
        parsed.scheme,
        new_netloc,
        new_path,
        parsed.params,
        parsed.query,
        parsed.fragment
    ))


def validate_prerequisites() -> None:
    """Check that required tools are installed."""
    log_info("Validating prerequisites...")

    if not shutil.which("atlas"):
        log_error("atlas not found. Install from https://atlasgo.io/")
        sys.exit(1)

    if not shutil.which("ssh"):
        log_error("ssh not found")
        sys.exit(1)

    log_success("Prerequisites validated")


def start_ssh_tunnel() -> subprocess.Popen:
    """Start SSH tunnel in background."""
    global _tunnel_process

    log_info(f"Starting SSH tunnel (localhost:{LOCAL_DB_PORT} -> {SSH_HOST}:{REMOTE_DB_PORT})...")

    _tunnel_process = subprocess.Popen(
        [
            "ssh", "-N", "-L",
            f"{LOCAL_DB_PORT}:localhost:{REMOTE_DB_PORT}",
            SSH_HOST
        ],
        stdout=subprocess.DEVNULL,
        stderr=subprocess.PIPE
    )

    # Register cleanup handler
    atexit.register(cleanup_tunnel)

    # Wait for tunnel to establish
    time.sleep(2)

    # Check if process is still running
    if _tunnel_process.poll() is not None:
        stderr = _tunnel_process.stderr.read().decode() if _tunnel_process.stderr else ""
        log_error(f"SSH tunnel failed to start: {stderr}")
        sys.exit(1)

    log_success("SSH tunnel established")
    return _tunnel_process


def run_atlas_migration(main_url: str, dev_url: str) -> int:
    """Run Atlas schema apply command."""
    schema_path = REPO_ROOT / "db" / "schema.sql"

    if not schema_path.exists():
        log_error(f"Schema file not found: {schema_path}")
        return 1

    log_info("Running Atlas migration...")
    log_info(f"Schema: {schema_path}")

    result = subprocess.run([
        "atlas", "schema", "apply",
        "--url", main_url,
        "--dev-url", dev_url,
        "--to", f"file://{schema_path}"
    ])

    return result.returncode


def main() -> None:
    print()
    print(f"{BLUE}============================================{NC}")
    print(f"{BLUE}    MeshPump Database Migration{NC}")
    print(f"{BLUE}============================================{NC}")
    print()

    # Step 1: Validate prerequisites
    validate_prerequisites()

    # Step 2: Load environment
    log_info("Loading environment...")
    env = load_env()

    db_uri = env.get("BE_DB_URI")
    if not db_uri:
        log_error("BE_DB_URI not found in .env")
        sys.exit(1)

    log_success("Environment loaded")

    # Step 3: Start SSH tunnel
    start_ssh_tunnel()

    # Step 4: Build rewritten URLs
    main_url = rewrite_db_url(db_uri, "localhost", LOCAL_DB_PORT)
    dev_url = rewrite_db_url(db_uri, "localhost", LOCAL_DB_PORT, "_dev")

    log_info(f"Main DB: {main_url.split('@')[1] if '@' in main_url else main_url}")
    log_info(f"Dev DB:  {dev_url.split('@')[1] if '@' in dev_url else dev_url}")

    # Step 5: Run migration
    exit_code = run_atlas_migration(main_url, dev_url)

    # Cleanup happens via atexit
    if exit_code == 0:
        print()
        log_success("Migration completed successfully")
    else:
        print()
        log_error(f"Migration failed with exit code {exit_code}")

    sys.exit(exit_code)


if __name__ == "__main__":
    main()
