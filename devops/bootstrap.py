#!/usr/bin/env python3
"""
MeshPump Bootstrap Script

Initializes the development environment by:
1. Generating secrets and creating .env from template
2. Starting and initializing the database
3. Enrolling the distributed worker
"""

import logging
import os
import re
import secrets
import string
import subprocess
import sys
import time
from dataclasses import dataclass
from pathlib import Path
from typing import Optional

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(message)s",
)
logger = logging.getLogger(__name__)


@dataclass
class EnvSecrets:
    """Secrets that are preserved across bootstrap runs."""

    oidc_callback_secret: str
    worker_id: str
    worker_secret: str


@dataclass
class Config:
    """Runtime configuration parsed from .env file."""

    be_protocol: str = "http"
    be_domain: str = "localtest.me"
    be_api_subdomain: str = "api"
    be_app_subdomain: str = "app"
    be_rpc_subdomain: str = "rpc"
    be_init_admin_email: str = ""
    be_init_admin_password: str = ""


def generate_secret(length: int = 32) -> str:
    """Generate a cryptographically secure random string."""
    alphabet = string.ascii_letters + string.digits
    return "".join(secrets.choice(alphabet) for _ in range(length))


def run_command(
    cmd: list[str],
    *,
    capture_output: bool = False,
    check: bool = True,
    timeout: Optional[int] = None,
) -> subprocess.CompletedProcess:
    """Run a shell command with proper error handling."""
    logger.debug(f"Running: {' '.join(cmd)}")
    return subprocess.run(
        cmd,
        capture_output=capture_output,
        text=True,
        check=check,
        timeout=timeout,
    )


def docker_compose(
    *args: str,
    capture_output: bool = False,
    check: bool = True,
) -> subprocess.CompletedProcess:
    """Run a docker compose command."""
    return run_command(
        ["docker", "compose", *args],
        capture_output=capture_output,
        check=check,
    )


def parse_env_file(env_path: Path) -> dict[str, str]:
    """Parse a .env file into a dictionary."""
    env_vars: dict[str, str] = {}
    if not env_path.exists():
        return env_vars

    with open(env_path) as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            if "=" in line:
                key, _, value = line.partition("=")
                env_vars[key.strip()] = value.strip()
    return env_vars


def extract_preserved_secrets(env_path: Path) -> EnvSecrets:
    """Extract secrets from existing .env file for preservation."""
    env_vars = parse_env_file(env_path)
    return EnvSecrets(
        oidc_callback_secret=env_vars.get("BE_OIDC_CALLBACK_SECRET", ""),
        worker_id=env_vars.get("WORKER_ID", ""),
        worker_secret=env_vars.get("WORKER_SECRET", ""),
    )


def load_config(env_path: Path) -> Config:
    """Load configuration from .env file."""
    env_vars = parse_env_file(env_path)
    return Config(
        be_protocol=env_vars.get("BE_PROTOCOL", "http"),
        be_domain=env_vars.get("BE_DOMAIN", "localtest.me"),
        be_api_subdomain=env_vars.get("BE_API_SUBDOMAIN", "api"),
        be_app_subdomain=env_vars.get("BE_APP_SUBDOMAIN", "app"),
        be_rpc_subdomain=env_vars.get("BE_RPC_SUBDOMAIN", "rpc"),
        be_init_admin_email=env_vars.get("BE_INIT_ADMIN_EMAIL", ""),
        be_init_admin_password=env_vars.get("BE_INIT_ADMIN_PASSWORD", ""),
    )


def update_env_value(env_path: Path, key: str, value: str) -> None:
    """Update a single value in the .env file."""
    content = env_path.read_text()
    pattern = rf"^{re.escape(key)}=.*$"
    replacement = f"{key}={value}"
    updated = re.sub(pattern, replacement, content, flags=re.MULTILINE)
    env_path.write_text(updated)


def create_env_file(
    project_root: Path,
    preserved: EnvSecrets,
) -> None:
    """Create .env file from template with generated/preserved secrets."""
    template_path = project_root / ".env.template"
    env_path = project_root / ".env"

    logger.info("Generating secrets...")

    # Use preserved OIDC callback secret if available, otherwise generate new one
    if preserved.oidc_callback_secret:
        oidc_callback_secret = preserved.oidc_callback_secret
        logger.info("  - Preserving existing BE_OIDC_CALLBACK_SECRET")
    else:
        oidc_callback_secret = generate_secret(32)
        logger.info("  - Generated new BE_OIDC_CALLBACK_SECRET")

    # Generate a client ID placeholder
    oidc_client_id = "meshpump-spa"

    # Read template and substitute placeholders
    logger.info("Creating .env from template...")
    template_content = template_path.read_text()
    env_content = template_content.replace("__OIDC_CALLBACK_SECRET__", oidc_callback_secret)
    env_content = env_content.replace("__OIDC_CLIENT_ID__", oidc_client_id)

    # Remove existing .env if present
    if env_path.exists():
        logger.warning("WARNING: Removing existing .env file and regenerating from template...")
        logger.warning("         Any custom changes will be lost!")
        env_path.unlink()

    env_path.write_text(env_content)

    # Restore preserved worker credentials
    if preserved.worker_id and preserved.worker_secret:
        update_env_value(env_path, "WORKER_ID", preserved.worker_id)
        update_env_value(env_path, "WORKER_SECRET", preserved.worker_secret)
        logger.info("Worker credentials preserved from previous .env")

    logger.info(".env created successfully")


def wait_for_database(max_attempts: int = 30, interval: float = 2.0) -> bool:
    """Wait for the database to be ready."""
    logger.info("Waiting for database to be ready...")

    for attempt in range(1, max_attempts + 1):
        result = docker_compose(
            "exec", "-T", "db",
            "pg_isready", "-U", "shadowapi",
            capture_output=True,
            check=False,
        )
        if result.returncode == 0:
            return True
        logger.info(f"Waiting for database... ({attempt}/{max_attempts})")
        time.sleep(interval)

    return False


def wait_for_backend(max_attempts: int = 30, interval: float = 2.0) -> bool:
    """Wait for backend to be ready."""
    logger.info("Waiting for backend to be ready...")

    for attempt in range(1, max_attempts + 1):
        result = docker_compose(
            "exec", "-T", "backend",
            "shadowapi", "--help",
            capture_output=True,
            check=False,
        )
        if result.returncode == 0:
            return True
        logger.info(f"Waiting for backend... ({attempt}/{max_attempts})")
        time.sleep(interval)

    return False


def check_worker_exists(worker_id: str) -> bool:
    """Check if a worker exists in the database."""
    query = f"SELECT COUNT(*) FROM registered_worker WHERE uuid = '{worker_id}';"
    result = docker_compose(
        "exec", "-T", "db",
        "psql", "-U", "shadowapi", "-d", "shadowapi", "-t", "-c", query,
        capture_output=True,
        check=False,
    )

    if result.returncode != 0:
        return False

    count = result.stdout.strip()
    return count == "1"


def create_enrollment_token() -> Optional[str]:
    """Create an enrollment token for the worker."""
    logger.info("Creating enrollment token...")

    result = docker_compose(
        "exec", "-T", "backend",
        "shadowapi", "create-enrollment-token",
        "--name", "bootstrap-worker",
        "--global",
        "--expires-in", "1",
        capture_output=True,
        check=False,
    )

    if result.returncode != 0:
        logger.warning("Failed to create enrollment token")
        return None

    return result.stdout.strip()


def enroll_worker(token: str) -> tuple[Optional[str], Optional[str]]:
    """Enroll a worker using the enrollment token."""
    logger.info("Enrolling worker...")

    result = docker_compose(
        "run", "--rm", "worker",
        "/bin/worker", "enroll",
        "--server=grpc2nats:9090",
        f"--token={token}",
        "--name=default-worker",
        capture_output=True,
        check=False,
    )

    output = result.stdout + result.stderr
    worker_id = None
    worker_secret = None

    for line in output.splitlines():
        if "Worker ID:" in line:
            parts = line.split()
            if len(parts) >= 3:
                worker_id = parts[2]
        elif "Worker Secret:" in line:
            parts = line.split()
            if len(parts) >= 3:
                worker_secret = parts[2]

    if not worker_id or not worker_secret:
        logger.warning("Failed to enroll worker")
        logger.warning(f"Enrollment output:\n{output}")
        return None, None

    return worker_id, worker_secret


def setup_distributed_worker(project_root: Path) -> Optional[str]:
    """Set up the distributed worker (idempotent)."""
    logger.info("Setting up distributed worker...")

    env_path = project_root / ".env"
    env_vars = parse_env_file(env_path)
    existing_worker_id = env_vars.get("WORKER_ID", "")

    # If we have credentials, verify the worker exists in the database
    if existing_worker_id:
        logger.info(f"Checking if worker {existing_worker_id} exists in database...")
        if check_worker_exists(existing_worker_id):
            logger.info(f"Worker already enrolled: {existing_worker_id}")
            return existing_worker_id

        logger.info(f"Worker {existing_worker_id} not found in database (stale credentials)")
        logger.info("Clearing stale credentials and re-enrolling...")
        update_env_value(env_path, "WORKER_ID", "")
        update_env_value(env_path, "WORKER_SECRET", "")

    # Wait for backend
    if not wait_for_backend():
        logger.error("Backend did not become ready in time")
        return None

    # Create enrollment token
    token = create_enrollment_token()
    if not token:
        logger.warning("Worker will not be enrolled. You can manually enroll the worker later.")
        return None

    logger.info("Enrollment token created")

    # Enroll worker
    worker_id, worker_secret = enroll_worker(token)
    if not worker_id or not worker_secret:
        return None

    # Update .env with worker credentials
    update_env_value(env_path, "WORKER_ID", worker_id)
    update_env_value(env_path, "WORKER_SECRET", worker_secret)
    logger.info(f"Worker enrolled: {worker_id}")

    return worker_id


def print_completion_message(config: Config, worker_id: Optional[str]) -> None:
    """Print the bootstrap completion message."""
    proto = config.be_protocol
    domain = config.be_domain
    api_sub = config.be_api_subdomain
    rpc_sub = config.be_rpc_subdomain

    logger.info("")
    logger.info("=== Bootstrap Complete ===")
    logger.info("")
    logger.info("Services:")
    logger.info(f"  - Frontend (SPA):  {proto}://{domain}")
    logger.info(f"  - API:             {proto}://{api_sub}.{domain}")
    logger.info(f"  - gRPC (workers):  {proto}://{rpc_sub}.{domain}:9090")
    logger.info("")
    logger.info("Workspaces:")
    logger.info(f"  - Internal: {proto}://{domain}/w/internal")
    logger.info(f"  - Demo:     {proto}://{domain}/w/demo")
    logger.info("")
    logger.info(f"Test login:       {config.be_init_admin_email} / {config.be_init_admin_password}")
    if worker_id:
        logger.info(f"Worker ID:        {worker_id}")
    logger.info("")
    logger.info("The admin user has super_admin role and owns 'internal' and 'demo' workspaces.")


def main() -> int:
    """Main bootstrap function."""
    # Determine project root (parent of devops/)
    script_dir = Path(__file__).parent.resolve()
    project_root = script_dir.parent
    os.chdir(project_root)

    logger.info("=== MeshPump Bootstrap ===")

    # Step 1: Preserve existing secrets and create .env
    env_path = project_root / ".env"
    preserved = extract_preserved_secrets(env_path)
    create_env_file(project_root, preserved)

    # Step 2: Start database
    logger.info("Starting database...")
    docker_compose("up", "-d", "db")
    time.sleep(5)  # Initial wait

    if not wait_for_database():
        logger.error("Database did not become ready in time")
        return 1

    # Step 3: Run database migrations
    logger.info("Running database migrations...")
    run_command(["make", "sync-db"])

    # Step 4: Start all services
    logger.info("Starting all services...")
    docker_compose("up", "-d")

    # Step 5: Enroll distributed worker
    worker_id = setup_distributed_worker(project_root)

    # Reload config to get test credentials
    config = load_config(env_path)

    # Step 6: Recreate services to pick up latest .env
    logger.info("")
    logger.info("Recreating services with updated configuration...")
    docker_compose("up", "-d", "--force-recreate")
    logger.info("All services recreated.")

    # Print completion message
    print_completion_message(config, worker_id)

    return 0


if __name__ == "__main__":
    sys.exit(main())
