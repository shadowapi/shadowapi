import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router';
import { Table, Button, Space, Typography, message, Tag, Result, Popconfirm, Select, Switch } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import { useAuth, isAdmin } from '../../lib/auth';
import type { components } from '../../api/v1';

const { Title } = Typography;

type WorkerUsageLimit = components['schemas']['worker_usage_limit'];
type RegisteredWorker = components['schemas']['registered_worker'];

const limitTypeLabels: Record<string, string> = {
  messages_fetch: 'Messages Fetch',
  messages_push: 'Messages Push',
};

const resetPeriodLabels: Record<string, string> = {
  daily: 'Daily',
  weekly: 'Weekly',
  monthly: 'Monthly',
  rolling_24h: 'Rolling 24h',
  rolling_7d: 'Rolling 7d',
  rolling_30d: 'Rolling 30d',
};

function WorkerUsageLimits() {
  const navigate = useNavigate();
  const { slug } = useWorkspace();
  const { user: currentUser } = useAuth();
  const [loading, setLoading] = useState(true);
  const [limits, setLimits] = useState<WorkerUsageLimit[]>([]);
  const [workers, setWorkers] = useState<RegisteredWorker[]>([]);
  const [workerFilter, setWorkerFilter] = useState<string | undefined>(undefined);

  const loadData = useCallback(async () => {
    setLoading(true);

    // Load workers for display
    const workersRes = await client.GET('/workers');
    if (workersRes.data) {
      setWorkers(workersRes.data);
    }

    // Load limits for all workers
    const allLimits: WorkerUsageLimit[] = [];

    if (workersRes.data) {
      for (const worker of workersRes.data) {
        if (!worker.uuid) continue;
        const { data } = await client.GET('/access/worker/{worker_uuid}/usage-limits', {
          params: {
            path: { worker_uuid: worker.uuid },
            query: { workspace_slug: slug },
          },
        });
        if (data) {
          allLimits.push(...data);
        }
      }
    }

    setLimits(allLimits);
    setLoading(false);
  }, [slug]);

  useEffect(() => {
    if (isAdmin(currentUser)) {
      loadData();
    }
  }, [loadData, currentUser]);

  const handleDelete = async (workerUuid: string, uuid: string) => {
    const { error } = await client.DELETE('/access/worker/{worker_uuid}/usage-limits/{uuid}', {
      params: { path: { worker_uuid: workerUuid, uuid } },
    });
    if (error) {
      message.error('Failed to delete worker usage limit');
      return;
    }
    message.success('Worker usage limit deleted');
    loadData();
  };

  // Filter limits by worker
  const filteredLimits = workerFilter
    ? limits.filter((l) => l.worker_uuid === workerFilter)
    : limits;

  // Create worker lookup
  const workerMap = new Map(workers.map((w) => [w.uuid, w]));

  const columns: ColumnsType<WorkerUsageLimit> = [
    {
      title: 'Worker',
      dataIndex: 'worker_uuid',
      key: 'worker_uuid',
      render: (uuid: string) => {
        const worker = workerMap.get(uuid);
        return worker?.name || uuid.slice(0, 8) + '...';
      },
    },
    {
      title: 'Workspace',
      dataIndex: 'workspace_slug',
      key: 'workspace_slug',
    },
    {
      title: 'Limit Type',
      dataIndex: 'limit_type',
      key: 'limit_type',
      render: (value: string) => (
        <Tag color={value === 'messages_fetch' ? 'blue' : 'green'}>
          {limitTypeLabels[value] || value}
        </Tag>
      ),
    },
    {
      title: 'Limit Value',
      dataIndex: 'limit_value',
      key: 'limit_value',
      render: (value: number | null) =>
        value === null ? <Tag color="purple">Unlimited</Tag> : value.toLocaleString(),
    },
    {
      title: 'Reset Period',
      dataIndex: 'reset_period',
      key: 'reset_period',
      render: (value: string) => resetPeriodLabels[value] || value,
    },
    {
      title: 'Enabled',
      dataIndex: 'is_enabled',
      key: 'is_enabled',
      render: (value: boolean) => <Switch checked={value} disabled size="small" />,
    },
    {
      title: '',
      key: 'actions',
      width: 60,
      render: (_, record) => (
        <Popconfirm
          title="Delete worker limit"
          description="Are you sure you want to delete this worker usage limit?"
          onConfirm={() => handleDelete(record.worker_uuid, record.uuid!)}
          okButtonProps={{ danger: true }}
          okText="Delete"
        >
          <Button
            type="text"
            danger
            icon={<DeleteOutlined />}
            title="Delete"
            onClick={(e) => e.stopPropagation()}
          />
        </Popconfirm>
      ),
    },
  ];

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

  // Get unique workers for filter
  const workerOptions = workers.map((w) => ({
    label: w.name || w.uuid?.slice(0, 8) + '...',
    value: w.uuid,
  }));

  return (
    <>
      <Space
        style={{
          marginBottom: 16,
          display: 'flex',
          justifyContent: 'space-between',
        }}
      >
        <Space>
          <Title level={4} style={{ margin: 0 }}>
            Worker Usage Limits
          </Title>
          <Select
            placeholder="Filter by worker"
            allowClear
            style={{ width: 200 }}
            value={workerFilter}
            onChange={(value) => setWorkerFilter(value)}
            options={[{ label: 'All Workers', value: undefined }, ...workerOptions]}
          />
        </Space>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate(`/w/${slug}/access/worker-usage-limits/new`)}
        >
          Create Worker Limit
        </Button>
      </Space>
      <Table
        columns={columns}
        dataSource={filteredLimits}
        rowKey="uuid"
        loading={loading}
        pagination={false}
        onRow={(record) => ({
          onClick: () => navigate(`/w/${slug}/access/worker-usage-limits/${record.uuid}?worker=${record.worker_uuid}`),
          style: { cursor: 'pointer' },
        })}
      />
    </>
  );
}

export default WorkerUsageLimits;
