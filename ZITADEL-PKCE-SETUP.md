# ZITADEL PKCE Setup Documentation

This document explains how the ZITADEL PKCE application setup works in the ShadowAPI project, enabling secure authentication without client secrets.

## 🎯 Overview

The ShadowAPI project has been configured to automatically create a PKCE (Proof Key for Code Exchange) application in ZITADEL during the first `docker compose up -d`. This replaces session-based authentication with the more secure OAuth2 PKCE flow.

## 📁 Files Overview

### Configuration Files
- `devops/zitadel-init-steps.yaml` - ZITADEL instance initialization (org, users, machine keys)
- `devops/setup-zitadel-pkce.sh` - Automated PKCE application creation script
- `zitadel-pkce-config.json` - Generated PKCE application configuration
- `secrets/.zitadel-setup-complete` - First-run completion marker

### Docker Compose Integration
- `compose.yaml` includes `zitadel-setup` service that runs automatically after ZITADEL is healthy

## 🚀 How It Works

### 1. ZITADEL Instance Setup
When you run `docker compose up -d`, ZITADEL initializes with:
- **Instance**: ShadowAPI
- **Organization**: ShadowAPI
- **Admin User**: admin@example.com / Admin123!
- **Service User**: shadowapi-admin-service (for API access)

### 2. Automatic PKCE Application Creation
The `zitadel-setup` container automatically:
1. Waits for ZITADEL to be healthy
2. Authenticates using the machine key (JWT flow)
3. Creates a "ShadowAPI" project
4. Creates a "ShadowAPI Frontend" PKCE application
5. Saves configuration to `zitadel-pkce-config.json`
6. Marks setup as complete to prevent re-running

### 3. Generated Configuration
After successful setup, `zitadel-pkce-config.json` contains:

```json
{
    "project_id": "339013429979250696",
    "project_name": "ShadowAPI",
    "app_id": "339013484052217864",
    "app_name": "ShadowAPI Frontend",
    "client_id": "339013484052283400",
    "auth_url": "http://auth.localtest.me/oauth/v2/authorize",
    "token_url": "http://auth.localtest.me/oauth/v2/token",
    "userinfo_url": "http://auth.localtest.me/oidc/v1/userinfo",
    "redirect_uris": [
        "http://localtest.me/auth/callback",
        "http://localhost:5173/auth/callback"
    ]
}
```

## 🔧 PKCE Application Settings

The auto-created application has these security settings:

| Setting | Value | Description |
|---------|-------|-------------|
| **Auth Method** | `NONE` | No client secret required (PKCE) |
| **Response Types** | `CODE` | Authorization code flow |
| **Grant Types** | `AUTHORIZATION_CODE`, `REFRESH_TOKEN` | Standard PKCE grants |
| **App Type** | `WEB` | Web application |
| **Dev Mode** | `true` | Allows localhost redirects |

### Redirect URIs
- Production: `http://localtest.me/auth/callback`
- Development: `http://localhost:5173/auth/callback`

### Post-logout URIs
- Production: `http://localtest.me/`
- Development: `http://localhost:5173/`

## 🔄 Usage

### First Time Setup
```bash
docker compose up -d
```

The setup happens automatically. Monitor with:
```bash
docker logs sa-zitadel-setup -f
```

### Subsequent Runs
The setup is skipped if `secrets/.zitadel-setup-complete` exists and configuration is present.

### Manual Re-setup
To force a fresh setup:
```bash
rm -f zitadel-pkce-config.json secrets/.zitadel-setup-complete
docker compose restart zitadel-setup
```

## 🖥️ Frontend Integration

Use the generated `client_id` from `zitadel-pkce-config.json` in your frontend:

```typescript
const oidcConfig = {
  clientId: "339013484052283400", // From zitadel-pkce-config.json
  authUrl: "http://auth.localtest.me/oauth/v2/authorize",
  tokenUrl: "http://auth.localtest.me/oauth/v2/token",
  redirectUri: window.location.origin + "/auth/callback"
};
```

## 🔍 Verification

### Check Database
```bash
# Verify project exists
docker exec sa-db psql -U zitadel -d zitadel -c \
  "SELECT id, name FROM projections.projects4 WHERE name = 'ShadowAPI';"

# Verify application exists
docker exec sa-db psql -U zitadel -d zitadel -c \
  "SELECT id, name FROM projections.apps7 WHERE name = 'ShadowAPI Frontend';"
```

### Access ZITADEL Console
- URL: http://auth.localtest.me/ui/console
- Login: admin@example.com / Admin123!

## 🛠️ Troubleshooting

### Setup Container Fails
```bash
# Check logs
docker logs sa-zitadel-setup

# Common issues:
# 1. ZITADEL not ready - increase wait time in script
# 2. Network connectivity - verify containers can reach each other
# 3. Machine key missing - check if zitadel-init-steps.yaml is mounted
```

### Missing Configuration
```bash
# Check if setup completed
ls -la secrets/.zitadel-setup-complete
ls -la zitadel-pkce-config.json

# If missing, setup didn't complete successfully
docker logs sa-zitadel-setup
```

### Manual Application Creation
If automation fails, create manually via console:
1. Go to http://auth.localtest.me/ui/console
2. Login as admin@example.com
3. Create project "ShadowAPI"
4. Add Web application with PKCE settings above

## 📋 Key Benefits

✅ **Security**: PKCE flow eliminates client secret management
✅ **Automation**: Zero manual setup required
✅ **Consistency**: Same configuration every time
✅ **Development**: Works locally and in production
✅ **Standards**: Follows OAuth2/OIDC best practices

## 🔗 Related Files

- `backend/internal/auth/` - Backend OIDC integration
- `front/src/shauth/` - Frontend auth implementation
- `devops/zitadel-config.yaml` - ZITADEL runtime configuration
- `compose.yaml` - Docker services orchestration