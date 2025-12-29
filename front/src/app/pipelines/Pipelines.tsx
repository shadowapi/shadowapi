import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router';
import { Table, Button, Space, Typography, message, Tag, Popconfirm } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import type { components } from '../../api/v1';

const { Title } = Typography;

type Pipeline = components['schemas']['pipeline'];
type Datasource = components['schemas']['datasource'];
type Storage = components['schemas']['storage'];

type PipelineTypeKey = 'email' | 'email_oauth' | 'telegram' | 'whatsapp' | 'linkedin';

const typeLabels: Record<PipelineTypeKey, { label: string; color: string }> = {
  email: { label: 'Email', color: 'blue' },
  email_oauth: { label: 'Email OAuth', color: 'cyan' },
  telegram: { label: 'Telegram', color: 'geekblue' },
  whatsapp: { label: 'WhatsApp', color: 'green' },
  linkedin: { label: 'LinkedIn', color: 'purple' },
};

function Pipelines() {
  const navigate = useNavigate();
  const { slug } = useWorkspace();
  const [loading, setLoading] = useState(true);
  const [pipelines, setPipelines] = useState<Pipeline[]>([]);
  const [datasources, setDatasources] = useState<Record<string, Datasource>>({});
  const [storages, setStorages] = useState<Record<string, Storage>>({});

  const loadData = useCallback(async () => {
    setLoading(true);

    // Fetch pipelines, datasources, and storages in parallel
    const [pipelinesRes, datasourcesRes, storagesRes] = await Promise.all([
      client.GET('/pipeline'),
      client.GET('/datasource'),
      client.GET('/storage'),
    ]);

    if (pipelinesRes.error) {
      message.error('Failed to load pipelines');
      setLoading(false);
      return;
    }

    setPipelines(pipelinesRes.data?.pipelines || []);

    // Create lookup maps
    const dsMap: Record<string, Datasource> = {};
    for (const ds of datasourcesRes.data || []) {
      if (ds.uuid) dsMap[ds.uuid] = ds;
    }
    setDatasources(dsMap);

    const stMap: Record<string, Storage> = {};
    for (const st of storagesRes.data || []) {
      if (st.uuid) stMap[st.uuid] = st;
    }
    setStorages(stMap);

    setLoading(false);
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleDelete = async (record: Pipeline) => {
    const { error } = await client.DELETE('/pipeline/{uuid}', {
      params: { path: { uuid: record.uuid! } },
    });
    if (error) {
      message.error('Failed to delete pipeline');
      return;
    }
    message.success('Pipeline deleted');
    loadData();
  };

  const columns: ColumnsType<Pipeline> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => {
        const config = typeLabels[type as PipelineTypeKey];
        return config ? (
          <Tag color={config.color}>{config.label}</Tag>
        ) : (
          <Tag>{type || '-'}</Tag>
        );
      },
    },
    {
      title: 'Data Source',
      dataIndex: 'datasource_uuid',
      key: 'datasource_uuid',
      render: (uuid: string) => {
        const ds = datasources[uuid];
        return ds?.name || uuid || '-';
      },
    },
    {
      title: 'Storage',
      dataIndex: 'storage_uuid',
      key: 'storage_uuid',
      render: (uuid: string) => {
        const st = storages[uuid];
        return st?.name || uuid || '-';
      },
    },
    {
      title: 'Status',
      dataIndex: 'is_enabled',
      key: 'is_enabled',
      render: (isEnabled: boolean) => (
        <Tag color={isEnabled ? 'success' : 'default'}>
          {isEnabled ? 'Enabled' : 'Disabled'}
        </Tag>
      ),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => date ? new Date(date).toLocaleDateString() : '-',
    },
    {
      title: '',
      key: 'actions',
      width: 60,
      render: (_, record) => (
        <Space size="small" onClick={(e) => e.stopPropagation()}>
          <Popconfirm
            title="Delete pipeline"
            description="Are you sure you want to delete this pipeline?"
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
          Data Pipelines
        </Title>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate(`/w/${slug}/pipelines/new`)}
        >
          Add Pipeline
        </Button>
      </Space>
      <Table
        columns={columns}
        dataSource={pipelines}
        rowKey="uuid"
        loading={loading}
        pagination={false}
        onRow={(record) => ({
          onClick: () => navigate(`/w/${slug}/pipelines/${record.uuid}`),
          style: { cursor: 'pointer' },
        })}
      />
    </>
  );
}

export default Pipelines;
