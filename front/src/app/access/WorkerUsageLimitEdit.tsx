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
  Row,
  Col,
} from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import { useAuth, isAdmin } from '../../lib/auth';
import type { components } from '../../api/v1';

const { Title, Paragraph } = Typography;

type WorkerUsageLimit = components['schemas']['worker_usage_limit'];
type RegisteredWorker = components['schemas']['registered_worker'];

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

function WorkerUsageLimitEdit() {
  const navigate = useNavigate();
  const { uuid } = useParams<{ uuid: string }>();
  const [searchParams] = useSearchParams();
  const workerUuidFromQuery = searchParams.get('worker');
  const { slug } = useWorkspace();
  const { user: currentUser } = useAuth();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [unlimited, setUnlimited] = useState(false);
  const [workers, setWorkers] = useState<RegisteredWorker[]>([]);
  const [selectedWorkerUuid, setSelectedWorkerUuid] = useState<string | undefined>(workerUuidFromQuery || undefined);

  const isNew = !uuid;

  useEffect(() => {
    // Load workers for the dropdown
    client.GET('/workers').then(({ data }) => {
      if (data) {
        setWorkers(data);
      }
    });
  }, []);

  useEffect(() => {
    if (!isNew && isAdmin(currentUser) && workerUuidFromQuery) {
      setLoading(true);
      client
        .GET('/access/worker/{worker_uuid}/usage-limits', {
          params: {
            path: { worker_uuid: workerUuidFromQuery },
            query: { workspace_slug: slug },
          },
        })
        .then(({ data, error }) => {
          if (error) {
            message.error('Failed to load worker usage limit');
            navigate(`/w/${slug}/access/worker-usage-limits`);
            return;
          }
          const limit = data?.find((l) => l.uuid === uuid);
          if (limit) {
            form.setFieldsValue(limit);
            setSelectedWorkerUuid(limit.worker_uuid);
            setUnlimited(limit.limit_value === null);
          } else {
            message.error('Worker limit not found');
            navigate(`/w/${slug}/access/worker-usage-limits`);
          }
        })
        .finally(() => setLoading(false));
    }
  }, [uuid, isNew, currentUser, form, navigate, slug, workerUuidFromQuery]);

  const handleSubmit = async (values: Partial<WorkerUsageLimit>) => {
    if (!selectedWorkerUuid) {
      message.error('Please select a worker');
      return;
    }

    setSaving(true);

    const payload: Partial<WorkerUsageLimit> = {
      ...values,
      worker_uuid: selectedWorkerUuid,
      workspace_slug: slug,
      limit_value: unlimited ? null : values.limit_value,
    };

    if (isNew) {
      const { error } = await client.POST('/access/worker/{worker_uuid}/usage-limits', {
        params: { path: { worker_uuid: selectedWorkerUuid } },
        body: payload as WorkerUsageLimit,
      });
      if (error) {
        message.error((error as { detail?: string }).detail || 'Failed to create worker usage limit');
        setSaving(false);
        return;
      }
      message.success('Worker usage limit created');
    } else {
      const { error } = await client.PUT('/access/worker/{worker_uuid}/usage-limits/{uuid}', {
        params: { path: { worker_uuid: selectedWorkerUuid, uuid: uuid! } },
        body: payload as WorkerUsageLimit,
      });
      if (error) {
        message.error((error as { detail?: string }).detail || 'Failed to update worker usage limit');
        setSaving(false);
        return;
      }
      message.success('Worker usage limit updated');
    }

    setSaving(false);
    navigate(`/w/${slug}/access/worker-usage-limits`);
  };

  const handleDelete = async () => {
    if (!selectedWorkerUuid) return;

    setDeleting(true);
    const { error } = await client.DELETE('/access/worker/{worker_uuid}/usage-limits/{uuid}', {
      params: { path: { worker_uuid: selectedWorkerUuid, uuid: uuid! } },
    });
    if (error) {
      message.error('Failed to delete worker usage limit');
      setDeleting(false);
      return;
    }
    message.success('Worker usage limit deleted');
    navigate(`/w/${slug}/access/worker-usage-limits`);
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

  const workerOptions = workers.map((w) => ({
    label: w.name || w.uuid?.slice(0, 8) + '...',
    value: w.uuid,
  }));

  return (
    <>
      <Space style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Title level={4} style={{ margin: 0 }}>
          {isNew ? 'Create Worker Usage Limit' : 'Edit Worker Usage Limit'}
        </Title>
        {!isNew && (
          <Popconfirm
            title="Delete worker limit"
            description="Are you sure you want to delete this worker usage limit?"
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

      <Row gutter={24}>
        <Col xs={24} lg={16}>
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
            >
              <Form.Item
                label="Worker"
                rules={[{ required: true, message: 'Please select a worker' }]}
              >
                <Select
                  value={selectedWorkerUuid}
                  onChange={setSelectedWorkerUuid}
                  options={workerOptions}
                  placeholder="Select worker"
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
                  <Button onClick={() => navigate(`/w/${slug}/access/worker-usage-limits`)}>Cancel</Button>
                </Space>
              </Form.Item>
            </Form>
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title="About Worker Limits" size="small">
            <Paragraph>
              Worker limits control how many messages a specific worker can process,
              regardless of user policies. Use this to balance load across workers.
            </Paragraph>
            <Title level={5}>Worker</Title>
            <Paragraph type="secondary">
              The worker instance this limit applies to. Workers are identified by
              their registered name or UUID.
            </Paragraph>
            <Title level={5}>Limit Type</Title>
            <Paragraph type="secondary">
              <strong>Messages Fetch</strong> - limits incoming/received messages.
              <br />
              <strong>Messages Push</strong> - limits outgoing/sent messages.
            </Paragraph>
            <Title level={5}>Limit Value</Title>
            <Paragraph type="secondary">
              Maximum number of messages this worker can process per reset period.
              Toggle to "Unlimited" for no restrictions.
            </Paragraph>
            <Title level={5}>Reset Period</Title>
            <Paragraph type="secondary">
              When the usage counter resets. Fixed periods (daily, weekly, monthly) reset
              at midnight UTC. Rolling periods count from each message timestamp.
            </Paragraph>
            <Title level={5}>Enabled</Title>
            <Paragraph type="secondary">
              When disabled, this limit is ignored and the worker has no processing
              restrictions from this configuration.
            </Paragraph>
          </Card>
        </Col>
      </Row>
    </>
  );
}

export default WorkerUsageLimitEdit;
