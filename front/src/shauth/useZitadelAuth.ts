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

  // Step 1: Get session token from our backend
  const getSessionToken = async () => {
    setLoading(true)
    setError(null)

    try {
      console.log('Getting session token from backend...')
      const response = await client.POST('/user/session', {})

      if (response.error) {
        console.error('Backend error:', response.error)
        throw new Error(response.error.detail || 'Failed to get session token')
      }

      console.log('Session token received:', response.data)

      // Проверяем что данные пришли
      if (!response.data) {
        throw new Error('No data received from backend')
      }

      const { session_token, zitadel_url, expires_in } = response.data

      if (!session_token) {
        console.error('No session token in response:', response.data)
        throw new Error('No session token received from backend')
      }

      if (!zitadel_url) {
        console.error('No Zitadel URL in response:', response.data)
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
      console.error('Failed to get session token:', message)
      setError(message)
      throw new Error(message)
    } finally {
      setLoading(false)
    }
  }

  // Step 2: Create Zitadel session with username
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

      console.log('Request headers being sent:', headers)
      console.log('Full request URL:', `${zitadelUrl}/v2/sessions`)

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

  // Step 3: Add password to session
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

      console.log('Password request headers being sent:', headers)
      console.log('Password request URL:', `${zitadelUrl}/v2/sessions/${sessionId}`)

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

  // Complete authentication flow
  const authenticateWithZitadel = async (email: string, password: string) => {
    console.log('Starting authentication flow for:', email)
    setLoading(true)
    setError(null)

    try {
      // Step 1: Get session token from our backend
      console.log('Step 1: Getting session token from backend...')
      const tokenData = await getSessionToken()
      console.log('Token data received:', {
        hasToken: !!tokenData.sessionToken,
        zitadelUrl: tokenData.zitadelUrl,
        expiresIn: tokenData.expiresIn
      })

      // Step 2: Create session with username in Zitadel
      // Use the email directly as the login name (Zitadel v4.2.2+ behavior)
      const zitadelLoginName = email
      console.log('Step 2: Creating Zitadel session...')
      console.log(`Converting email "${email}" to Zitadel login name: "${zitadelLoginName}"`)
      const session = await createZitadelSession(zitadelLoginName, tokenData.zitadelUrl, tokenData.sessionToken)
      console.log('Session created:', session)

      // Step 3: Add password to session
      console.log('Step 3: Adding password to session...')
      const authenticatedSession = await addPasswordToSession(
        session.sessionId,
        password,
        tokenData.zitadelUrl,
        tokenData.sessionToken
      )

      // TODO: Store authenticated session token for API calls
      console.log('Authentication successful', authenticatedSession)

      return authenticatedSession
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Authentication failed'
      console.error('Authentication failed:', err)
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
