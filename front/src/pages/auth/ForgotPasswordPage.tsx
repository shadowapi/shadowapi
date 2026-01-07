import { useState } from 'react';
import { Link } from 'react-router';
import { Card, Form, Input, Button, Alert, Typography, Result } from 'antd';
import { MailOutlined, ArrowLeftOutlined } from '@ant-design/icons';
import client from '../../api/client';

const { Title, Text } = Typography;

function ForgotPasswordPage() {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (values: { email: string }) => {
    setLoading(true);
    setError(null);

    const { error: submitError } = await client.POST('/password/reset', {
      body: { email: values.email },
    });

    setLoading(false);

    if (submitError) {
      // API returns error_message field
      const errorMessage = (submitError as { error_message?: string }).error_message;
      if (submitError.status === 429) {
        setError(errorMessage || 'Please wait before requesting another reset');
      } else {
        setError(errorMessage || 'Failed to send reset email. Please try again.');
      }
      return;
    }

    setSuccess(true);
  };

  if (success) {
    return (
      <Card style={{ width: 400, maxWidth: '100%' }}>
        <Result
          status="success"
          title="Check Your Email"
          subTitle="If an account exists with that email, we've sent a password reset link."
          extra={
            <Link to="/login">
              <Button type="primary">Back to Login</Button>
            </Link>
          }
        />
      </Card>
    );
  }

  return (
    <Card style={{ width: 400, maxWidth: '100%' }}>
      <div style={{ textAlign: 'center', marginBottom: 24 }}>
        <img src="/logo.svg" alt="MeshPump logo" style={{ height: 48, marginBottom: 16 }} />
        <Title level={3} style={{ margin: 0 }}>Forgot Password</Title>
        <Text type="secondary">
          Enter your email to receive a reset link
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
          name="email"
          rules={[
            { required: true, message: 'Please enter your email' },
            { type: 'email', message: 'Please enter a valid email' },
          ]}
        >
          <Input
            prefix={<MailOutlined />}
            placeholder="Email address"
            size="large"
            autoComplete="email"
          />
        </Form.Item>

        <Form.Item style={{ marginBottom: 16 }}>
          <Button type="primary" htmlType="submit" size="large" loading={loading} block>
            Send Reset Link
          </Button>
        </Form.Item>

        <div style={{ textAlign: 'center' }}>
          <Link to="/login">
            <ArrowLeftOutlined /> Back to Login
          </Link>
        </div>
      </Form>
    </Card>
  );
}

export default ForgotPasswordPage;
