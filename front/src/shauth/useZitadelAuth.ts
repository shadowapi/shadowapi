import { useCallback, useState } from 'react'
import client from '../api/client'
import { config } from '../config/env'
import { ZitadelClient, ZitadelClientError, OIDCTokens } from './ZitadelClient'

export interface ZitadelSessionContext {
  sessionToken: string
  zitadelUrl: string
  expiresIn?: number
}

export function useZitadelAuth() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})

  const createSessionContext = useCallback(async (): Promise<ZitadelSessionContext> => {
    const sessionResponse = await client.POST('/user/session', {})
    if (sessionResponse.error || !sessionResponse.data) {
      throw new Error(sessionResponse.error?.detail || 'Failed to get session token')
    }

    const { session_token: sessionToken, zitadel_url: zitadelUrl, expires_in: expiresIn } = sessionResponse.data
    if (!sessionToken || !zitadelUrl) {
      throw new Error('Missing session token or Zitadel URL')
    }

    const normalize = (value?: string) => (value ? value.replace(/\/$/, '') : value)
    const normalizedUrl = normalize(zitadelUrl) ?? zitadelUrl
    const publicUrl = normalize(config.zitadel.publicUrl)
    const resolvedUrl = publicUrl ?? normalizedUrl

    return {
      sessionToken,
      zitadelUrl: resolvedUrl,
      expiresIn,
    }
  }, [])

  const authenticateAndFinalizeAuthRequest = async (
    email: string,
    password: string,
    authRequestId?: string,
    codeVerifier?: string,
    sessionOverride?: ZitadelSessionContext,
  ): Promise<OIDCTokens> => {
    setLoading(true)
    setError(null)
    setFieldErrors({})

    try {
      // Step 1: Get backend session token for Zitadel API authentication
      const sessionContext = sessionOverride ?? (await createSessionContext())

      // Step 2: Create Zitadel client with bearer token
      const zitadelClient = new ZitadelClient({
        apiUrl: sessionContext.zitadelUrl,
        bearerToken: sessionContext.sessionToken,
        clientId: config.zitadel.clientId,
        redirectUri: config.zitadel.redirectUri,
      })

      // Step 3: Authenticate with username/password
      const tokens = await zitadelClient.authenticateWithPassword(email, password, authRequestId, codeVerifier)

      return tokens
    } catch (err) {
      let message = 'Authentication failed'

      if (err instanceof ZitadelClientError) {
        message = err.message

        // Map common error codes to field-specific errors
        if (err.code === 'PRECONDITION_FAILED' || err.code === 'NOT_FOUND' || err.code === 'INVALID_ARGUMENT') {
          const fieldMessage = 'Invalid email or password'
          setFieldErrors({ email: fieldMessage })
          message = fieldMessage
        }
      } else if (err instanceof Error) {
        message = err.message
      }

      setError(message)
      throw err
    } finally {
      setLoading(false)
    }
  }

  return {
    loading,
    error,
    fieldErrors,
    authenticateAndFinalizeAuthRequest,
    createSessionContext,
  }
}
