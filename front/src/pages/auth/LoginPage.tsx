import { useEffect, useRef } from "react";
import { useNavigate, useLocation, useSearchParams, Link } from "react-router";
import { Card, Form, Input, Button, Alert, Typography, Spin } from "antd";
import { UserOutlined, LockOutlined, LoadingOutlined } from "@ant-design/icons";
import { useAuth } from "../../lib/auth";

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
  const loginChallenge = searchParams.get("login_challenge");

  const from =
    (location.state as { from?: { pathname: string } })?.from?.pathname ||
    "/workspaces";

  useEffect(() => {
    // Only redirect if authenticated AND there's no login_challenge
    // (when there's a login_challenge, we need to complete the OAuth2 flow)
    if (isAuthenticated && !loginChallenge) {
      navigate(from, { replace: true });
    }
  }, [isAuthenticated, navigate, from, loginChallenge]);

  // Auto-initiate OAuth2 flow when no login_challenge is present
  useEffect(() => {
    if (
      !loginChallenge &&
      !isAuthenticated &&
      !isLoading &&
      !redirectInitiated.current
    ) {
      redirectInitiated.current = true;
      // Calling login with empty credentials initiates the OAuth2 flow
      login("", "", undefined).catch(() => {
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

  // Show spinner when no login_challenge (redirecting to OAuth2 flow)
  // Only show error if OAuth2 initiation fails
  if (!loginChallenge) {
    if (error) {
      return (
        <Card style={{ width: 400, maxWidth: "100%" }}>
          <Alert message={error} type="error" showIcon />
        </Card>
      );
    }
    return (
      <Card style={{ width: 400, maxWidth: "100%" }}>
        <div style={{ textAlign: "center", padding: "40px 0" }}>
          <Spin indicator={<LoadingOutlined style={{ fontSize: 32 }} spin />} />
          <div style={{ marginTop: 16 }}>
            <Typography.Text type="secondary">
              Redirecting to login...
            </Typography.Text>
          </div>
        </div>
      </Card>
    );
  }

  return (
    <Card style={{ width: 400, maxWidth: "100%" }}>
      <div style={{ textAlign: "center", marginBottom: 24 }}>
        <Link to="/start">
          <img
            src="/logo.svg"
            alt="MeshPump logo"
            style={{ height: 64, marginBottom: 16 }}
          />
        </Link>
        <Title level={3} style={{ margin: 0 }}>
          MeshPump
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
            { required: true, message: "Please enter your email" },
            { type: "email", message: "Please enter a valid email" },
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
          rules={[{ required: true, message: "Please enter your password" }]}
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder="Password"
            size="large"
            autoComplete="current-password"
          />
        </Form.Item>

        <Form.Item style={{ marginBottom: 16 }}>
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

        <div style={{ textAlign: "center" }}>
          <Link to="/forgot-password">Forgot password?</Link>
        </div>
      </Form>
    </Card>
  );
}

export default LoginPage;
