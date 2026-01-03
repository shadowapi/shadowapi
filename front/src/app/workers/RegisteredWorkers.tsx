import { useState, useEffect, useCallback, useRef } from 'react';
import {
  Typography,
  Space,
  Button,
  Table,
  Tag,
  message,
  Popconfirm,
} from 'antd';
import {
  DeleteOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  SyncOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import type { components } from '../../api/v1';

const { Title } = Typography;

type RegisteredWorker = components['schemas']['registered_worker'];

const POLLING_INTERVAL = 10000; // 10 seconds

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

function RegisteredWorkers() {
  const [workers, setWorkers] = useState<RegisteredWorker[]>([]);
  const [loading, setLoading] = useState(true);
  const pollingPausedRef = useRef(false);

  const loadWorkers = useCallback(async () => {
    setLoading(true);
    const { data, error } = await client.GET('/workers');
    if (error) {
      message.error('Failed to load workers');
      setLoading(false);
      return;
    }
    setWorkers(data || []);
    setLoading(false);
  }, []);

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

  useEffect(() => {
    const poll = async () => {
      if (pollingPausedRef.current) return;
      loadWorkers();
    };

    poll();
    const intervalId = setInterval(poll, POLLING_INTERVAL);

    return () => clearInterval(intervalId);
  }, [loadWorkers]);

  useEffect(() => {
    const handleVisibilityChange = () => {
      pollingPausedRef.current = document.hidden;
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange);
  }, []);

  const columns: ColumnsType<RegisteredWorker> = [
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

  const onlineCount = workers.filter((w) => w.status === 'online').length;

  return (
    <>
      <Space style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Space>
          <Title level={4} style={{ margin: 0 }}>Registered Workers</Title>
          {onlineCount > 0 && (
            <Tag color="green">{onlineCount} online</Tag>
          )}
        </Space>
        <Button icon={<ReloadOutlined />} onClick={loadWorkers}>
          Refresh
        </Button>
      </Space>
      <Table
        columns={columns}
        dataSource={workers}
        rowKey="uuid"
        loading={loading}
        pagination={false}
        locale={{ emptyText: 'No workers registered' }}
      />
    </>
  );
}

export default RegisteredWorkers;
