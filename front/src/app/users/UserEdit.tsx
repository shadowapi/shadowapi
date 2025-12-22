import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router';
import {
  Form,
  Input,
  Button,
  Space,
  Typography,
  message,
  Checkbox,
  Popconfirm,
  Result,
  Spin,
} from 'antd';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import { useAuth } from '../../lib/auth';
import type { components } from '../../api/v1';

const { Title } = Typography;

type User = components['schemas']['user'];

interface UserFormValues {
  email: string;
  first_name: string;
  last_name: string;
  password?: string;
  is_enabled: boolean;
  is_admin: boolean;
}

function UserEdit() {
  const navigate = useNavigate();
  const { uuid } = useParams<{ uuid: string }>();
  const { slug } = useWorkspace();
  const { user: currentUser } = useAuth();
  const isNew = !uuid;
  const [form] = Form.useForm<UserFormValues>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

  // Admin access check
  if (!currentUser?.is_admin) {
    return (
      <Result
        status="403"
        title="Access Denied"
        subTitle="You need administrator privileges to access this page."
        extra={
          <Button type="primary" onClick={() => navigate(`/w/${slug}/`)}>
            Back to Dashboard
          </Button>
        }
      />
    );
  }

  useEffect(() => {
    if (!isNew && uuid) {
      const loadUser = async () => {
        setLoading(true);
        const { data, error } = await client.GET('/user/{uuid}', {
          params: { path: { uuid } },
        });
        if (error) {
          message.error('Failed to load user');
          navigate(`/w/${slug}/users`);
          return;
        }
        form.setFieldsValue({
          email: data.email,
          first_name: data.first_name,
          last_name: data.last_name,
          is_enabled: data.is_enabled ?? true,
          is_admin: data.is_admin ?? false,
        });
        setLoading(false);
      };
      loadUser();
    }
  }, [uuid, isNew, form, slug, navigate]);

  const onFinish = async (values: UserFormValues) => {
    setSaving(true);

    const userData: User = {
      email: values.email,
      first_name: values.first_name,
      last_name: values.last_name,
      password: values.password || '',
      is_enabled: values.is_enabled,
      is_admin: values.is_admin,
    };

    let result;
    if (isNew) {
      result = await client.POST('/user', {
        body: userData,
      });
    } else {
      result = await client.PUT('/user/{uuid}', {
        params: { path: { uuid: uuid! } },
        body: userData,
      });
    }

    if (result.error) {
      message.error(
        (result.error as { detail?: string }).detail ||
          `Failed to ${isNew ? 'create' : 'update'} user`
      );
      setSaving(false);
      return;
    }

    message.success(`User ${isNew ? 'created' : 'updated'} successfully`);
    navigate(`/w/${slug}/users`);
  };

  const onDelete = async () => {
    if (!uuid) return;
    setDeleting(true);
    const { error } = await client.DELETE('/user/{uuid}', {
      params: { path: { uuid } },
    });
    if (error) {
      message.error('Failed to delete user');
      setDeleting(false);
      return;
    }
    message.success('User deleted');
    navigate(`/w/${slug}/users`);
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <>
      <Title level={4}>{isNew ? 'Add' : 'Edit'} User</Title>
      <Form
        form={form}
        layout="vertical"
        onFinish={onFinish}
        initialValues={{
          is_enabled: true,
          is_admin: false,
        }}
        style={{ maxWidth: 500 }}
      >
        <Form.Item
          label="Email"
          name="email"
          rules={[
            { required: true, message: 'Please enter an email address' },
            { type: 'email', message: 'Please enter a valid email address' },
          ]}
        >
          <Input placeholder="user@example.com" />
        </Form.Item>

        <Form.Item
          label="First Name"
          name="first_name"
          rules={[{ required: true, message: 'Please enter a first name' }]}
        >
          <Input placeholder="John" />
        </Form.Item>

        <Form.Item
          label="Last Name"
          name="last_name"
          rules={[{ required: true, message: 'Please enter a last name' }]}
        >
          <Input placeholder="Doe" />
        </Form.Item>

        {isNew && (
          <Form.Item
            label="Password"
            name="password"
            rules={[
              { required: true, message: 'Please enter a password' },
              { min: 8, message: 'Password must be at least 8 characters' },
            ]}
          >
            <Input.Password placeholder="Enter password" />
          </Form.Item>
        )}

        <Form.Item name="is_enabled" valuePropName="checked">
          <Checkbox>User is enabled</Checkbox>
        </Form.Item>

        <Form.Item name="is_admin" valuePropName="checked">
          <Checkbox>User is administrator</Checkbox>
        </Form.Item>

        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={saving}>
              {isNew ? 'Create' : 'Update'}
            </Button>
            <Button onClick={() => navigate(`/w/${slug}/users`)}>Cancel</Button>
            {!isNew && uuid !== currentUser?.uuid && (
              <Popconfirm
                title="Delete User"
                description="Are you sure you want to delete this user? This action cannot be undone."
                onConfirm={onDelete}
                okButtonProps={{ danger: true }}
                okText="Delete"
              >
                <Button danger loading={deleting}>
                  Delete
                </Button>
              </Popconfirm>
            )}
          </Space>
        </Form.Item>
      </Form>
    </>
  );
}

export default UserEdit;
