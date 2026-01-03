import { useState, useEffect, useCallback, useRef } from 'react';
import {
  Typography,
  Space,
  Button,
  Table,
  Tag,
  message,
  Popconfirm,
  Tooltip,
} from 'antd';
import {
  DeleteOutlined,
  ReloadOutlined,
  StopOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  SyncOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import type { components } from '../../api/v1';

const { Title } = Typography;

type WorkerJob = components['schemas']['worker_jobs'];

const POLLING_INTERVAL = 10000; // 10 seconds

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

function InternalJobs() {
  const [jobs, setJobs] = useState<WorkerJob[]>([]);
  const [loading, setLoading] = useState(true);
  const pollingPausedRef = useRef(false);

  const loadJobs = useCallback(async () => {
    setLoading(true);
    const { data, error } = await client.GET('/workerjobs', {
      params: { query: { offset: 0, limit: 100 } },
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

  const columns: ColumnsType<WorkerJob> = [
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

  const runningCount = jobs.filter((j) => j.status === 'running').length;

  return (
    <>
      <Space style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Space>
          <Title level={4} style={{ margin: 0 }}>Internal Jobs</Title>
          {runningCount > 0 && (
            <Tag color="blue">{runningCount} running</Tag>
          )}
        </Space>
        <Button icon={<ReloadOutlined />} onClick={loadJobs}>
          Refresh
        </Button>
      </Space>
      <Table
        columns={columns}
        dataSource={jobs}
        rowKey="uuid"
        loading={loading}
        pagination={{ pageSize: 20 }}
      />
    </>
  );
}

export default InternalJobs;
