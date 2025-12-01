/**
 * OAuth2 client for communicating with the backend OAuth2 endpoints
 */

const OAUTH2_BASE_URL = '/api/v1/auth/oauth2';

export interface OAuth2AuthorizeResponse {
  authorization_url: string;
  state: string;
}

export interface OAuth2RefreshResponse {
  expires_in: number;
}

export interface OAuth2SessionResponse {
  authenticated: boolean;
  expires_in?: number;
}

export interface OAuth2LogoutResponse {
  success: boolean;
}

export class OAuth2Error extends Error {
  status?: number;
  details?: unknown;

  constructor(message: string, status?: number, details?: unknown) {
    super(message);
    this.name = 'OAuth2Error';
    this.status = status;
    this.details = details;
  }
}

/**
 * Initiate the OAuth2 authorization flow
 * Returns the authorization URL to redirect to
 */
export async function initiateOAuth2Flow(
  redirectUri: string
): Promise<OAuth2AuthorizeResponse> {
  const response = await fetch(`${OAUTH2_BASE_URL}/authorize`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
    },
    body: JSON.stringify({
      redirect_uri: redirectUri,
    }),
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new OAuth2Error(
      error.detail || 'Failed to initiate OAuth2 flow',
      response.status,
      error
    );
  }

  return response.json();
}

/**
 * Refresh the access token using the refresh token cookie
 */
export async function refreshToken(): Promise<OAuth2RefreshResponse> {
  const response = await fetch(`${OAUTH2_BASE_URL}/refresh`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      Accept: 'application/json',
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new OAuth2Error(
      error.detail || 'Failed to refresh token',
      response.status,
      error
    );
  }

  return response.json();
}

/**
 * Check session status without triggering token refresh.
 * Returns unauthenticated for most errors, but throws for 404 (tenant not found).
 */
export async function checkSession(): Promise<OAuth2SessionResponse> {
  try {
    const response = await fetch(`${OAUTH2_BASE_URL}/session`, {
      method: 'GET',
      credentials: 'include',
      headers: {
        Accept: 'application/json',
      },
    });

    // 404 means tenant doesn't exist - throw to signal this to the caller
    if (response.status === 404) {
      throw new OAuth2Error('Tenant not found', 404);
    }

    if (!response.ok) {
      // Other errors - return unauthenticated instead of throwing
      return { authenticated: false };
    }

    return response.json();
  } catch (err) {
    // Re-throw OAuth2Error (including 404 tenant not found)
    if (err instanceof OAuth2Error) {
      throw err;
    }
    // Network error - return unauthenticated
    return { authenticated: false };
  }
}

/**
 * Logout and revoke tokens
 */
export async function logout(): Promise<OAuth2LogoutResponse> {
  const response = await fetch(`${OAUTH2_BASE_URL}/logout`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      Accept: 'application/json',
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new OAuth2Error(
      error.detail || 'Failed to logout',
      response.status,
      error
    );
  }

  return response.json();
}

/**
 * Redirect to the OAuth2 authorization URL
 */
export function redirectToAuth(authorizationUrl: string): void {
  window.location.href = authorizationUrl;
}

/**
 * Check if we're returning from an OAuth2 callback
 * The callback redirects to the redirect_uri after setting cookies
 */
export function isOAuth2Callback(): boolean {
  const params = new URLSearchParams(window.location.search);
  return params.has('oauth2_success') || params.has('oauth2_error');
}

/**
 * Clear OAuth2 callback params from URL
 */
export function clearCallbackParams(): void {
  const url = new URL(window.location.href);
  url.searchParams.delete('oauth2_success');
  url.searchParams.delete('oauth2_error');
  window.history.replaceState({}, '', url.toString());
}
