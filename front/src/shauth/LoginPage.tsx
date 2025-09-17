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
import { useNavigate, useSearchParams } from "react-router-dom"
import { useForm, Controller } from "react-hook-form"

interface FormFields {
  username: string;
  password: string;
}

export function LoginPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [loginError, setLoginError] = useState<string | null>(null)

  const form = useForm({
    defaultValues: { username: "", password: "" },
  })

  const onSubmit = async (fields: FormFields) => {
    setLoginError(null)
    try {
      const zitadelUrl = import.meta.env.VITE_ZITADEL_URL || 'http://auth.localtest.me'

      // Step 1: Create session with username check
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
        if (sessionResponse.status === 404) {
          setLoginError('Username not found')
        } else {
          const errorText = await sessionResponse.text()
          setLoginError(`Session creation failed: ${errorText || sessionResponse.statusText}`)
        }
        return
      }

      const sessionData = await sessionResponse.json()
      const sessionId = sessionData.sessionId

      // Step 2: Update session with password
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
        // Step 3: Redirect to application with session established
        const returnTo = searchParams.get('returnTo') || '/'
        navigate(returnTo)
      } else if (passwordResponse.status === 401) {
        setLoginError('Invalid password')
      } else {
        const errorText = await passwordResponse.text()
        setLoginError(`Authentication failed: ${errorText || passwordResponse.statusText}`)
      }
    } catch (error) {
      setLoginError(`Network error: ${error instanceof Error ? error.message : 'Unknown error'}`)
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
          <Header>Login to ShadowAPI</Header>

          {loginError && (
            <View backgroundColor="negative" padding="size-100" borderRadius="regular">
              <Flex gap="size-100" alignItems="center">
                <Alert color="negative" />
                <Text>{loginError}</Text>
              </Flex>
            </View>
          )}

          {/* Username/Password Login Form */}
          <Form onSubmit={form.handleSubmit(onSubmit)}>
            <Flex direction="column" gap="size-100">
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
                  // description="Enter your Zitadel username"
                  />
                )}
              />
              <Controller
                name="password"
                control={form.control}
                rules={{ required: 'Password is required' }}
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
              <Flex justifyContent="space-between" alignItems="center" marginTop="size-150">
                <Text>
                  Don't have an account?{" "}
                  <Link href="/signup">Sign up</Link>
                </Text>
                <Button
                  variant="cta"
                  type="submit"
                >
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
