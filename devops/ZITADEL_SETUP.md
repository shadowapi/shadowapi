# Zitadel Application Setup

Zitadel's `start-from-init` creates the instance, org, and service account automatically, but **not projects or applications**.

## Quick Setup

Run this script after starting the stack:

```bash
./devops/create-zitadel-app.sh
```

This creates the ShadowAPI project and OIDC application, then outputs the Client ID to update in your `.env` files.

## Manual Setup (Console UI)

1. Go to `http://auth.localtest.me/ui/console`
2. Login: `admin@example.com` / `Admin123!`
3. Create project "ShadowAPI"
4. Add OIDC app "ShadowAPI Frontend":
   - Type: User Agent
   - Auth Method: NONE (PKCE)
   - Redirect URIs: `http://localtest.me/login`, `http://localhost:5173/login`
   - Dev Mode: Enabled
5. Copy Client ID to `.env` files

## Current IDs

- Client ID: `340429438871235336`
- Admin: `admin@example.com` / `Admin123!`
