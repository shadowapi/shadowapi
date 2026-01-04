#!/usr/bin/env python3
"""
MeshPump Bootstrap Script

Initializes the development environment by:
1. Generating secrets and creating .env from template
2. Generating hydra.yaml from template
3. Starting and initializing the database
4. Configuring OAuth2 client in Hydra
5. Enrolling the distributed worker
"""

import json
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

    hydra_secrets_system: str
    oidc_pairwise_salt: str
    worker_id: str
    worker_secret: str


@dataclass
class Config:
    """Runtime configuration parsed from .env file."""

    be_protocol: str = "http"
    be_domain: str = "localtest.me"
    be_api_subdomain: str = "api"
    be_app_subdomain: str = "app"
    be_oidc_subdomain: str = "oidc"
    be_ssr_subdomain: str = "www"
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
        hydra_secrets_system=env_vars.get("HYDRA_SECRETS_SYSTEM", ""),
        oidc_pairwise_salt=env_vars.get("OIDC_PAIRWISE_SALT", ""),
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
        be_oidc_subdomain=env_vars.get("BE_OIDC_SUBDOMAIN", "oidc"),
        be_ssr_subdomain=env_vars.get("BE_SSR_SUBDOMAIN", "www"),
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
) -> tuple[str, str]:
    """Create .env file from template with generated/preserved secrets."""
    template_path = project_root / ".env.template"
    env_path = project_root / ".env"

    logger.info("Generating secrets...")

    # Use preserved secrets if available, otherwise generate new ones
    if preserved.hydra_secrets_system:
        hydra_secret = preserved.hydra_secrets_system
        logger.info("  - Preserving existing HYDRA_SECRETS_SYSTEM")
    else:
        hydra_secret = generate_secret(32)
        logger.info("  - Generated new HYDRA_SECRETS_SYSTEM")

    if preserved.oidc_pairwise_salt:
        oidc_salt = preserved.oidc_pairwise_salt
        logger.info("  - Preserving existing OIDC_PAIRWISE_SALT")
    else:
        oidc_salt = generate_secret(16)
        logger.info("  - Generated new OIDC_PAIRWISE_SALT")

    # Read template and substitute placeholders
    logger.info("Creating .env from template...")
    template_content = template_path.read_text()
    env_content = template_content.replace("__HYDRA_SECRETS_SYSTEM__", hydra_secret)
    env_content = env_content.replace("__OIDC_PAIRWISE_SALT__", oidc_salt)
    env_content = env_content.replace("__OAUTH2_CLIENT_ID__", "pending-creation")

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
    return hydra_secret, oidc_salt


def generate_hydra_config(project_root: Path) -> None:
    """Generate hydra.yaml from template using environment variables."""
    logger.info("Generating hydra.yaml from template...")

    template_path = project_root / "devops" / "ory" / "hydra" / "hydra.template.yaml"
    output_path = project_root / "devops" / "ory" / "hydra" / "hydra.yaml"
    env_path = project_root / ".env"

    # Load environment variables from .env
    env_vars = parse_env_file(env_path)

    # Read template
    template_content = template_path.read_text()

    # Substitute ${VAR} placeholders
    def replace_var(match: re.Match) -> str:
        var_name = match.group(1)
        return env_vars.get(var_name, "")

    output_content = re.sub(r"\$\{(\w+)\}", replace_var, template_content)
    output_path.write_text(output_content)

    logger.info("hydra.yaml created successfully")


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


def wait_for_hydra(max_attempts: int = 30, interval: float = 2.0) -> bool:
    """Wait for Hydra to be ready."""
    logger.info("Waiting for Hydra...")

    for attempt in range(1, max_attempts + 1):
        result = docker_compose(
            "exec", "-T", "hydra",
            "hydra", "version",
            capture_output=True,
            check=False,
        )
        if result.returncode == 0:
            return True
        logger.info(f"Waiting for Hydra... ({attempt}/{max_attempts})")
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


def get_existing_oauth2_client(client_name: str) -> Optional[str]:
    """Check if an OAuth2 client with the given name already exists."""
    result = docker_compose(
        "exec", "-T", "hydra",
        "hydra", "list", "oauth2-clients",
        "--endpoint", "http://localhost:4445",
        "--format", "json",
        capture_output=True,
        check=False,
    )

    if result.returncode != 0:
        return None

    try:
        data = json.loads(result.stdout)
        items = data.get("items", [])
        for item in items:
            if item.get("client_name") == client_name:
                return item.get("client_id")
    except (json.JSONDecodeError, KeyError):
        pass

    return None


def create_oauth2_client(redirect_uri: str, client_name: str) -> Optional[str]:
    """Create a new OAuth2 client in Hydra."""
    result = docker_compose(
        "exec", "-T", "hydra",
        "hydra", "create", "oauth2-client",
        "--endpoint", "http://localhost:4445",
        "--name", client_name,
        "--grant-type", "authorization_code,refresh_token",
        "--response-type", "code",
        "--scope", "openid,offline_access,profile,email",
        "--redirect-uri", redirect_uri,
        "--token-endpoint-auth-method", "none",
        "--format", "json",
        capture_output=True,
        check=False,
    )

    if result.returncode != 0:
        logger.error(f"Failed to create OAuth2 client: {result.stderr}")
        return None

    try:
        data = json.loads(result.stdout)
        return data.get("client_id")
    except json.JSONDecodeError:
        logger.error(f"Failed to parse OAuth2 client response: {result.stdout}")
        return None


def setup_oauth2_client(project_root: Path, config: Config) -> Optional[str]:
    """Create or retrieve existing OAuth2 client."""
    logger.info("Creating OAuth2 client...")

    redirect_uri = (
        f"{config.be_protocol}://{config.be_api_subdomain}.{config.be_domain}"
        f"/api/v1/auth/oauth2/callback"
    )
    client_name = "ShadowAPI SPA"

    # Check if client already exists
    existing_client_id = get_existing_oauth2_client(client_name)
    if existing_client_id:
        logger.info(f"OAuth2 client already exists: {existing_client_id}")
        client_id = existing_client_id
    else:
        client_id = create_oauth2_client(redirect_uri, client_name)
        if client_id:
            logger.info(f"OAuth2 client created: {client_id}")
        else:
            logger.error("Failed to create OAuth2 client")
            return None

    # Update .env with client ID
    env_path = project_root / ".env"
    update_env_value(env_path, "BE_OAUTH2_SPA_CLIENT_ID", client_id)
    logger.info(f"Updated .env with client ID: {client_id}")

    return client_id


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


def print_completion_message(config: Config, client_id: str, worker_id: Optional[str]) -> None:
    """Print the bootstrap completion message."""
    proto = config.be_protocol
    domain = config.be_domain
    api_sub = config.be_api_subdomain
    oidc_sub = config.be_oidc_subdomain
    ssr_sub = config.be_ssr_subdomain
    rpc_sub = config.be_rpc_subdomain

    logger.info("")
    logger.info("=== Bootstrap Complete ===")
    logger.info("")
    logger.info("Services:")
    logger.info(f"  - Frontend (SPA):  {proto}://{domain}")
    logger.info(f"  - API:             {proto}://{api_sub}.{domain}")
    logger.info(f"  - gRPC (workers):  {proto}://{rpc_sub}.{domain}:9090")
    logger.info(f"  - OIDC:            {proto}://{oidc_sub}.{domain}")
    logger.info(f"  - SSR (www):       {proto}://{ssr_sub}.{domain}")
    logger.info("")
    logger.info("Workspaces:")
    logger.info(f"  - Internal: {proto}://{domain}/w/internal")
    logger.info(f"  - Demo:     {proto}://{domain}/w/demo")
    logger.info("")
    logger.info(f"Test login:       {config.be_init_admin_email} / {config.be_init_admin_password}")
    logger.info(f"OAuth2 Client ID: {client_id}")
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

    # Step 1.5: Generate hydra.yaml
    generate_hydra_config(project_root)

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

    # Step 4: Start Hydra
    logger.info("Starting Hydra...")
    docker_compose("up", "-d", "hydra")

    if not wait_for_hydra():
        logger.error("Hydra did not become ready in time")
        return 1

    # Step 5: Create OAuth2 client
    config = load_config(env_path)
    client_id = setup_oauth2_client(project_root, config)
    if not client_id:
        return 1

    # Step 6: Start all services
    logger.info("Starting all services...")
    docker_compose("up", "-d")

    # Step 7: Enroll distributed worker
    worker_id = setup_distributed_worker(project_root)

    # Reload config to get test credentials
    config = load_config(env_path)

    # Step 8: Recreate services to pick up latest .env
    logger.info("")
    logger.info("Recreating services with updated configuration...")
    docker_compose("up", "-d", "--force-recreate")
    logger.info("All services recreated.")

    # Print completion message
    print_completion_message(config, client_id, worker_id)

    return 0


if __name__ == "__main__":
    sys.exit(main())
