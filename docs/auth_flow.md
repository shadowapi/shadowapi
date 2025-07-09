# Authentication Flow

ShadowAPI supports multiple authentication scenarios that combine API only access, classic email/password login and ZITADEL based OAuth2 login.  The following diagram describes how each flow interacts with the backend and how cookies are used to maintain a session.

## API only

Machine‑to‑machine requests can authenticate by sending a `Bearer` token in the `Authorization` header.  The token value must match `SA_AUTH_BEARER_TOKEN` configured in `config.yaml`.  No cookies are created.  The request is processed directly by the API server.

## Email/Password login

1. The front‑end posts the user's email and password to `/login`.
2. If credentials are valid, the backend stores a session in memory and responds with cookie `sa_session`.
3. Subsequent API calls include this cookie to prove authentication.
4. Calling `/logout` removes the session and clears the cookie.

## ZITADEL login

1. The front‑end navigates the browser to `/login/zitadel`.
2. The backend generates a PKCE code challenge and redirects the user to ZITADEL.
3. After successful login ZITADEL redirects back to `/auth/callback` with an authorization code.
4. The backend exchanges the code for tokens, stores the local session and sets two cookies:
   - `sa_session` – local session identifier.
   - `zitadel_access_token` – raw access token from ZITADEL allowing the front‑end to detect disabled users.
5. Visiting `/logout` initiates the proper logout flow and clears both cookies.

## Cookies

- `sa_session` – established after successful email/password login or ZITADEL login for enabled users.
- `zitadel_access_token` – forwarded after OAuth2 login regardless of the user's status.  It is removed during logout.

## Front‑end flow

The React front‑end probes `/api/v1/session` to check authentication state.  When `sa_session` is present the API returns `{"active": true}`; otherwise it returns `{"active": false}` with a reason.  The login page uses this information to either redirect to the dashboard, display an error or show a disabled user message when only `zitadel_access_token` exists.

