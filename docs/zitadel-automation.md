# Zitadel Service User Automation

This document describes the automated setup for Zitadel service users in ShadowAPI.

## 🚀 Quick Start

### 1. Automatic Setup (Recommended)

Set up Zitadel service user with a single command:

```bash
# Setup and enable Zitadel authentication
task enable-zitadel-auth

# Or just setup without enabling
task setup-zitadel
```

### 2. Manual Steps

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
| `task setup-zitadel` | Create service user and generate keys |
| `task enable-zitadel-auth` | Setup and enable Zitadel authentication |
| `task disable-zitadel-auth` | Switch back to database authentication |
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

## 🔄 Switching Authentication Methods

### Enable Zitadel

```bash
task enable-zitadel-auth
```

This will:
1. Run service user setup if needed
2. Add `SA_AUTH_USER_MANAGER=zitadel` to `.env`
3. Extract and set `SA_AUTH_ZITADEL_SERVICE_USER_ID`

### Disable Zitadel (use database)

```bash
task disable-zitadel-auth
```

This will:
1. Remove Zitadel settings from `.env`
2. Set `SA_AUTH_USER_MANAGER=db`

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
# Start with database auth (default)
task dev-up

# Switch to Zitadel when needed
task enable-zitadel-auth
task dev-down && task dev-up
```

### Testing Both Methods

```bash
# Test database auth
task disable-zitadel-auth
# ... test user operations ...

# Test Zitadel auth
task enable-zitadel-auth
# ... test user operations ...
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
| `SA_AUTH_USER_MANAGER` | User manager type (`db` or `zitadel`) | `db` |
| `SA_AUTH_ZITADEL_MANAGEMENT_URL` | Zitadel Management API URL | `http://zitadel:8080` |
| `SA_AUTH_ZITADEL_SERVICE_USER_ID` | Service user ID from key file | - |
| `SA_AUTH_ZITADEL_SERVICE_USER_KEY_PATH` | Path to private key file | `/secrets/zitadel-service-key.json` |