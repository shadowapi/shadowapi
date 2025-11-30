import { useEffect, useRef } from 'react';
import { useNavigate, useLocation, useSearchParams } from 'react-router';
import { Card, Form, Input, Button, Alert, Typography, Spin } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useAuth } from '../../lib/auth';

const { Title } = Typography;

interface LoginFormValues {
  email: string;
  password: string;
}

function LoginPage() {
  const { login, isAuthenticated, isLoading, error, clearError } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const [form] = Form.useForm();
  const redirectInitiated = useRef(false);

  // Get login_challenge from URL if present (OAuth2 flow from Hydra)
  const loginChallenge = searchParams.get('login_challenge');

  const from = (location.state as { from?: { pathname: string } })?.from
    ?.pathname || '/app';

  useEffect(() => {
    // Only redirect if authenticated AND there's no login_challenge
    // (when there's a login_challenge, we need to complete the OAuth2 flow)
    if (isAuthenticated && !loginChallenge) {
      navigate(from, { replace: true });
    }
  }, [isAuthenticated, navigate, from, loginChallenge]);

  // Auto-initiate OAuth2 flow when no login_challenge is present
  useEffect(() => {
    if (!loginChallenge && !isAuthenticated && !isLoading && !redirectInitiated.current) {
      redirectInitiated.current = true;
      // Calling login with empty credentials initiates the OAuth2 flow
      login('', '', undefined).catch(() => {
        // Reset flag if redirect fails so user can retry
        redirectInitiated.current = false;
      });
    }
  }, [loginChallenge, isAuthenticated, isLoading, login]);

  const handleSubmit = async (values: LoginFormValues) => {
    clearError();
    try {
      await login(values.email, values.password, loginChallenge || undefined);
    } catch {
      // Error is handled by auth context
    }
  };

  // Show loading state when no login_challenge (redirecting to OAuth2 flow)
  if (!loginChallenge) {
    return (
      <Card style={{ width: 400, maxWidth: '100%' }}>
        <div style={{ textAlign: 'center', padding: 24 }}>
          <Spin size="large" />
          <Typography.Text style={{ display: 'block', marginTop: 16 }}>
            Redirecting to login...
          </Typography.Text>
          {error && (
            <Alert
              message={error}
              type="error"
              showIcon
              style={{ marginTop: 16 }}
            />
          )}
        </div>
      </Card>
    );
  }

  return (
    <Card style={{ width: 400, maxWidth: '100%' }}>
      <div style={{ textAlign: 'center', marginBottom: 24 }}>
        <Title level={3} style={{ margin: 0 }}>
          ShadowAPI
        </Title>
        <Typography.Text type="secondary">
          Sign in to your account
        </Typography.Text>
      </div>

      {error && (
        <Alert
          message={error}
          type="error"
          showIcon
          closable
          onClose={clearError}
          style={{ marginBottom: 24 }}
        />
      )}

      <Form
        form={form}
        name="login"
        onFinish={handleSubmit}
        layout="vertical"
        requiredMark={false}
      >
        <Form.Item
          name="email"
          rules={[
            { required: true, message: 'Please enter your email' },
            { type: 'email', message: 'Please enter a valid email' },
          ]}
        >
          <Input
            prefix={<UserOutlined />}
            placeholder="Email"
            size="large"
            autoComplete="email"
          />
        </Form.Item>

        <Form.Item
          name="password"
          rules={[{ required: true, message: 'Please enter your password' }]}
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder="Password"
            size="large"
            autoComplete="current-password"
          />
        </Form.Item>

        <Form.Item style={{ marginBottom: 0 }}>
          <Button
            type="primary"
            htmlType="submit"
            size="large"
            loading={isLoading}
            block
          >
            Sign in
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
}

export default LoginPage;
