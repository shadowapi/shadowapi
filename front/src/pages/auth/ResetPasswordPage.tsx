import { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router';
import { Card, Form, Input, Button, Alert, Typography, Spin, Result } from 'antd';
import { LockOutlined } from '@ant-design/icons';
import client from '../../api/client';

const { Title, Text } = Typography;

interface ResetInfo {
  email: string;
  expires_at: string;
}

function ResetPasswordPage() {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm();

  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [resetInfo, setResetInfo] = useState<ResetInfo | null>(null);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    if (!token) {
      setError('Invalid reset link');
      setLoading(false);
      return;
    }

    const validateToken = async () => {
      const { data, error: fetchError } = await client.GET('/password/reset/{token}', {
        params: { path: { token } },
      });

      if (fetchError) {
        setError('This reset link is invalid or has expired.');
        setLoading(false);
        return;
      }

      setResetInfo(data as ResetInfo);
      setLoading(false);
    };

    validateToken();
  }, [token]);

  const handleSubmit = async (values: { new_password: string }) => {
    if (!token) return;

    setSubmitting(true);
    setError(null);

    const { error: submitError } = await client.POST('/password/reset/confirm', {
      body: {
        token,
        new_password: values.new_password,
      },
    });

    setSubmitting(false);

    if (submitError) {
      setError(submitError.detail || 'Failed to reset password');
      return;
    }

    setSuccess(true);
    setTimeout(() => navigate('/login'), 3000);
  };

  if (loading) {
    return (
      <Card style={{ width: 450, maxWidth: '100%' }}>
        <div style={{ textAlign: 'center', padding: 40 }}>
          <Spin size="large" />
          <Text style={{ display: 'block', marginTop: 16 }}>Validating reset link...</Text>
        </div>
      </Card>
    );
  }

  if (error && !resetInfo) {
    return (
      <Card style={{ width: 450, maxWidth: '100%' }}>
        <Result
          status="error"
          title="Invalid Reset Link"
          subTitle={error}
          extra={
            <Link to="/forgot-password">
              <Button type="primary">Request New Link</Button>
            </Link>
          }
        />
      </Card>
    );
  }

  if (success) {
    return (
      <Card style={{ width: 450, maxWidth: '100%' }}>
        <Result
          status="success"
          title="Password Reset!"
          subTitle="Your password has been changed. Redirecting to login..."
        />
      </Card>
    );
  }

  return (
    <Card style={{ width: 450, maxWidth: '100%' }}>
      <div style={{ textAlign: 'center', marginBottom: 24 }}>
        <img src="/logo.svg" alt="MeshPump logo" style={{ height: 48, marginBottom: 16 }} />
        <Title level={3} style={{ margin: 0 }}>Reset Password</Title>
        <Text type="secondary">
          Create a new password for {resetInfo?.email}
        </Text>
      </div>

      {error && (
        <Alert
          message={error}
          type="error"
          showIcon
          closable
          onClose={() => setError(null)}
          style={{ marginBottom: 24 }}
        />
      )}

      <Form form={form} layout="vertical" onFinish={handleSubmit}>
        <Form.Item
          name="new_password"
          label="New Password"
          rules={[
            { required: true, message: 'Please enter a new password' },
            { min: 8, message: 'Password must be at least 8 characters' },
          ]}
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder="Enter new password"
            size="large"
          />
        </Form.Item>

        <Form.Item
          name="confirm_password"
          label="Confirm Password"
          dependencies={['new_password']}
          rules={[
            { required: true, message: 'Please confirm your password' },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('new_password') === value) {
                  return Promise.resolve();
                }
                return Promise.reject(new Error('Passwords do not match'));
              },
            }),
          ]}
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder="Confirm new password"
            size="large"
          />
        </Form.Item>

        <Form.Item style={{ marginBottom: 0 }}>
          <Button type="primary" htmlType="submit" size="large" loading={submitting} block>
            Reset Password
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
}

export default ResetPasswordPage;
