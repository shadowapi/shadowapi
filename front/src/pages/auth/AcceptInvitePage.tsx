import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router';
import { Card, Form, Input, Button, Alert, Typography, Spin, Result } from 'antd';
import { LockOutlined, UserOutlined } from '@ant-design/icons';
import client from '../../api/client';

const { Title, Text } = Typography;

interface InviteInfo {
  email: string;
  workspace_name: string;
  workspace_slug: string;
  role: string;
  expires_at: string;
  inviter_name?: string;
}

interface AcceptFormValues {
  password: string;
  confirm_password: string;
  first_name: string;
  last_name: string;
}

function AcceptInvitePage() {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm();

  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [inviteInfo, setInviteInfo] = useState<InviteInfo | null>(null);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    if (!token) {
      setError('Invalid invite link');
      setLoading(false);
      return;
    }

    const fetchInvite = async () => {
      const { data, error: fetchError } = await client.GET('/invite/{token}', {
        params: { path: { token } },
      });

      if (fetchError) {
        setError('This invite link is invalid or has expired.');
        setLoading(false);
        return;
      }

      setInviteInfo(data as InviteInfo);
      setLoading(false);
    };

    fetchInvite();
  }, [token]);

  const handleSubmit = async (values: AcceptFormValues) => {
    if (!token) return;

    setSubmitting(true);
    setError(null);

    const { error: submitError } = await client.POST('/invite/accept', {
      body: {
        token,
        password: values.password,
        first_name: values.first_name,
        last_name: values.last_name,
      },
    });

    setSubmitting(false);

    if (submitError) {
      setError(submitError.detail || 'Failed to accept invite');
      return;
    }

    setSuccess(true);

    // Redirect to login after 3 seconds
    setTimeout(() => {
      navigate('/login');
    }, 3000);
  };

  if (loading) {
    return (
      <Card style={{ width: 450, maxWidth: '100%' }}>
        <div style={{ textAlign: 'center', padding: 40 }}>
          <Spin size="large" />
          <Text style={{ display: 'block', marginTop: 16 }}>Loading invite...</Text>
        </div>
      </Card>
    );
  }

  if (error && !inviteInfo) {
    return (
      <Card style={{ width: 450, maxWidth: '100%' }}>
        <Result
          status="error"
          title="Invalid Invite"
          subTitle={error}
          extra={
            <Button type="primary" onClick={() => navigate('/login')}>
              Go to Login
            </Button>
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
          title="Account Created!"
          subTitle={`You've been added to ${inviteInfo?.workspace_name}. Redirecting to login...`}
        />
      </Card>
    );
  }

  return (
    <Card style={{ width: 450, maxWidth: '100%' }}>
      <div style={{ textAlign: 'center', marginBottom: 24 }}>
        <img src="/logo.svg" alt="MeshPump logo" style={{ height: 48, marginBottom: 16 }} />
        <Title level={3} style={{ margin: 0 }}>Join {inviteInfo?.workspace_name}</Title>
        <Text type="secondary">
          {inviteInfo?.inviter_name
            ? `${inviteInfo.inviter_name} invited you to join as ${inviteInfo.role}`
            : `You've been invited to join as ${inviteInfo?.role}`
          }
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
        <div style={{ background: '#f5f5f5', padding: 12, borderRadius: 6, marginBottom: 16 }}>
          <Text strong>Email: </Text>
          <Text>{inviteInfo?.email}</Text>
        </div>

        <Form.Item
          name="first_name"
          label="First Name"
          rules={[{ required: true, message: 'Please enter your first name' }]}
        >
          <Input prefix={<UserOutlined />} placeholder="First name" size="large" />
        </Form.Item>

        <Form.Item
          name="last_name"
          label="Last Name"
          rules={[{ required: true, message: 'Please enter your last name' }]}
        >
          <Input prefix={<UserOutlined />} placeholder="Last name" size="large" />
        </Form.Item>

        <Form.Item
          name="password"
          label="Password"
          rules={[
            { required: true, message: 'Please create a password' },
            { min: 8, message: 'Password must be at least 8 characters' },
          ]}
        >
          <Input.Password prefix={<LockOutlined />} placeholder="Create password" size="large" />
        </Form.Item>

        <Form.Item
          name="confirm_password"
          label="Confirm Password"
          dependencies={['password']}
          rules={[
            { required: true, message: 'Please confirm your password' },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('password') === value) {
                  return Promise.resolve();
                }
                return Promise.reject(new Error('Passwords do not match'));
              },
            }),
          ]}
        >
          <Input.Password prefix={<LockOutlined />} placeholder="Confirm password" size="large" />
        </Form.Item>

        <Form.Item style={{ marginBottom: 0 }}>
          <Button type="primary" htmlType="submit" size="large" loading={submitting} block>
            Create Account & Join
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
}

export default AcceptInvitePage;
