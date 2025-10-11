const CHARSET = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~'
const PKCE_STORAGE_KEY = 'shadowapi_zitadel_pkce'

export type StoredPkce = {
  codeVerifier: string
  createdAt: number
  state?: string
  returnTo?: string
  codeChallengeMethod?: 'S256' | 'plain'
  sessionToken?: string
  zitadelUrl?: string
  sessionExpiresIn?: number
}

export function generateCodeVerifier(length = 64): string {
  const array = new Uint8Array(length)
  if (typeof window !== 'undefined' && window.crypto?.getRandomValues) {
    window.crypto.getRandomValues(array)
  } else {
    for (let i = 0; i < length; i++) array[i] = Math.floor(Math.random() * CHARSET.length)
  }
  return Array.from(array, (byte) => CHARSET[byte % CHARSET.length]).join('')
}

export async function generateCodeChallenge(
  verifier: string,
): Promise<{ challenge: string; method: 'S256' | 'plain' }> {
  if (typeof window === 'undefined') {
    throw new Error('PKCE requires a browser environment')
  }

  if (!window.crypto?.subtle) {
    console.warn('PKCE S256 unavailable; falling back to plain code challenge')
    return { challenge: verifier, method: 'plain' }
  }

  const encoder = new TextEncoder()
  const digest = await window.crypto.subtle.digest('SHA-256', encoder.encode(verifier))
  const hashArray = Array.from(new Uint8Array(digest))
  const base64 = btoa(String.fromCharCode(...hashArray))
  const challenge = base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
  return { challenge, method: 'S256' }
}

type StorePkceOptions = {
  verifier: string
  state?: string
  returnTo?: string
  codeChallengeMethod?: 'S256' | 'plain'
  sessionToken?: string
  zitadelUrl?: string
  sessionExpiresIn?: number
}

export function storePkce({
  verifier,
  state,
  returnTo,
  codeChallengeMethod,
  sessionToken,
  zitadelUrl,
  sessionExpiresIn,
}: StorePkceOptions) {
  if (typeof window === 'undefined') return
  const payload: StoredPkce = {
    codeVerifier: verifier,
    createdAt: Date.now(),
    state,
    returnTo,
    codeChallengeMethod,
    sessionToken,
    zitadelUrl,
    sessionExpiresIn,
  }
  window.sessionStorage.setItem(PKCE_STORAGE_KEY, JSON.stringify(payload))
}

export function loadPkce(): StoredPkce | null {
  if (typeof window === 'undefined') return null
  const raw = window.sessionStorage.getItem(PKCE_STORAGE_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as StoredPkce
  } catch (error) {
    console.error('Failed to parse stored PKCE payload', error)
    window.sessionStorage.removeItem(PKCE_STORAGE_KEY)
    return null
  }
}

export function clearPkce() {
  if (typeof window === 'undefined') return
  window.sessionStorage.removeItem(PKCE_STORAGE_KEY)
}

export const PKCE_STORAGE_KEY_NAME = PKCE_STORAGE_KEY
