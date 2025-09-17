import {
  Form,
  Flex,
  View,
  Header,
  Button,
  TextField,
  Link,
  Text,
} from '@adobe/react-spectrum';
import Alert from '@spectrum-icons/workflow/Alert'

import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { useForm, Controller } from "react-hook-form"

interface FormFields {
  username: string;
  email: string;
  firstName: string;
  lastName: string;
  password: string;
  confirmPassword: string;
}

export function SignupPage() {
  const navigate = useNavigate()
  const [signupError, setSignupError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const form = useForm({
    defaultValues: {
      username: "",
      email: "",
      firstName: "",
      lastName: "",
      password: "",
      confirmPassword: ""
    },
  })

  const onSubmit = async (fields: FormFields) => {
    setSignupError(null)
    setIsLoading(true)

    if (fields.password !== fields.confirmPassword) {
      setSignupError('Passwords do not match')
      setIsLoading(false)
      return
    }

    try {
      const zitadelUrl = import.meta.env.VITE_ZITADEL_URL || 'http://auth.localtest.me'

      const userResponse = await fetch(`${zitadelUrl}/v2/users/new`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          human: {
            username: fields.username,
            profile: {
              givenName: fields.firstName,
              familyName: fields.lastName,
              displayName: `${fields.firstName} ${fields.lastName}`
            },
            email: {
              email: fields.email,
              isVerified: false
            },
            password: {
              password: fields.password,
              changeRequired: false
            }
          }
        }),
      })

      if (!userResponse.ok) {
        const errorText = await userResponse.text()
        let errorMessage = 'Registration failed'

        if (userResponse.status === 409) {
          errorMessage = 'Username or email already exists'
        } else if (userResponse.status === 400) {
          errorMessage = 'Invalid registration data'
        } else {
          errorMessage = `Registration failed: ${errorText || userResponse.statusText}`
        }

        setSignupError(errorMessage)
        setIsLoading(false)
        return
      }

      const sessionResponse = await fetch(`${zitadelUrl}/v2/sessions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          checks: {
            user: {
              loginName: fields.username
            }
          }
        }),
      })

      if (!sessionResponse.ok) {
        setSignupError('Registration successful, but login failed. Please try logging in.')
        setIsLoading(false)
        return
      }

      const sessionData = await sessionResponse.json()
      const sessionId = sessionData.sessionId

      const passwordResponse = await fetch(`${zitadelUrl}/v2/sessions/${sessionId}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          checks: {
            password: {
              password: fields.password
            }
          }
        }),
      })

      if (passwordResponse.ok) {
        navigate('/')
      } else {
        setSignupError('Registration successful, but login failed. Please try logging in.')
      }
    } catch (error) {
      setSignupError(`Network error: ${error instanceof Error ? error.message : 'Unknown error'}`)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Flex direction="row" alignItems="center" justifyContent="center" flexBasis="100%" height="100vh">
      <View
        padding="size-200"
        backgroundColor="gray-200"
        borderRadius="medium"
        width="size-4600"
      >
        <Flex direction="column" gap="size-200">
          <Header>Sign Up for ShadowAPI</Header>

          {signupError && (
            <View backgroundColor="negative" padding="size-100" borderRadius="regular">
              <Flex gap="size-100" alignItems="center">
                <Alert color="negative" />
                <Text>{signupError}</Text>
              </Flex>
            </View>
          )}

          <Form onSubmit={form.handleSubmit(onSubmit)}>
            <Flex direction="column" gap="size-100">
              <Flex direction="row" gap="size-100">
                <Controller
                  name="firstName"
                  control={form.control}
                  rules={{ required: 'First name is required' }}
                  render={({
                    field: { name, value, onChange, onBlur, ref },
                    fieldState: { invalid, error },
                  }) => (
                    <TextField
                      label="First Name"
                      type="text"
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
                  name="lastName"
                  control={form.control}
                  rules={{ required: 'Last name is required' }}
                  render={({
                    field: { name, value, onChange, onBlur, ref },
                    fieldState: { invalid, error },
                  }) => (
                    <TextField
                      label="Last Name"
                      type="text"
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
              </Flex>

              <Controller
                name="username"
                control={form.control}
                rules={{ required: 'Username is required' }}
                render={({
                  field: { name, value, onChange, onBlur, ref },
                  fieldState: { invalid, error },
                }) => (
                  <TextField
                    label="Username"
                    type="text"
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
                name="email"
                control={form.control}
                rules={{
                  required: 'Email is required',
                  pattern: {
                    value: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
                    message: 'Please enter a valid email address'
                  }
                }}
                render={({
                  field: { name, value, onChange, onBlur, ref },
                  fieldState: { invalid, error },
                }) => (
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
                rules={{
                  required: 'Password is required',
                  minLength: {
                    value: 8,
                    message: 'Password must be at least 8 characters long'
                  }
                }}
                render={({
                  field: { name, value, onChange, onBlur, ref },
                  fieldState: { invalid, error },
                }) => (
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

              <Controller
                name="confirmPassword"
                control={form.control}
                rules={{
                  required: 'Please confirm your password',
                  validate: (value) =>
                    value === form.watch('password') || 'Passwords do not match'
                }}
                render={({
                  field: { name, value, onChange, onBlur, ref },
                  fieldState: { invalid, error },
                }) => (
                  <TextField
                    label="Confirm Password"
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
                  Already have an account?{" "}
                  <Link href="/login">Login</Link>
                </Text>
                <Button
                  variant="cta"
                  type="submit"
                  isDisabled={isLoading}
                >
                  {isLoading ? 'Creating Account...' : 'Sign Up'}
                </Button>
              </Flex>
            </Flex>
          </Form>
        </Flex>
      </View>
    </Flex>
  )
}
