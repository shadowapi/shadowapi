/**
 * Zitadel API Client for browser-based authentication
 *
 * Updated to remove introspection for USER_AGENT apps
 *
 * This client implements the username/password authentication flow
 * as described in https://zitadel.com/docs/guides/integrate/login-ui/username-password
 */

export interface ZitadelSession {
  sessionId: string
  sessionToken: string
  factors?: {
    user?: {
      id: string
      loginName: string
      displayName?: string
    }
    password?: {
      verifiedAt: string
    }
  }
}

export interface ZitadelCallbackResponse {
  callbackUrl: string
}

export interface OIDCTokens {
  access_token: string
  id_token: string
  refresh_token?: string
  token_type: string
  expires_in: number
  scope?: string
}

export interface ZitadelError {
  code?: string
  message?: string
  details?: Array<{
    '@type'?: string
    [key: string]: unknown
  }>
}

export class ZitadelClientError extends Error {
  constructor(
    message: string,
    public code?: string,
    public statusCode?: number
  ) {
    super(message)
    this.name = 'ZitadelClientError'
  }
}

export interface ZitadelClientConfig {
  apiUrl: string
  bearerToken: string
  clientId: string
  redirectUri: string
}

/**
 * Client for interacting with Zitadel v2 APIs
 */
export class ZitadelClient {
  constructor(private config: ZitadelClientConfig) {}

  private get headers(): HeadersInit {
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${this.config.bearerToken}`,
    }
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      let errorMessage = `Request failed with status ${response.status}`
      let errorCode: string | undefined

      try {
        const errorData: ZitadelError = await response.json()
        errorCode = errorData.code
        errorMessage = errorData.message || errorMessage
      } catch {
        // If JSON parsing fails, use default error message
      }

      throw new ZitadelClientError(errorMessage, errorCode, response.status)
    }

    return response.json()
  }

  /**
   * Step 1: Create a session with username/login name
   */
  async createSession(loginName: string): Promise<ZitadelSession> {
    const response = await fetch(`${this.config.apiUrl}/v2/sessions`, {
      method: 'POST',
      headers: this.headers,
      body: JSON.stringify({
        checks: {
          user: {
            loginName
          }
        }
      })
    })

    return this.handleResponse<ZitadelSession>(response)
  }

  /**
   * Step 2: Verify password for an existing session
   */
  async verifyPassword(sessionId: string, password: string): Promise<ZitadelSession> {
    const response = await fetch(`${this.config.apiUrl}/v2/sessions/${sessionId}`, {
      method: 'PATCH',
      headers: this.headers,
      body: JSON.stringify({
        checks: {
          password: {
            password
          }
        }
      })
    })

    return this.handleResponse<ZitadelSession>(response)
  }

  /**
   * Step 3: Finalize the auth request with authenticated session (OIDC flow)
   */
  async finalizeAuthRequest(
    authRequestId: string,
    session: ZitadelSession
  ): Promise<ZitadelCallbackResponse> {
    const response = await fetch(
      `${this.config.apiUrl}/v2/oidc/auth_requests/${authRequestId}`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session: {
            sessionId: session.sessionId,
            sessionToken: session.sessionToken
          }
        })
      }
    )

    return this.handleResponse<ZitadelCallbackResponse>(response)
  }

  /**
   * Step 4: Exchange authorization code for tokens
   */
  async exchangeCodeForTokens(code: string): Promise<OIDCTokens> {
    const response = await fetch(`${this.config.apiUrl}/oauth/v2/token`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded'
      },
      body: new URLSearchParams({
        grant_type: 'authorization_code',
        client_id: this.config.clientId,
        code,
        redirect_uri: this.config.redirectUri
      })
    })

    return this.handleResponse<OIDCTokens>(response)
  }

  /**
   * Introspect a token (typically used when session token is used directly)
   */
  async introspectToken(token: string): Promise<{
    active: boolean
    scope?: string
    client_id?: string
    username?: string
    token_type?: string
    exp?: number
    iat?: number
    sub?: string
  }> {
    const response = await fetch(`${this.config.apiUrl}/oauth/v2/introspect`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
        'Authorization': `Bearer ${token}`,
      },
      body: new URLSearchParams({
        token,
        client_id: this.config.clientId
      })
    })

    return this.handleResponse(response)
  }

  /**
   * Complete username/password authentication flow
   */
  async authenticateWithPassword(
    loginName: string,
    password: string,
    authRequestId?: string
  ): Promise<OIDCTokens> {
    // Step 1: Create session with username
    const session = await this.createSession(loginName)

    // Step 2: Verify password
    const authenticatedSession = await this.verifyPassword(session.sessionId, password)

    // If we have an authRequestId, complete the OIDC flow
    if (authRequestId) {
      // Step 3: Finalize the auth request
      const callback = await this.finalizeAuthRequest(authRequestId, authenticatedSession)

      // Extract code from callback URL
      const callbackUrl = new URL(callback.callbackUrl)
      const code = callbackUrl.searchParams.get('code')

      if (!code) {
        throw new ZitadelClientError('No authorization code in callback URL')
      }

      // Step 4: Exchange code for tokens
      return this.exchangeCodeForTokens(code)
    } else {
      // Fallback: Use session token directly as access token
      // According to Zitadel docs: "the session token can be used as OAuth2 access token"
      // No introspection needed for USER_AGENT apps - the token is self-contained
      return {
        access_token: authenticatedSession.sessionToken,
        id_token: authenticatedSession.sessionToken,
        token_type: 'Bearer',
        expires_in: 3600, // Default to 1 hour
        scope: 'openid profile email'
      }
    }
  }
}
