import { useState, useEffect, useCallback, useRef } from 'react';
import {
  Tabs,
  Typography,
  Space,
  Button,
  Table,
  Tag,
  message,
  Modal,
  Form,
  Input,
  DatePicker,
  Checkbox,
  Popconfirm,
  Tooltip,
  Alert,
} from 'antd';
import {
  PlusOutlined,
  DeleteOutlined,
  CopyOutlined,
  ReloadOutlined,
  StopOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  SyncOutlined,
  RobotOutlined,
  ThunderboltOutlined,
  KeyOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import client from '../../api/client';
import type { components } from '../../api/v1';

const { Title } = Typography;

type RegisteredWorker = components['schemas']['registered_worker'];
type WorkerEnrollmentToken = components['schemas']['worker_enrollment_token'];
type WorkerJob = components['schemas']['worker_jobs'];

const POLLING_INTERVAL = 10000; // 10 seconds

// Helper: Format relative time
function formatTimeAgo(dateStr?: string): string {
  if (!dateStr) return '-';
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `${diffHours}h ago`;
  const diffDays = Math.floor(diffHours / 24);
  return `${diffDays}d ago`;
}

// Helper: Calculate job duration
function calculateDuration(startedAt?: string, finishedAt?: string): string {
  if (!startedAt) return '-';
  const start = new Date(startedAt);
  const end = finishedAt ? new Date(finishedAt) : new Date();
  const diffMs = end.getTime() - start.getTime();

  if (diffMs < 1000) return '<1s';
  if (diffMs < 60000) return `${Math.floor(diffMs / 1000)}s`;
  if (diffMs < 3600000) return `${Math.floor(diffMs / 60000)}m ${Math.floor((diffMs % 60000) / 1000)}s`;
  return `${Math.floor(diffMs / 3600000)}h ${Math.floor((diffMs % 3600000) / 60000)}m`;
}

// Helper: Determine token status
function getTokenStatus(token: WorkerEnrollmentToken): { status: string; color: string } {
  if (token.used_at) return { status: 'used', color: 'default' };
  if (token.expires_at && new Date(token.expires_at) < new Date()) return { status: 'expired', color: 'error' };
  return { status: 'available', color: 'success' };
}

interface CreateTokenFormValues {
  name: string;
  is_global: boolean;
  expires_at?: dayjs.Dayjs;
}

function Workers() {
  // Tab state
  const [activeTab, setActiveTab] = useState('external-workers');

  // External workers state
  const [workers, setWorkers] = useState<RegisteredWorker[]>([]);
  const [workersLoading, setWorkersLoading] = useState(true);

  // Internal jobs state
  const [jobs, setJobs] = useState<WorkerJob[]>([]);
  const [jobsLoading, setJobsLoading] = useState(true);

  // Enrollment tokens state
  const [tokens, setTokens] = useState<WorkerEnrollmentToken[]>([]);
  const [tokensLoading, setTokensLoading] = useState(true);

  // Token creation modal state
  const [createTokenModalOpen, setCreateTokenModalOpen] = useState(false);
  const [createdToken, setCreatedToken] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [form] = Form.useForm<CreateTokenFormValues>();

  // Polling control
  const pollingPausedRef = useRef(false);

  // Load external workers
  const loadWorkers = useCallback(async () => {
    setWorkersLoading(true);
    const { data, error } = await client.GET('/workers');
    if (error) {
      message.error('Failed to load workers');
      setWorkersLoading(false);
      return;
    }
    setWorkers(data || []);
    setWorkersLoading(false);
  }, []);

  // Load internal jobs
  const loadJobs = useCallback(async () => {
    setJobsLoading(true);
    const { data, error } = await client.GET('/workerjobs', {
      params: { query: { offset: 0, limit: 100 } },
    });
    if (error) {
      message.error('Failed to load jobs');
      setJobsLoading(false);
      return;
    }
    setJobs(data?.jobs || []);
    setJobsLoading(false);
  }, []);

  // Load enrollment tokens
  const loadTokens = useCallback(async () => {
    setTokensLoading(true);
    const { data, error } = await client.GET('/workers/enrollment-tokens');
    if (error) {
      message.error('Failed to load tokens');
      setTokensLoading(false);
      return;
    }
    setTokens(data || []);
    setTokensLoading(false);
  }, []);

  // Delete worker
  const handleDeleteWorker = async (uuid: string) => {
    const { error } = await client.DELETE('/workers/{uuid}', {
      params: { path: { uuid } },
    });
    if (error) {
      message.error('Failed to delete worker');
      return;
    }
    message.success('Worker deleted');
    loadWorkers();
  };

  // Cancel job
  const handleCancelJob = async (uuid: string) => {
    const { error } = await client.POST('/workerjobs/{uuid}/cancel', {
      params: { path: { uuid } },
    });
    if (error) {
      message.error('Failed to cancel job');
      return;
    }
    message.success('Job cancellation requested');
    loadJobs();
  };

  // Delete job
  const handleDeleteJob = async (uuid: string) => {
    const { error } = await client.DELETE('/workerjobs/{uuid}', {
      params: { path: { uuid } },
    });
    if (error) {
      message.error('Failed to delete job');
      return;
    }
    message.success('Job deleted');
    loadJobs();
  };

  // Revoke token
  const handleRevokeToken = async (uuid: string) => {
    const { error } = await client.DELETE('/workers/enrollment-tokens/{uuid}', {
      params: { path: { uuid } },
    });
    if (error) {
      message.error('Failed to revoke token');
      return;
    }
    message.success('Token revoked');
    loadTokens();
  };

  // Create token
  const handleCreateToken = async () => {
    try {
      const values = await form.validateFields();
      setCreating(true);

      const { data, error } = await client.POST('/workers/enrollment-tokens', {
        body: {
          name: values.name,
          is_global: values.is_global || false,
          expires_at: values.expires_at?.toISOString(),
        },
      });

      if (error) {
        message.error('Failed to create token');
        setCreating(false);
        return;
      }

      // Store the one-time token for display
      setCreatedToken(data?.token || null);
      message.success('Token created successfully');

      // Refresh token list
      loadTokens();
    } finally {
      setCreating(false);
    }
  };

  // Close token modal
  const handleCloseModal = () => {
    setCreateTokenModalOpen(false);
    setCreatedToken(null);
    form.resetFields();
  };

  // Copy token to clipboard
  const handleCopyToken = () => {
    if (createdToken) {
      navigator.clipboard.writeText(createdToken);
      message.success('Token copied to clipboard');
    }
  };

  // Manual refresh handler
  const handleRefresh = () => {
    switch (activeTab) {
      case 'external-workers':
        loadWorkers();
        break;
      case 'internal-jobs':
        loadJobs();
        break;
      case 'enrollment-tokens':
        loadTokens();
        break;
    }
  };

  // Initial load and polling
  useEffect(() => {
    const poll = async () => {
      if (pollingPausedRef.current) return;

      switch (activeTab) {
        case 'external-workers':
          loadWorkers();
          break;
        case 'internal-jobs':
          loadJobs();
          break;
        case 'enrollment-tokens':
          loadTokens();
          break;
      }
    };

    // Initial load
    poll();

    // Set up interval
    const intervalId = setInterval(poll, POLLING_INTERVAL);

    return () => clearInterval(intervalId);
  }, [activeTab, loadWorkers, loadJobs, loadTokens]);

  // Pause polling when tab is not visible
  useEffect(() => {
    const handleVisibilityChange = () => {
      pollingPausedRef.current = document.hidden;
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange);
  }, []);

  // Pause polling when modal is open
  useEffect(() => {
    pollingPausedRef.current = createTokenModalOpen;
  }, [createTokenModalOpen]);

  // External workers columns
  const workersColumns: ColumnsType<RegisteredWorker> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const config: Record<string, { color: string; icon: React.ReactNode }> = {
          online: { color: 'success', icon: <CheckCircleOutlined /> },
          offline: { color: 'error', icon: <CloseCircleOutlined /> },
          draining: { color: 'warning', icon: <SyncOutlined spin /> },
        };
        const statusConfig = config[status] || { color: 'default', icon: null };
        return (
          <Tag color={statusConfig.color} icon={statusConfig.icon}>
            {status}
          </Tag>
        );
      },
    },
    {
      title: 'Version',
      dataIndex: 'version',
      key: 'version',
      render: (v: string) => v || '-',
    },
    {
      title: 'Connected From',
      dataIndex: 'connected_from',
      key: 'connected_from',
      render: (v: string) => v || '-',
    },
    {
      title: 'Last Heartbeat',
      dataIndex: 'last_heartbeat',
      key: 'last_heartbeat',
      render: (v: string) => formatTimeAgo(v),
    },
    {
      title: 'Scope',
      key: 'scope',
      render: (_, record) =>
        record.is_global ? <Tag color="purple">Global</Tag> : <Tag color="green">Workspace</Tag>,
    },
    {
      title: '',
      key: 'actions',
      width: 60,
      render: (_, record) => (
        <Popconfirm
          title="Delete this worker?"
          description="This action cannot be undone."
          onConfirm={() => handleDeleteWorker(record.uuid!)}
          okText="Delete"
          okButtonProps={{ danger: true }}
        >
          <Button type="text" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  // Internal jobs columns
  const jobsColumns: ColumnsType<WorkerJob> = [
    {
      title: 'Subject',
      dataIndex: 'subject',
      key: 'subject',
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string, record) => {
        const config: Record<string, { color: string; icon: React.ReactNode }> = {
          running: { color: 'processing', icon: <SyncOutlined spin /> },
          done: { color: 'success', icon: <CheckCircleOutlined /> },
          completed: { color: 'success', icon: <CheckCircleOutlined /> },
          failed: { color: 'error', icon: <CloseCircleOutlined /> },
        };
        const statusConfig = config[status] || { color: 'default', icon: null };
        const errorMessage = status === 'failed' && record.data ? (record.data as { error?: string })?.error : undefined;

        return (
          <Tooltip title={errorMessage}>
            <Tag color={statusConfig.color} icon={statusConfig.icon}>
              {status}
            </Tag>
          </Tooltip>
        );
      },
    },
    {
      title: 'Started At',
      dataIndex: 'started_at',
      key: 'started_at',
      render: (v: string) => (v ? new Date(v).toLocaleString() : '-'),
    },
    {
      title: 'Finished At',
      dataIndex: 'finished_at',
      key: 'finished_at',
      render: (v: string) => (v ? new Date(v).toLocaleString() : '-'),
    },
    {
      title: 'Duration',
      key: 'duration',
      render: (_, record) => calculateDuration(record.started_at, record.finished_at),
    },
    {
      title: '',
      key: 'actions',
      width: 80,
      render: (_, record) => (
        <Space>
          {record.status === 'running' && (
            <Popconfirm
              title="Cancel this job?"
              onConfirm={() => handleCancelJob(record.uuid!)}
              okText="Cancel Job"
              okButtonProps={{ danger: true }}
            >
              <Button type="text" danger icon={<StopOutlined />} title="Cancel" />
            </Popconfirm>
          )}
          {record.status !== 'running' && (
            <Popconfirm
              title="Delete this job record?"
              onConfirm={() => handleDeleteJob(record.uuid!)}
              okText="Delete"
              okButtonProps={{ danger: true }}
            >
              <Button type="text" danger icon={<DeleteOutlined />} title="Delete" />
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  // Enrollment tokens columns
  const tokensColumns: ColumnsType<WorkerEnrollmentToken> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Scope',
      key: 'scope',
      render: (_, record) =>
        record.is_global ? <Tag color="purple">Global</Tag> : <Tag color="green">Workspace</Tag>,
    },
    {
      title: 'Expires At',
      dataIndex: 'expires_at',
      key: 'expires_at',
      render: (v: string) => (v ? new Date(v).toLocaleString() : 'Never'),
    },
    {
      title: 'Status',
      key: 'status',
      render: (_, record) => {
        const { status, color } = getTokenStatus(record);
        return <Tag color={color}>{status}</Tag>;
      },
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (v: string) => (v ? new Date(v).toLocaleDateString() : '-'),
    },
    {
      title: '',
      key: 'actions',
      width: 60,
      render: (_, record) => {
        const { status } = getTokenStatus(record);
        return status === 'available' ? (
          <Popconfirm
            title="Revoke this token?"
            description="The token will no longer be usable for enrollment."
            onConfirm={() => handleRevokeToken(record.uuid!)}
            okText="Revoke"
            okButtonProps={{ danger: true }}
          >
            <Button type="text" danger icon={<DeleteOutlined />} title="Revoke" />
          </Popconfirm>
        ) : null;
      },
    },
  ];

  // Tab items configuration
  const tabItems = [
    {
      key: 'external-workers',
      label: (
        <Space>
          <RobotOutlined />
          External Workers
          {workers.filter((w) => w.status === 'online').length > 0 && (
            <Tag color="green" style={{ marginLeft: 4 }}>
              {workers.filter((w) => w.status === 'online').length}
            </Tag>
          )}
        </Space>
      ),
      children: (
        <Table
          columns={workersColumns}
          dataSource={workers}
          rowKey="uuid"
          loading={workersLoading}
          pagination={false}
          locale={{ emptyText: 'No external workers registered' }}
        />
      ),
    },
    {
      key: 'internal-jobs',
      label: (
        <Space>
          <ThunderboltOutlined />
          Internal Jobs
          {jobs.filter((j) => j.status === 'running').length > 0 && (
            <Tag color="blue" style={{ marginLeft: 4 }}>
              {jobs.filter((j) => j.status === 'running').length}
            </Tag>
          )}
        </Space>
      ),
      children: (
        <Table
          columns={jobsColumns}
          dataSource={jobs}
          rowKey="uuid"
          loading={jobsLoading}
          pagination={{ pageSize: 20 }}
        />
      ),
    },
    {
      key: 'enrollment-tokens',
      label: (
        <Space>
          <KeyOutlined />
          Enrollment Tokens
        </Space>
      ),
      children: (
        <Table
          columns={tokensColumns}
          dataSource={tokens}
          rowKey="uuid"
          loading={tokensLoading}
          pagination={false}
          locale={{ emptyText: 'No enrollment tokens created' }}
        />
      ),
    },
  ];

  return (
    <>
      <Title level={4}>Workers</Title>
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={tabItems}
        tabBarExtraContent={
          <Space>
            {activeTab === 'enrollment-tokens' && (
              <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateTokenModalOpen(true)}>
                Create Token
              </Button>
            )}
            <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
              Refresh
            </Button>
          </Space>
        }
      />

      {/* Create Token Modal */}
      <Modal
        title={createdToken ? 'Token Created' : 'Create Enrollment Token'}
        open={createTokenModalOpen}
        onCancel={handleCloseModal}
        footer={
          createdToken ? (
            <Button type="primary" onClick={handleCloseModal}>
              Done
            </Button>
          ) : undefined
        }
        onOk={createdToken ? undefined : handleCreateToken}
        confirmLoading={creating}
        okText="Create"
      >
        {createdToken ? (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Alert
              type="warning"
              message="Save this token now!"
              description="This token will only be shown once. Copy it now and store it securely."
              showIcon
            />
            <Input.TextArea
              value={createdToken}
              readOnly
              autoSize
              style={{ fontFamily: 'monospace', marginTop: 16 }}
            />
            <Button icon={<CopyOutlined />} onClick={handleCopyToken} style={{ marginTop: 8 }}>
              Copy to Clipboard
            </Button>
          </Space>
        ) : (
          <Form form={form} layout="vertical">
            <Form.Item
              name="name"
              label="Worker Name"
              rules={[{ required: true, message: 'Please enter a worker name' }]}
            >
              <Input placeholder="worker-01" />
            </Form.Item>
            <Form.Item name="is_global" valuePropName="checked">
              <Checkbox>Global access (all workspaces)</Checkbox>
            </Form.Item>
            <Form.Item name="expires_at" label="Expiration (optional)">
              <DatePicker showTime style={{ width: '100%' }} />
            </Form.Item>
          </Form>
        )}
      </Modal>
    </>
  );
}

export default Workers;
