import { useState } from 'react'
import client from '../api/client'
import { config } from '../config/env'

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

export interface ZitadelError {
  code?: string
  message?: string
  details?: Array<{
    '@type': string
    violations?: Array<{
      field: string
      description: string
    }>
  }>
}

export interface OIDCTokens {
  access_token: string
  id_token: string
  refresh_token?: string
  token_type: string
  expires_in: number
}

export function useZitadelAuth() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [authConfig, setAuthConfig] = useState<ZitadelAuthConfig | null>(null)

  const parseZitadelError = async (response: Response): Promise<{ message: string; fieldErrors: Record<string, string> }> => {
    try {
      const errorData: ZitadelError = await response.json()
      
      let message = errorData.message || `Request failed with status ${response.status}`
      const fieldErrors: Record<string, string> = {}

      // Extract field-specific errors from violations
      if (errorData.details) {
        for (const detail of errorData.details) {
          if (detail.violations) {
            for (const violation of detail.violations) {
              if (violation.field && violation.description) {
                // Map common Zitadel field names to our form fields
                const fieldName = violation.field.toLowerCase().includes('loginname') || 
                                violation.field.toLowerCase().includes('email') ? 'email' : 
                                violation.field.toLowerCase().includes('password') ? 'password' : violation.field
                fieldErrors[fieldName] = violation.description
              }
            }
          }
        }
      }

      // Common Zitadel error codes to user-friendly messages
      if (errorData.code) {
        switch (errorData.code) {
          case 'PRECONDITION_FAILED':
            message = 'Invalid email or password'
            fieldErrors.email = 'Invalid email or password'
            break
          case 'NOT_FOUND':
            message = 'User not found'
            fieldErrors.email = 'User not found'
            break
          case 'INVALID_ARGUMENT':
            message = 'Invalid credentials'
            fieldErrors.email = 'Invalid credentials'
            break
          default:
            // Keep the original message or use a fallback
            break
        }
      }

      return { message, fieldErrors }
    } catch (parseError) {
      // Fallback if JSON parsing fails
      const text = await response.text()
      return { 
        message: text || `Request failed with status ${response.status}`, 
        fieldErrors: {} 
      }
    }
  }

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
        const { message, fieldErrors } = await parseZitadelError(response)
        setFieldErrors(fieldErrors)
        throw new Error(message)
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
        const { message, fieldErrors } = await parseZitadelError(response)
        setFieldErrors(fieldErrors)
        throw new Error(message)
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

  const exchangeSessionTokenForJWT = async (
    sessionToken: string,
    zitadelUrl: string
  ): Promise<OIDCTokens> => {
    setLoading(true)
    setError(null)

    try {
      console.log('Exchanging session token for JWT tokens...')

      const params = new URLSearchParams({
        grant_type: 'urn:ietf:params:oauth:grant-type:token-exchange',
        client_id: config.zitadel.clientId,
        subject_token: sessionToken,
        subject_token_type: 'urn:ietf:params:oauth:token-type:access_token',
        requested_token_type: 'urn:ietf:params:oauth:token-type:jwt',
        scope: 'openid profile email',
      })

      const response = await fetch(`${zitadelUrl}/oauth/v2/token`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: params.toString(),
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`Token exchange failed: ${errorText}`)
      }

      const tokens: OIDCTokens = await response.json()
      console.log('Successfully obtained JWT tokens via token exchange')

      return tokens
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to exchange session token for JWT'
      setError(message)
      throw new Error(message)
    } finally {
      setLoading(false)
    }
  }

  const authenticateAndFinalizeAuthRequest = async (
    email: string,
    password: string
  ): Promise<OIDCTokens> => {
    setLoading(true)
    setError(null)
    setFieldErrors({})

    try {
      console.log('Step 1: Getting backend session token...')
      // Step 1: Get session token from backend
      const tokenData = await getSessionToken()

      console.log('Step 2: Creating Zitadel session...')
      // Step 2: Create Zitadel session with username
      const session = await createZitadelSession(email, tokenData.zitadelUrl, tokenData.sessionToken)

      console.log('Step 3: Adding password to session...')
      // Step 3: Add password to session
      const authenticatedSession = await addPasswordToSession(
        session.sessionId,
        password,
        tokenData.zitadelUrl,
        tokenData.sessionToken
      )

      console.log('Step 4: Exchanging session token for JWT tokens...')
      // Step 4: Exchange session token for JWT tokens
      return await exchangeSessionTokenForJWT(authenticatedSession.sessionToken, tokenData.zitadelUrl)
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
    fieldErrors,
    authConfig,
    getSessionToken,
    createZitadelSession,
    addPasswordToSession,
    authenticateAndFinalizeAuthRequest
  }
}
