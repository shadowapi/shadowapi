# ZITADEL Configuration

The ZITADEL identity provider is expected to run on `https://reactima.com`.
Create a project and web application with the following settings:

- **Redirect URIs**: `https://reactima.com/login/callback`
- **Post logout redirect URI**: `https://reactima.com/logout/callback`

All URIs must start with `https://`. Using `http://` is only possible when the ZITADEL instance is in development mode.
