import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Flex, Header, Link, Text, View } from '@adobe/react-spectrum'
import { useQuery } from '@tanstack/react-query'
import { sessionOptions } from './query'

export function LoginPage() {
  const navigate = useNavigate()
  const session = useQuery(sessionOptions())
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  // Detect “Zitadel cookie present but no active session” → disabled account.
  const disabledByAdmin = useMemo(() => {
    return !session.data?.active && document.cookie.split(';').some((c) => c.trim().startsWith('zitadel_access_token='))
  }, [session.data])

  useEffect(() => {
    if (session.data?.active) {
      navigate('/')
    }
  }, [session.data, navigate])

  if (session.isPending) {
    return <span>Loading...</span>
  }
  if (session.isError) {
    return (
      <span>
        Error '{session.error.name}': {session.error.message}
      </span>
    )
  }

  const zitadelLogin = () => {
    window.location.href = '/login/zitadel'
  }

  return (
    <Flex direction="row" alignItems="center" justifyContent="center" flexBasis="100%" height="100vh">
      <View padding="size-200" backgroundColor="gray-200" borderRadius="medium" width="size-3600">
        <Flex direction="column" gap="size-100">
          <Header>Login</Header>
          <input type="email" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          {disabledByAdmin && <Text color="negative">User is disabled, contact Admin</Text>}

          {errorMsg && <Text color="negative">{errorMsg}</Text>}
          <Button
            variant="primary"
            alignSelf="end"
            width="size-1250"
            isDisabled={submitting}
            onPress={async () => {
              setSubmitting(true)
              setErrorMsg(null)
              const resp = await fetch('/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({ email, password }),
              })
              setSubmitting(false)
              if (resp.ok) {
                navigate('/')
              } else {
                setErrorMsg('Invalid email or password')
              }
            }}
          >
            {submitting ? 'Logging in...' : 'Login'}
          </Button>
          <Button variant="cta" alignSelf="end" marginTop="size-150" width="size-1250" onPress={zitadelLogin}>
            <span style={{ whiteSpace: 'nowrap' }}>Login with ZITADEL</span>
          </Button>
        </Flex>
      </View>
    </Flex>
  )
}
