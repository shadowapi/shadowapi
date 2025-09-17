# Zitadel Service User Automation

This document describes the automated setup for Zitadel service users in ShadowAPI.

## 🚀 Quick Start

Set up Zitadel service user with a single command:

```bash
# Setup Zitadel authentication (default)
task setup-zitadel
```

This will:
1. Create service user and generate keys
2. Configure environment variables
3. Enable Zitadel authentication

### Manual Steps

If you prefer manual setup:

```bash
# 1. Start Zitadel
task dev-up

# 2. Run setup script directly
./devops/setup-zitadel-service-user.sh

# 3. Configure environment
export SA_AUTH_USER_MANAGER=zitadel
export SA_AUTH_ZITADEL_SERVICE_USER_ID=<user-id-from-key-file>

# 4. Restart backend
task dev-down && task dev-up
```

## 📋 Available Commands

| Command | Description |
|---------|-------------|
| `task setup-zitadel` | Setup service user and enable Zitadel authentication |
| `task clean-zitadel-keys` | Remove generated keys |

## 🔧 How It Works

### 1. Docker Automation

The `zitadel-setup` service in `compose.yaml`:
- Waits for Zitadel to be healthy
- Runs the setup script automatically
- Creates service user and keys
- Mounts keys to `./secrets/` directory

### 2. Setup Script

`devops/setup-zitadel-service-user.sh`:
- Gets admin access token
- Creates machine user via Management API
- Generates private key for JWT authentication
- Grants necessary permissions (ORG_OWNER)
- Saves key file to `./secrets/zitadel-service-key.json`

### 3. Backend Integration

The backend automatically:
- Loads the private key on startup
- Creates JWT tokens for authentication
- Caches access tokens with refresh logic
- Handles all Management API requests

## 📁 Generated Files

```
./secrets/
└── zitadel-service-key.json    # Service user private key (DO NOT COMMIT!)
```

## 🔄 Authentication Method

ShadowAPI uses Zitadel authentication by default. The system is configured to use the Zitadel Management API for all user operations including registration, login, and user management.

## 🔐 Security Notes

### Key Management

- Keys are stored in `./secrets/` (gitignored)
- Files have 600 permissions (owner read-only)
- Keys are generated fresh for each environment

### Permissions

The service user gets these roles:
- `ORG_OWNER`: Full organization access
- Access to Management API endpoints

### Best Practices

1. **Never commit keys** - `./secrets/` is in `.gitignore`
2. **Rotate keys regularly** - run `task clean-zitadel-keys && task setup-zitadel`
3. **Use different keys per environment** - dev/staging/prod
4. **Monitor service user activity** - check Zitadel audit logs

## 🛠️ Troubleshooting

### Service User Creation Fails

```bash
# Check Zitadel health
curl http://auth.localtest.me/debug/ready

# Check admin credentials
docker logs sa-zitadel

# Recreate service user
task clean-zitadel-keys && task setup-zitadel
```

### JWT Authentication Fails

```bash
# Check key file format
cat ./secrets/zitadel-service-key.json | jq .

# Check backend logs for JWT errors
docker logs sa-backend | grep -i jwt

# Verify environment variables
docker exec sa-backend env | grep ZITADEL
```

### Permission Errors

The service user needs these permissions in Zitadel Console:
- Organization → Authorization → Grant `ORG_OWNER` role
- Project → Authorization → Grant `PROJECT_OWNER` role (if using projects)

## 🔄 Development Workflow

### Daily Development

```bash
# Setup Zitadel authentication (first time only)
task setup-zitadel

# Start development environment
task dev-up
```

### Clean Environment

```bash
task clean
task clean-zitadel-keys
task dev-up
```

## 🏗️ Architecture

```
┌─────────────────┐    JWT     ┌──────────────────┐
│   ShadowAPI     │ ────────→  │     Zitadel      │
│    Backend      │            │  Management API  │
└─────────────────┘            └──────────────────┘
         │                              │
         │ Private Key                  │ User Data
         ▼                              ▼
┌─────────────────┐            ┌──────────────────┐
│ ./secrets/      │            │   Zitadel DB     │
│ key.json        │            │   (Users)        │
└─────────────────┘            └──────────────────┘
```

## 📝 Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SA_AUTH_USER_MANAGER` | User manager type (always `zitadel`) | `zitadel` |
| `SA_AUTH_ZITADEL_MANAGEMENT_URL` | Zitadel Management API URL | `http://zitadel:8080` |
| `SA_AUTH_ZITADEL_SERVICE_USER_ID` | Service user ID from key file | - |
| `SA_AUTH_ZITADEL_SERVICE_USER_KEY_PATH` | Path to private key file | `/secrets/zitadel-service-key.json` |