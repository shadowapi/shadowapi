import { useEffect, useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Button, Flex, Form, Header, Link, Text, TextField, View, ProgressCircle } from '@adobe/react-spectrum'
import Alert from '@spectrum-icons/workflow/Alert'
import { useZitadelAuth } from './useZitadelAuth'
import { useAuth } from './AuthContext'
import { config } from '../config/env'

interface FormFields {
  email: string
  password: string
}

export function LoginPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { loading, error, fieldErrors, authenticateAndFinalizeAuthRequest } = useZitadelAuth()
  const { login } = useAuth()

  const form = useForm({
    defaultValues: { email: '', password: '' },
  })

  const authRequestId = searchParams.get('authRequest')

  // If we have authRequest and pending credentials, auto-fill and submit
  useEffect(() => {
    if (authRequestId) {
      const pendingLogin = sessionStorage.getItem('pending_login')
      if (pendingLogin) {
        try {
          const credentials = JSON.parse(pendingLogin)
          form.setValue('email', credentials.email)
          form.setValue('password', credentials.password)
          // Auto-submit the form
          form.handleSubmit(onSubmit)()
        } catch (err) {
          console.error('Failed to parse pending login:', err)
          sessionStorage.removeItem('pending_login')
        }
      }
    }
  }, [authRequestId])

  const onSubmit = async (fields: FormFields) => {
    try {
      if (!authRequestId) {
        // Step 1: No authRequest yet - initiate OIDC flow by redirecting to authorize endpoint
        // Store credentials temporarily so we can use them after redirect
        sessionStorage.setItem('pending_login', JSON.stringify({ email: fields.email, password: fields.password }))

        const authorizeParams = new URLSearchParams({
          client_id: config.zitadel.clientId,
          redirect_uri: config.zitadel.redirectUri,
          response_type: 'code',
          scope: 'openid profile email',
        })

        const authorizeUrl = `${config.zitadel.url}/oauth/v2/authorize?${authorizeParams.toString()}`

        // Redirect browser to Zitadel authorize endpoint
        // This will redirect back to our redirect_uri with an authRequest parameter
        window.location.href = authorizeUrl
        return
      }

      // Step 2: We have authRequest - authenticate and finalize
      const tokens = await authenticateAndFinalizeAuthRequest(
        fields.email,
        fields.password,
        authRequestId
      )

      // Clear pending login
      sessionStorage.removeItem('pending_login')

      // Store JWT tokens in auth context
      login(
        fields.email,
        tokens.access_token,
        tokens.id_token,
        tokens.refresh_token,
        tokens.expires_in
      )

      // If successful, redirect to the desired page
      const returnTo = searchParams.get('returnTo') || '/'
      navigate(returnTo)
    } catch (err) {
      // Error is already handled by useZitadelAuth hook
      console.error('Login failed:', err)
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
                    message: 'Invalid email address'
                  }
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
                  Don't have an account? <Link href="/signup">Sign up</Link>
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

