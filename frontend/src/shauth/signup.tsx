import { Link } from 'react-router-dom'
import { Button, Card, Typography } from 'antd'

const { Title } = Typography

export function SignupPage() {
  const signupUrl = `${import.meta.env.VITE_ZITADEL_INSTANCE_URL}/oauth/v2/authorize?client_id=${import.meta.env.VITE_ZITADEL_CLIENT_ID}&response_type=code&scope=openid&redirect_uri=${encodeURIComponent(import.meta.env.VITE_ZITADEL_REDIRECT_URI)}`

  return (
    <div style={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center' }}>
      <Card style={{ width: 380 }}>
        <Title level={3}>Sign Up</Title>

        <Button
          type="primary"
          block
          onClick={() => {
            window.location.href = signupUrl
          }}
        >
          Sign up with ZITADEL
        </Button>

        <div style={{ marginTop: 16, textAlign: 'right' }}>
          <Link to="/login">Back to login</Link>
        </div>
      </Card>
    </div>
  )
}
