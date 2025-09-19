import { useState } from 'react'
import client from '../api/client'

export interface ZitadelSession {
  sessionId: string
  sessionToken: string
  changeDate?: string
}

export interface ZitadelSessionResponse {
  sessionId: string
  sessionToken: string
  changeDate?: string
  factors?: {
    user?: {
      verifiedAt: string
    }
    password?: {
      verifiedAt: string
    }
  }
}

export interface ZitadelAuthConfig {
  zitadelUrl: string
}

export function useZitadelAuth() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [authConfig, setAuthConfig] = useState<ZitadelAuthConfig | null>(null)

  const getSessionToken = async () => {
    setLoading(true)
    setError(null)

    try {
      const response = await client.POST('/user/session', {})

      if (response.error) {
        throw new Error(response.error.detail || 'Failed to get session token')
      }

      if (!response.data) {
        throw new Error('No data received from backend')
      }

      const { session_token, zitadel_url, expires_in } = response.data

      if (!session_token) {
        throw new Error('No session token received from backend')
      }

      if (!zitadel_url) {
        throw new Error('No Zitadel URL received from backend')
      }

      const config = {
        zitadelUrl: zitadel_url
      }

      setAuthConfig(config)

      console.log('Returning token data:', {
        sessionToken: session_token?.substring(0, 20) + '...',
        zitadelUrl: zitadel_url,
        expiresIn: expires_in
      })

      return {
        sessionToken: session_token,
        zitadelUrl: zitadel_url,
        expiresIn: expires_in
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to get session token'
      setError(message)
      throw new Error(message)
    } finally {
      setLoading(false)
    }
  }

  const createZitadelSession = async (loginName: string, zitadelUrl: string, bearerToken: string): Promise<ZitadelSession> => {
    setLoading(true)
    setError(null)

    try {
      console.log('Creating Zitadel session:', {
        loginName,
        zitadelUrl,
        bearerToken: bearerToken ? `Bearer ${bearerToken.substring(0, 20)}...` : 'NO TOKEN',
        fullBearerToken: bearerToken
      })

      const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${bearerToken}`,
      }

      const response = await fetch(`${zitadelUrl}/v2/sessions`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          checks: {
            user: {
              loginName: loginName
            }
          }
        })
      })

      if (!response.ok) {
        const errorData = await response.text()
        throw new Error(`Failed to create session: ${response.status} - ${errorData}`)
      }

      const sessionData: ZitadelSessionResponse = await response.json()

      return {
        sessionId: sessionData.sessionId,
        sessionToken: sessionData.sessionToken,
        changeDate: sessionData.changeDate
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to create Zitadel session'
      setError(message)
      throw new Error(message)
    } finally {
      setLoading(false)
    }
  }

  const addPasswordToSession = async (
    sessionId: string,
    password: string,
    zitadelUrl: string,
    bearerToken: string
  ): Promise<ZitadelSession> => {
    setLoading(true)
    setError(null)

    try {
      console.log('Adding password to session:', {
        sessionId,
        zitadelUrl,
        bearerToken: bearerToken ? `Bearer ${bearerToken.substring(0, 20)}...` : 'NO TOKEN',
        fullBearerToken: bearerToken
      })

      const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${bearerToken}`,
      }

      const response = await fetch(`${zitadelUrl}/v2/sessions/${sessionId}`, {
        method: 'PATCH',
        headers,
        body: JSON.stringify({
          checks: {
            password: {
              password: password
            }
          }
        })
      })

      if (!response.ok) {
        const errorData = await response.text()
        throw new Error(`Authentication failed: ${response.status} - ${errorData}`)
      }

      const sessionData: ZitadelSessionResponse = await response.json()

      return {
        sessionId: sessionData.sessionId,
        sessionToken: sessionData.sessionToken,
        changeDate: sessionData.changeDate
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Password verification failed'
      setError(message)
      throw new Error(message)
    } finally {
      setLoading(false)
    }
  }

  const authenticateWithZitadel = async (email: string, password: string) => {
    setLoading(true)
    setError(null)

    try {
      const tokenData = await getSessionToken()

      const zitadelLoginName = email
      const session = await createZitadelSession(zitadelLoginName, tokenData.zitadelUrl, tokenData.sessionToken)

      console.log('Step 3: Adding password to session...')
      const authenticatedSession = await addPasswordToSession(
        session.sessionId,
        password,
        tokenData.zitadelUrl,
        tokenData.sessionToken
      )

      return authenticatedSession
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Authentication failed'
      setError(message)
      throw new Error(message)
    } finally {
      setLoading(false)
    }
  }

  return {
    loading,
    error,
    authConfig,
    getSessionToken,
    createZitadelSession,
    addPasswordToSession,
    authenticateWithZitadel
  }
}
