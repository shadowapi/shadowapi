import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router';
import {
  Form,
  Input,
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

type UsageLimit = components['schemas']['usage_limit'];

const limitTypeOptions = [
  { label: 'Messages Fetch', value: 'messages_fetch' },
  { label: 'Messages Push', value: 'messages_push' },
];

const resetPeriodOptions = [
  { label: 'Daily', value: 'daily' },
  { label: 'Weekly', value: 'weekly' },
  { label: 'Monthly', value: 'monthly' },
  { label: 'Rolling 24 Hours', value: 'rolling_24h' },
  { label: 'Rolling 7 Days', value: 'rolling_7d' },
  { label: 'Rolling 30 Days', value: 'rolling_30d' },
];

function UsageLimitEdit() {
  const navigate = useNavigate();
  const { uuid } = useParams<{ uuid: string }>();
  const { slug } = useWorkspace();
  const { user: currentUser } = useAuth();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [unlimited, setUnlimited] = useState(false);

  const isNew = !uuid;

  useEffect(() => {
    if (!isNew && isAdmin(currentUser)) {
      setLoading(true);
      client
        .GET('/access/usage-limits/{uuid}', {
          params: { path: { uuid } },
        })
        .then(({ data, error }) => {
          if (error) {
            message.error('Failed to load usage limit');
            navigate(`/w/${slug}/access/usage-limits`);
            return;
          }
          if (data) {
            form.setFieldsValue(data);
            setUnlimited(data.limit_value === null);
          }
        })
        .finally(() => setLoading(false));
    }
  }, [uuid, isNew, currentUser, form, navigate, slug]);

  const handleSubmit = async (values: Partial<UsageLimit>) => {
    setSaving(true);

    const payload: Partial<UsageLimit> = {
      ...values,
      limit_value: unlimited ? null : values.limit_value,
    };

    if (isNew) {
      const { error } = await client.POST('/access/usage-limits', {
        body: payload as UsageLimit,
      });
      if (error) {
        message.error((error as { detail?: string }).detail || 'Failed to create usage limit');
        setSaving(false);
        return;
      }
      message.success('Usage limit created');
    } else {
      const { error } = await client.PUT('/access/usage-limits/{uuid}', {
        params: { path: { uuid: uuid! } },
        body: payload as UsageLimit,
      });
      if (error) {
        message.error((error as { detail?: string }).detail || 'Failed to update usage limit');
        setSaving(false);
        return;
      }
      message.success('Usage limit updated');
    }

    setSaving(false);
    navigate(`/w/${slug}/access/usage-limits`);
  };

  const handleDelete = async () => {
    setDeleting(true);
    const { error } = await client.DELETE('/access/usage-limits/{uuid}', {
      params: { path: { uuid: uuid! } },
    });
    if (error) {
      message.error('Failed to delete usage limit');
      setDeleting(false);
      return;
    }
    message.success('Usage limit deleted');
    navigate(`/w/${slug}/access/usage-limits`);
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

  return (
    <>
      <Space style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Title level={4} style={{ margin: 0 }}>
          {isNew ? 'Create Usage Limit' : 'Edit Usage Limit'}
        </Title>
        {!isNew && (
          <Popconfirm
            title="Delete usage limit"
            description="Are you sure you want to delete this usage limit?"
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
            reset_period: 'monthly',
            is_enabled: true,
            limit_value: 1000,
          }}
          style={{ maxWidth: 600 }}
        >
          <Form.Item
            name="policy_set_name"
            label="Policy Set Name"
            rules={[{ required: true, message: 'Please enter policy set name' }]}
            extra="The policy set this limit applies to (e.g., workspace_member)"
          >
            <Input placeholder="e.g., workspace_member" />
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
                    placeholder="Maximum allowed count per period"
                  />
                </Form.Item>
              )}
            </Space>
          </Form.Item>

          <Form.Item
            name="reset_period"
            label="Reset Period"
            rules={[{ required: true, message: 'Please select reset period' }]}
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
              <Button onClick={() => navigate(`/w/${slug}/access/usage-limits`)}>Cancel</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </>
  );
}

export default UsageLimitEdit;
