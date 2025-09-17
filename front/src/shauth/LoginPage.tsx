import { useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Button, Flex, Form, Header, Link, Text, TextField, View } from '@adobe/react-spectrum'
import Alert from '@spectrum-icons/workflow/Alert'

interface FormFields {
  email: string
  password: string
}

export function LoginPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [loginError, setLoginError] = useState<string | null>(null)

  const form = useForm({
    defaultValues: { email: '', password: '' },
  })

  const onSubmit = async (fields: FormFields) => {
    setLoginError(null)
    try {
      const response = await fetch('/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Accept: 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          email: fields.email,
          password: fields.password,
        }),
      })

      if (response.ok) {
        const data = await response.json()
        if (data.success) {
          const returnTo = searchParams.get('returnTo') || '/'
          navigate(returnTo)
        } else {
          setLoginError('Login failed')
        }
      } else if (response.status === 401) {
        setLoginError('Invalid email or password')
      } else {
        const errorData = await response.json().catch(() => ({ message: response.statusText }))
        setLoginError(`Authentication failed: ${errorData.message || 'Unknown error'}`)
      }
    } catch (error) {
      setLoginError(`Network error: ${error instanceof Error ? error.message : 'Unknown error'}`)
    }
  }

  return (
    <Flex direction="row" alignItems="center" justifyContent="center" flexBasis="100%" height="100vh">
      <View padding="size-200" backgroundColor="gray-200" borderRadius="medium" width="size-4600">
        <Flex direction="column" gap="size-200">
          <Header>Login to ShadowAPI</Header>

          {loginError && (
            <View backgroundColor="negative" padding="size-100" borderRadius="regular">
              <Flex gap="size-100" alignItems="center">
                <Alert color="negative" />
                <Text>{loginError}</Text>
              </Flex>
            </View>
          )}

          {/* Email/Password Login Form */}
          <Form onSubmit={form.handleSubmit(onSubmit)}>
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
                render={({ field: { name, value, onChange, onBlur, ref }, fieldState: { invalid, error } }) => (
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
                    validationState={invalid ? 'invalid' : undefined}
                    errorMessage={error?.message}
                  />
                )}
              />
              <Controller
                name="password"
                control={form.control}
                rules={{ required: 'Password is required' }}
                render={({ field: { name, value, onChange, onBlur, ref }, fieldState: { invalid, error } }) => (
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
                    validationState={invalid ? 'invalid' : undefined}
                    errorMessage={error?.message}
                  />
                )}
              />
              <Flex justifyContent="space-between" alignItems="center" marginTop="size-150">
                <Text>
                  Don't have an account? <Link href="/signup">Sign up</Link>
                </Text>
                <Button variant="cta" type="submit">
                  Login
                </Button>
              </Flex>
            </Flex>
          </Form>
        </Flex>
      </View>
    </Flex>
  )
}
