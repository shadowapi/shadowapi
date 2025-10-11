import { useEffect, useRef, useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Button, Flex, Form, Header, Link, Text, TextField, View, ProgressCircle } from '@adobe/react-spectrum'
import Alert from '@spectrum-icons/workflow/Alert'
import { useZitadelAuth, type ZitadelSessionContext } from './useZitadelAuth'
import { useAuth } from './AuthContext'
import { config } from '../config/env'
import { clearPkce, generateCodeChallenge, generateCodeVerifier, loadPkce, storePkce } from './pkce'

interface FormFields {
  email: string
  password: string
}

export function LoginPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { loading, error, fieldErrors, authenticateAndFinalizeAuthRequest, createSessionContext } = useZitadelAuth()
  const { login } = useAuth()
  const [pkceError, setPkceError] = useState<string | null>(null)
  const authRequestId = searchParams.get('authRequest') || searchParams.get('authRequestId') || undefined
  const authState = searchParams.get('state') || undefined
  const initAuthFlow = useRef(false)
  const sessionContextRef = useRef<ZitadelSessionContext | null>(null)

  const form = useForm({
    defaultValues: { email: '', password: '' },
  })

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }

    if (authRequestId) {
      const storedPkce = loadPkce()
      if (storedPkce?.state && authState && storedPkce.state !== authState) {
        setPkceError('Login session mismatch. Please restart the sign-in flow.')
      }
      if (storedPkce?.sessionToken && storedPkce?.zitadelUrl) {
        sessionContextRef.current = {
          sessionToken: storedPkce.sessionToken,
          zitadelUrl: storedPkce.zitadelUrl,
          expiresIn: storedPkce.sessionExpiresIn,
        }
      }
      return
    }

    if (initAuthFlow.current) {
      return
    }

    if (!config.zitadel.url || !config.zitadel.clientId || !config.zitadel.redirectUri) {
      setPkceError('Authentication service is not configured. Please contact support.')
      return
    }

    initAuthFlow.current = true

    const startAuthFlow = async () => {
      try {
        const sessionContext = await createSessionContext()
        sessionContextRef.current = sessionContext

        const codeVerifier = generateCodeVerifier()
        let codeChallenge = codeVerifier
        let codeChallengeMethod: 'S256' | 'plain' = 'plain'
        try {
          const result = await generateCodeChallenge(codeVerifier)
          codeChallenge = result.challenge
          codeChallengeMethod = result.method
        } catch (challengeError) {
          console.warn('Falling back to plain PKCE challenge', challengeError)
        }
        const stateValue = typeof crypto.randomUUID === 'function' ? crypto.randomUUID() : generateCodeVerifier(16)
        const returnTo = searchParams.get('returnTo') || '/'

        storePkce({
          verifier: codeVerifier,
          state: stateValue,
          returnTo,
          codeChallengeMethod,
          sessionToken: sessionContext.sessionToken,
          zitadelUrl: sessionContext.zitadelUrl,
          sessionExpiresIn: sessionContext.expiresIn,
        })

        const normalize = (value?: string) => (value ? value.replace(/\/$/, '') : value)
        const fallbackAuthorizeBase = normalize(config.zitadel.publicUrl) ?? normalize(config.zitadel.url) ?? ''
        const authorizeBase = normalize(sessionContext.zitadelUrl) ?? fallbackAuthorizeBase
        const authorizeUrl = new URL(`${authorizeBase}/oauth/v2/authorize`)
        authorizeUrl.searchParams.set('client_id', config.zitadel.clientId)
        authorizeUrl.searchParams.set('redirect_uri', config.zitadel.redirectUri)
        authorizeUrl.searchParams.set('response_type', 'code')
        authorizeUrl.searchParams.set('scope', 'openid profile email')
        authorizeUrl.searchParams.set('code_challenge', codeChallenge)
        authorizeUrl.searchParams.set('code_challenge_method', codeChallengeMethod)
        authorizeUrl.searchParams.set('state', stateValue)
        authorizeUrl.searchParams.set('prompt', 'login')

        window.location.href = authorizeUrl.toString()
      } catch (startError) {
        console.error('Failed to initiate PKCE flow', startError)
        initAuthFlow.current = false
        setPkceError('Failed to start the authentication flow. Please refresh and try again.')
      }
    }

    void startAuthFlow()
  }, [authRequestId, authState, searchParams, createSessionContext])

  const onSubmit = async (fields: FormFields) => {
    try {
      const existingPkceError = pkceError
      setPkceError(null)

      if (existingPkceError) {
        throw new Error(existingPkceError)
      }

      if (!authRequestId) {
        throw new Error('Login session expired. Please start the sign-in flow again.')
      }

      const storedPkce = loadPkce()
      if (!storedPkce?.codeVerifier) {
        throw new Error('Missing PKCE verifier. Please restart the sign-in flow.')
      }

      if (storedPkce.state && authState && storedPkce.state !== authState) {
        clearPkce()
        throw new Error('Login session mismatch. Please restart the sign-in flow.')
      }

      let sessionOverride = sessionContextRef.current ?? undefined
      if (storedPkce.sessionToken && storedPkce.zitadelUrl) {
        sessionOverride = {
          sessionToken: storedPkce.sessionToken,
          zitadelUrl: storedPkce.zitadelUrl,
          expiresIn: storedPkce.sessionExpiresIn,
        }
      }

      // Call the complete authentication flow via API only
      const tokens = await authenticateAndFinalizeAuthRequest(
        fields.email,
        fields.password,
        authRequestId,
        storedPkce.codeVerifier,
        sessionOverride,
      )

      // Store JWT tokens in auth context
      login(fields.email, tokens.access_token, tokens.id_token, tokens.refresh_token, tokens.expires_in)

      // If successful, redirect to the desired page
      const returnTo = searchParams.get('returnTo') || storedPkce.returnTo || '/'
      clearPkce()
      sessionContextRef.current = null
      navigate(returnTo)
    } catch (err) {
      // Error is already handled by useZitadelAuth hook
      console.error('Login failed:', err)
      if (
        err instanceof Error &&
        (err.message.includes('PKCE') ||
          err.message.includes('Login session') ||
          err.message.includes('Authentication service'))
      ) {
        setPkceError(err.message)
      }
    }
  }

  return (
    <Flex direction="row" alignItems="center" justifyContent="center" flexBasis="100%" height="100vh">
      <View padding="size-200" backgroundColor="gray-200" borderRadius="medium" width="size-4600">
        <Flex direction="column" gap="size-200">
          <Header>Login to ShadowAPI</Header>

          {error && !fieldErrors.email && !fieldErrors.password && (
            <View backgroundColor="negative" padding="size-100" borderRadius="regular">
              <Flex gap="size-100" alignItems="center">
                <Alert color="negative" />
                <Text>{error}</Text>
              </Flex>
            </View>
          )}
          {pkceError && (
            <View backgroundColor="negative" padding="size-100" borderRadius="regular">
              <Flex gap="size-100" alignItems="center">
                <Alert color="negative" />
                <Text>{pkceError}</Text>
              </Flex>
            </View>
          )}

          {/* Email/Password Login Form */}
          <Form onSubmit={form.handleSubmit(onSubmit)}>
            {loading && (
              <View backgroundColor="informative" padding="size-100" borderRadius="regular">
                <Flex gap="size-100" alignItems="center">
                  <ProgressCircle size="S" isIndeterminate />
                  <Text>Connecting to Zitadel authentication...</Text>
                </Flex>
              </View>
            )}
            <Flex direction="column" gap="size-100">
              <Controller
                name="email"
                control={form.control}
                rules={{
                  required: 'Email is required',
                  pattern: {
                    value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
                    message: 'Invalid email address',
                  },
                }}
                render={({ field: { name, value, onChange, onBlur, ref }, fieldState: { invalid, error } }) => {
                  const fieldError = fieldErrors.email || error?.message
                  const hasError = invalid || !!fieldErrors.email
                  return (
                    <TextField
                      label="Email"
                      type="email"
                      width="100%"
                      isRequired
                      name={name}
                      value={value}
                      onChange={onChange}
                      onBlur={onBlur}
                      ref={ref}
                      validationState={hasError ? 'invalid' : undefined}
                      errorMessage={fieldError}
                    />
                  )
                }}
              />
              <Controller
                name="password"
                control={form.control}
                rules={{ required: 'Password is required' }}
                render={({ field: { name, value, onChange, onBlur, ref }, fieldState: { invalid, error } }) => {
                  const fieldError = fieldErrors.password || error?.message
                  const hasError = invalid || !!fieldErrors.password
                  return (
                    <TextField
                      label="Password"
                      type="password"
                      width="100%"
                      isRequired
                      name={name}
                      value={value}
                      onChange={onChange}
                      onBlur={onBlur}
                      ref={ref}
                      validationState={hasError ? 'invalid' : undefined}
                      errorMessage={fieldError}
                    />
                  )
                }}
              />
              <Flex justifyContent="space-between" alignItems="center" marginTop="size-150">
                <Text>
                  {"Don't have an account? "}
                  <Link href="/signup">Sign up</Link>
                </Text>
                <Button variant="cta" type="submit" isDisabled={loading}>
                  {loading ? (
                    <Flex alignItems="center" gap="size-100">
                      <ProgressCircle size="S" isIndeterminate />
                      <Text>Authenticating...</Text>
                    </Flex>
                  ) : (
                    'Login'
                  )}
                </Button>
              </Flex>
            </Flex>
          </Form>
        </Flex>
      </View>
    </Flex>
  )
}
