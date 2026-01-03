import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import {
  Typography,
  Space,
  Button,
  Table,
  Tag,
  message,
  Popconfirm,
  Tooltip,
  Card,
  Row,
  Col,
  Statistic,
  Select,
} from 'antd';
import {
  DeleteOutlined,
  ReloadOutlined,
  StopOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  SyncOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import type { components } from '../../api/v1';

const { Title, Text } = Typography;

type WorkerJob = components['schemas']['worker_jobs'];

const POLLING_INTERVAL = 5000; // 5 seconds

// Human-readable job type titles
const JOB_TYPE_TITLES: Record<string, string> = {
  tokenRefresh: 'Refresh OAuth Token',
  testConnectionEmailOAuth: 'Test Email Connection',
  testConnectionPostgres: 'Test PostgreSQL Connection',
  emailOAuthFetch: 'Fetch OAuth Emails',
  emailApplyPipeline: 'Apply Email Pipeline',
  dummy: 'Test Job',
};

function getJobTitle(subject: string): string {
  const parts = subject.split('.');
  const jobType = parts[parts.length - 1];
  return JOB_TYPE_TITLES[jobType] || jobType;
}

const TIME_PERIOD_OPTIONS = [
  { value: 1, label: '1 minute' },
  { value: 5, label: '5 minutes' },
  { value: 15, label: '15 minutes' },
  { value: 30, label: '30 minutes' },
  { value: 60, label: '1 hour' },
];

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

function formatTimeAgo(dateStr?: string): string {
  if (!dateStr) return '-';
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSecs = Math.floor(diffMs / 1000);

  if (diffSecs < 60) return `${diffSecs}s ago`;
  const diffMins = Math.floor(diffSecs / 60);
  if (diffMins < 60) return `${diffMins}m ago`;
  const diffHours = Math.floor(diffMins / 60);
  return `${diffHours}h ${diffMins % 60}m ago`;
}

function isWithinTimeWindow(dateStr: string | undefined, windowMinutes: number): boolean {
  if (!dateStr) return false;
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  return diffMs <= windowMinutes * 60 * 1000;
}

function ActiveJobs() {
  const [jobs, setJobs] = useState<WorkerJob[]>([]);
  const [loading, setLoading] = useState(true);
  const [timePeriod, setTimePeriod] = useState(5); // Default 5 minutes
  const pollingPausedRef = useRef(false);

  const loadJobs = useCallback(async () => {
    setLoading(true);
    const { data, error } = await client.GET('/workerjobs', {
      params: { query: { offset: 0, limit: 200 } },
    });
    if (error) {
      message.error('Failed to load jobs');
      setLoading(false);
      return;
    }
    setJobs(data?.jobs || []);
    setLoading(false);
  }, []);

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

  useEffect(() => {
    const poll = async () => {
      if (pollingPausedRef.current) return;
      loadJobs();
    };

    poll();
    const intervalId = setInterval(poll, POLLING_INTERVAL);

    return () => clearInterval(intervalId);
  }, [loadJobs]);

  useEffect(() => {
    const handleVisibilityChange = () => {
      pollingPausedRef.current = document.hidden;
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange);
  }, []);

  // Filter jobs to show only recent ones or currently running
  const recentJobs = useMemo(() => {
    return jobs.filter((job) => {
      // Always show running jobs
      if (job.status === 'running') return true;
      // Show completed/failed jobs from the selected time period
      const relevantTime = job.finished_at || job.started_at;
      return isWithinTimeWindow(relevantTime, timePeriod);
    });
  }, [jobs, timePeriod]);

  // Calculate stats
  const runningCount = recentJobs.filter((j) => j.status === 'running').length;
  const completedCount = recentJobs.filter((j) => j.status === 'completed' || j.status === 'done').length;
  const failedCount = recentJobs.filter((j) => j.status === 'failed').length;

  const columns: ColumnsType<WorkerJob> = [
    {
      title: 'Job Type',
      dataIndex: 'subject',
      key: 'subject',
      render: (subject: string) => {
        const title = getJobTitle(subject);
        return (
          <Tooltip title={subject}>
            <Text strong>{title}</Text>
          </Tooltip>
        );
      },
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
      title: 'Started',
      dataIndex: 'started_at',
      key: 'started_at',
      render: (v: string) => formatTimeAgo(v),
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

  return (
    <>
      <Space style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Space>
          <Title level={4} style={{ margin: 0 }}>Active Jobs</Title>
          <Select
            value={timePeriod}
            onChange={setTimePeriod}
            options={TIME_PERIOD_OPTIONS}
            style={{ width: 120 }}
            suffixIcon={<ClockCircleOutlined />}
          />
          {runningCount > 0 && (
            <Tag color="processing" icon={<SyncOutlined spin />}>
              {runningCount} running
            </Tag>
          )}
        </Space>
        <Button icon={<ReloadOutlined />} onClick={loadJobs}>
          Refresh
        </Button>
      </Space>

      {/* Summary Cards */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Card size="small">
            <Statistic
              title="Running"
              value={runningCount}
              valueStyle={{ color: runningCount > 0 ? '#1890ff' : undefined }}
              prefix={<SyncOutlined spin={runningCount > 0} />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card size="small">
            <Statistic
              title="Completed"
              value={completedCount}
              valueStyle={{ color: '#3f8600' }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card size="small">
            <Statistic
              title="Failed"
              value={failedCount}
              valueStyle={{ color: failedCount > 0 ? '#cf1322' : undefined }}
              prefix={<CloseCircleOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Table
        columns={columns}
        dataSource={recentJobs}
        rowKey="uuid"
        loading={loading}
        pagination={{ pageSize: 20 }}
        locale={{ emptyText: `No jobs in the last ${TIME_PERIOD_OPTIONS.find(o => o.value === timePeriod)?.label || timePeriod + ' minutes'}` }}
      />

      {recentJobs.length > 0 && (
        <Text type="secondary" style={{ display: 'block', marginTop: 8 }}>
          Showing {recentJobs.length} job{recentJobs.length !== 1 ? 's' : ''} from the last {TIME_PERIOD_OPTIONS.find(o => o.value === timePeriod)?.label || timePeriod + ' minutes'}
        </Text>
      )}
    </>
  );
}

export default ActiveJobs;
