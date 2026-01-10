import { useEffect, useState } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router';
import {
  Form,
  Button,
  Space,
  Typography,
  message,
  Result,
  Popconfirm,
  Select,
  InputNumber,
  Switch,
  Card,
  Spin,
} from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import { useAuth, isAdmin } from '../../lib/auth';
import type { components } from '../../api/v1';

const { Title } = Typography;

type UserUsageLimitOverride = components['schemas']['user_usage_limit_override'];
type User = components['schemas']['user'];

const limitTypeOptions = [
  { label: 'Messages Fetch', value: 'messages_fetch' },
  { label: 'Messages Push', value: 'messages_push' },
];

const resetPeriodOptions = [
  { label: 'Inherit from Policy Set', value: null },
  { label: 'Daily', value: 'daily' },
  { label: 'Weekly', value: 'weekly' },
  { label: 'Monthly', value: 'monthly' },
  { label: 'Rolling 24 Hours', value: 'rolling_24h' },
  { label: 'Rolling 7 Days', value: 'rolling_7d' },
  { label: 'Rolling 30 Days', value: 'rolling_30d' },
];

function UserUsageLimitEdit() {
  const navigate = useNavigate();
  const { uuid } = useParams<{ uuid: string }>();
  const [searchParams] = useSearchParams();
  const userUuidFromQuery = searchParams.get('user');
  const { slug } = useWorkspace();
  const { user: currentUser } = useAuth();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [unlimited, setUnlimited] = useState(false);
  const [users, setUsers] = useState<User[]>([]);
  const [selectedUserUuid, setSelectedUserUuid] = useState<string | undefined>(userUuidFromQuery || undefined);

  const isNew = !uuid;

  useEffect(() => {
    // Load users for the dropdown
    client.GET('/user').then(({ data }) => {
      if (data) {
        setUsers(data);
      }
    });
  }, []);

  useEffect(() => {
    if (!isNew && isAdmin(currentUser) && userUuidFromQuery) {
      setLoading(true);
      // Find the override by iterating through user's overrides
      client
        .GET('/access/user/{user_uuid}/usage-limits', {
          params: {
            path: { user_uuid: userUuidFromQuery },
            query: { workspace_slug: slug },
          },
        })
        .then(({ data, error }) => {
          if (error) {
            message.error('Failed to load user usage limit override');
            navigate(`/w/${slug}/access/user-usage-limits`);
            return;
          }
          const override = data?.find((o) => o.uuid === uuid);
          if (override) {
            form.setFieldsValue(override);
            setSelectedUserUuid(override.user_uuid);
            setUnlimited(override.limit_value === null);
          } else {
            message.error('Override not found');
            navigate(`/w/${slug}/access/user-usage-limits`);
          }
        })
        .finally(() => setLoading(false));
    }
  }, [uuid, isNew, currentUser, form, navigate, slug, userUuidFromQuery]);

  const handleSubmit = async (values: Partial<UserUsageLimitOverride>) => {
    if (!selectedUserUuid) {
      message.error('Please select a user');
      return;
    }

    setSaving(true);

    const payload: Partial<UserUsageLimitOverride> = {
      ...values,
      user_uuid: selectedUserUuid,
      workspace_slug: slug,
      limit_value: unlimited ? null : values.limit_value,
    };

    if (isNew) {
      const { error } = await client.POST('/access/user/{user_uuid}/usage-limits', {
        params: { path: { user_uuid: selectedUserUuid } },
        body: payload as UserUsageLimitOverride,
      });
      if (error) {
        message.error((error as { detail?: string }).detail || 'Failed to create user usage limit override');
        setSaving(false);
        return;
      }
      message.success('User usage limit override created');
    } else {
      const { error } = await client.PUT('/access/user/{user_uuid}/usage-limits/{uuid}', {
        params: { path: { user_uuid: selectedUserUuid, uuid: uuid! } },
        body: payload as UserUsageLimitOverride,
      });
      if (error) {
        message.error((error as { detail?: string }).detail || 'Failed to update user usage limit override');
        setSaving(false);
        return;
      }
      message.success('User usage limit override updated');
    }

    setSaving(false);
    navigate(`/w/${slug}/access/user-usage-limits`);
  };

  const handleDelete = async () => {
    if (!selectedUserUuid) return;

    setDeleting(true);
    const { error } = await client.DELETE('/access/user/{user_uuid}/usage-limits/{uuid}', {
      params: { path: { user_uuid: selectedUserUuid, uuid: uuid! } },
    });
    if (error) {
      message.error('Failed to delete user usage limit override');
      setDeleting(false);
      return;
    }
    message.success('User usage limit override deleted');
    navigate(`/w/${slug}/access/user-usage-limits`);
  };

  // Admin access check - after hooks
  if (!isAdmin(currentUser)) {
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

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 50 }}>
        <Spin size="large" />
      </div>
    );
  }

  const userOptions = users.map((u) => ({
    label: u.email,
    value: u.uuid,
  }));

  return (
    <>
      <Space style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Title level={4} style={{ margin: 0 }}>
          {isNew ? 'Create User Usage Override' : 'Edit User Usage Override'}
        </Title>
        {!isNew && (
          <Popconfirm
            title="Delete override"
            description="Are you sure you want to delete this user usage limit override?"
            onConfirm={handleDelete}
            okButtonProps={{ danger: true, loading: deleting }}
            okText="Delete"
          >
            <Button danger icon={<DeleteOutlined />}>
              Delete
            </Button>
          </Popconfirm>
        )}
      </Space>

      <Card>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            limit_type: 'messages_fetch',
            reset_period: null,
            is_enabled: true,
            limit_value: 1000,
          }}
          style={{ maxWidth: 600 }}
        >
          <Form.Item
            label="User"
            rules={[{ required: true, message: 'Please select a user' }]}
          >
            <Select
              value={selectedUserUuid}
              onChange={setSelectedUserUuid}
              options={userOptions}
              placeholder="Select user"
              showSearch
              filterOption={(input, option) =>
                (option?.label as string)?.toLowerCase().includes(input.toLowerCase())
              }
              disabled={!isNew}
            />
          </Form.Item>

          <Form.Item
            name="limit_type"
            label="Limit Type"
            rules={[{ required: true, message: 'Please select limit type' }]}
          >
            <Select options={limitTypeOptions} />
          </Form.Item>

          <Form.Item label="Limit Value">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Switch
                checked={unlimited}
                onChange={setUnlimited}
                checkedChildren="Unlimited"
                unCheckedChildren="Limited"
              />
              {!unlimited && (
                <Form.Item name="limit_value" noStyle>
                  <InputNumber
                    min={0}
                    style={{ width: '100%' }}
                    placeholder="Override limit value"
                  />
                </Form.Item>
              )}
            </Space>
          </Form.Item>

          <Form.Item
            name="reset_period"
            label="Reset Period"
            extra="Leave as 'Inherit' to use the policy set's reset period"
          >
            <Select options={resetPeriodOptions} />
          </Form.Item>

          <Form.Item name="is_enabled" label="Enabled" valuePropName="checked">
            <Switch />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={saving}>
                {isNew ? 'Create' : 'Update'}
              </Button>
              <Button onClick={() => navigate(`/w/${slug}/access/user-usage-limits`)}>Cancel</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </>
  );
}

export default UserUsageLimitEdit;
