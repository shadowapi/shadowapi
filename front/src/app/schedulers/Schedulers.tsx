import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router';
import { Table, Button, Space, Typography, message, Tag, Popconfirm, Switch, Tooltip } from 'antd';
import { PlusOutlined, DeleteOutlined, PauseCircleOutlined, PlayCircleOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import type { components } from '../../api/v1';

const { Title } = Typography;

type Scheduler = components['schemas']['scheduler'];
type Pipeline = components['schemas']['pipeline'];

function Schedulers() {
  const navigate = useNavigate();
  const { slug } = useWorkspace();
  const [loading, setLoading] = useState(true);
  const [schedulers, setSchedulers] = useState<Scheduler[]>([]);
  const [pipelines, setPipelines] = useState<Record<string, Pipeline>>({});

  const loadData = useCallback(async () => {
    setLoading(true);

    const [schedulersRes, pipelinesRes] = await Promise.all([
      client.GET('/scheduler'),
      client.GET('/pipeline'),
    ]);

    if (schedulersRes.error) {
      message.error('Failed to load schedulers');
      setLoading(false);
      return;
    }

    setSchedulers(schedulersRes.data || []);

    // Create pipeline lookup map
    const pipelineMap: Record<string, Pipeline> = {};
    const pipelinesData = pipelinesRes.data as { pipelines?: Pipeline[] } | undefined;
    for (const p of pipelinesData?.pipelines || []) {
      if (p.uuid) pipelineMap[p.uuid] = p;
    }
    setPipelines(pipelineMap);

    setLoading(false);
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleDelete = async (record: Scheduler) => {
    const { error } = await client.DELETE('/scheduler/{uuid}', {
      params: { path: { uuid: record.uuid! } },
    });
    if (error) {
      message.error('Failed to delete scheduler');
      return;
    }
    message.success('Scheduler deleted');
    loadData();
  };

  const handleToggleEnabled = async (record: Scheduler, enabled: boolean) => {
    const { error } = await client.PUT('/scheduler/{uuid}', {
      params: { path: { uuid: record.uuid! } },
      body: {
        ...record,
        is_enabled: enabled,
      },
    });
    if (error) {
      message.error('Failed to update scheduler');
      return;
    }
    message.success(enabled ? 'Scheduler enabled' : 'Scheduler disabled');
    loadData();
  };

  const handleTogglePaused = async (record: Scheduler) => {
    const { error } = await client.PUT('/scheduler/{uuid}', {
      params: { path: { uuid: record.uuid! } },
      body: {
        ...record,
        is_paused: !record.is_paused,
      },
    });
    if (error) {
      message.error('Failed to update scheduler');
      return;
    }
    message.success(record.is_paused ? 'Scheduler resumed' : 'Scheduler paused');
    loadData();
  };

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleString();
  };

  const columns: ColumnsType<Scheduler> = [
    {
      title: 'Pipeline',
      dataIndex: 'pipeline_uuid',
      key: 'pipeline',
      render: (uuid: string) => {
        const pipeline = pipelines[uuid];
        return pipeline?.name || uuid || '-';
      },
    },
    {
      title: 'Type',
      dataIndex: 'schedule_type',
      key: 'type',
      render: (type: string) => (
        <Tag color={type === 'cron' ? 'blue' : 'purple'}>
          {type === 'cron' ? 'Recurring' : 'One-time'}
        </Tag>
      ),
    },
    {
      title: 'Schedule',
      key: 'schedule',
      render: (_, record) => {
        if (record.schedule_type === 'cron') {
          return (
            <Tooltip title={`Timezone: ${record.timezone || 'UTC'}`}>
              <code>{record.cron_expression || '-'}</code>
            </Tooltip>
          );
        }
        return formatDate(record.run_at ?? undefined);
      },
    },
    {
      title: 'Next Run',
      dataIndex: 'next_run',
      key: 'next_run',
      render: (date: string) => formatDate(date),
    },
    {
      title: 'Last Run',
      dataIndex: 'last_run',
      key: 'last_run',
      render: (date: string) => formatDate(date),
    },
    {
      title: 'Status',
      key: 'status',
      render: (_, record) => {
        if (!record.is_enabled) {
          return <Tag color="default">Disabled</Tag>;
        }
        if (record.is_paused) {
          return <Tag color="warning">Paused</Tag>;
        }
        return <Tag color="success">Active</Tag>;
      },
    },
    {
      title: 'Enabled',
      dataIndex: 'is_enabled',
      key: 'is_enabled',
      width: 80,
      render: (isEnabled: boolean, record) => (
        <Switch
          size="small"
          checked={isEnabled}
          onChange={(checked) => handleToggleEnabled(record, checked)}
          onClick={(_, e) => e.stopPropagation()}
        />
      ),
    },
    {
      title: '',
      key: 'actions',
      width: 100,
      render: (_, record) => (
        <Space size="small" onClick={(e) => e.stopPropagation()}>
          {record.is_enabled && (
            <Tooltip title={record.is_paused ? 'Resume' : 'Pause'}>
              <Button
                type="text"
                icon={record.is_paused ? <PlayCircleOutlined /> : <PauseCircleOutlined />}
                onClick={() => handleTogglePaused(record)}
              />
            </Tooltip>
          )}
          <Popconfirm
            title="Delete scheduler"
            description="Are you sure you want to delete this scheduler?"
            onConfirm={() => handleDelete(record)}
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
        </Space>
      ),
    },
  ];

  return (
    <>
      <Space
        style={{
          marginBottom: 16,
          display: 'flex',
          justifyContent: 'space-between',
        }}
      >
        <Title level={4} style={{ margin: 0 }}>
          Schedulers
        </Title>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate(`/w/${slug}/schedulers/new`)}
        >
          Add Scheduler
        </Button>
      </Space>
      <Table
        columns={columns}
        dataSource={schedulers}
        rowKey="uuid"
        loading={loading}
        pagination={false}
        onRow={(record) => ({
          onClick: () => navigate(`/w/${slug}/schedulers/${record.uuid}`),
          style: { cursor: 'pointer' },
        })}
      />
    </>
  );
}

export default Schedulers;
