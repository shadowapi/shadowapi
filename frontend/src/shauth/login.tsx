import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Card, Form, Input, Typography } from 'antd'
import { useSession } from './query'

const { Text, Title } = Typography

export function LoginPage() {
  const navigate = useNavigate()
  const { data: session, isLoading, error: sessionError } = useSession()
  const [submitting, setSubmitting] = useState(false)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  const disabledByAdmin = useMemo(() => {
    return !session?.active && document.cookie.split(';').some((c) => c.trim().startsWith('zitadel_access_token='))
  }, [session])

  useEffect(() => {
    if (session?.active) {
      navigate('/')
    }
  }, [session, navigate])

  if (isLoading) {
    return <span>Loading...</span>
  }

  if (sessionError) {
    return (
      <span>
        Error: {sessionError instanceof Error ? sessionError.message : 'An unexpected error occurred'}
      </span>
    )
  }

  const handleLogin = async (values: { email: string; password: string }) => {
    setSubmitting(true)
    setErrorMsg(null)
    try {
      const resp = await fetch('/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(values),
      })
      if (resp.ok) {
        navigate('/')
      } else {
        setErrorMsg('Invalid email or password')
      }
    } finally {
      setSubmitting(false)
    }
  }

  const zitadelLogin = () => {
    window.location.href = '/login/zitadel'
  }

  return (
    <div style={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center' }}>
      <Card style={{ width: 380 }}>
        <Title level={3}>Login</Title>

        <Form layout="vertical" onFinish={handleLogin} autoComplete="off">
          <Form.Item
            label="Email"
            name="email"
            rules={[{ required: true, message: 'Please enter your email' }]}
          >
            <Input type="email" placeholder="Email" />
          </Form.Item>

          <Form.Item
            label="Password"
            name="password"
            rules={[{ required: true, message: 'Please enter your password' }]}
          >
            <Input.Password placeholder="Password" />
          </Form.Item>

          {disabledByAdmin && (
            <Text type="danger">User is disabled, contact Admin</Text>
          )}

          {errorMsg && (
            <div style={{ marginBottom: 12 }}>
              <Text type="danger">{errorMsg}</Text>
            </div>
          )}

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={submitting} block>
              {submitting ? 'Logging in...' : 'Login'}
            </Button>
          </Form.Item>
        </Form>

        <Button type="default" block onClick={zitadelLogin}>
          Login with ZITADEL
        </Button>
      </Card>
    </div>
  )
}
