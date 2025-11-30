/**
 * PKCE (Proof Key for Code Exchange) utilities for OAuth2
 *
 * PKCE is used to protect the authorization code flow for public clients (like SPAs).
 * The code verifier is stored in sessionStorage during the OAuth flow.
 */

const STORAGE_KEY = 'oauth2_pkce_verifier';

/**
 * Generate a random code verifier (43-128 characters, URL-safe)
 */
export function generateCodeVerifier(): string {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return base64UrlEncode(array);
}

/**
 * Generate a code challenge from a verifier using S256 method
 */
export async function generateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await crypto.subtle.digest('SHA-256', data);
  return base64UrlEncode(new Uint8Array(digest));
}

/**
 * Base64 URL-safe encoding (no padding)
 */
function base64UrlEncode(buffer: Uint8Array): string {
  const base64 = btoa(String.fromCharCode(...buffer));
  return base64
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '');
}

/**
 * Store the code verifier in sessionStorage (cleared after OAuth callback)
 */
export function storeCodeVerifier(verifier: string): void {
  sessionStorage.setItem(STORAGE_KEY, verifier);
}

/**
 * Retrieve and remove the code verifier from sessionStorage
 */
export function retrieveCodeVerifier(): string | null {
  const verifier = sessionStorage.getItem(STORAGE_KEY);
  sessionStorage.removeItem(STORAGE_KEY);
  return verifier;
}

/**
 * Generate and store a new PKCE pair
 * Returns the code challenge for the authorization request
 */
export async function preparePKCE(): Promise<{ verifier: string; challenge: string }> {
  const verifier = generateCodeVerifier();
  const challenge = await generateCodeChallenge(verifier);
  storeCodeVerifier(verifier);
  return { verifier, challenge };
}
