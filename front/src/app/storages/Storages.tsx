import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router';
import { Table, Button, Space, Typography, message, Tag, Popconfirm } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import type { components } from '../../api/v1';

const { Title } = Typography;

type Storage = components['schemas']['storage'];

type StorageTypeKey = 's3' | 'postgres' | 'hostfiles';

const typeLabels: Record<StorageTypeKey, { label: string; color: string }> = {
  s3: { label: 'S3', color: 'blue' },
  postgres: { label: 'PostgreSQL', color: 'green' },
  hostfiles: { label: 'Host Files', color: 'orange' },
};

function Storages() {
  const navigate = useNavigate();
  const { slug } = useWorkspace();
  const [loading, setLoading] = useState(true);
  const [storages, setStorages] = useState<Storage[]>([]);

  const loadStorages = useCallback(async () => {
    setLoading(true);
    const { data, error } = await client.GET('/storage');
    if (error) {
      message.error('Failed to load storages');
      setLoading(false);
      return;
    }
    setStorages(data || []);
    setLoading(false);
  }, []);

  useEffect(() => {
    loadStorages();
  }, [loadStorages]);

  const getDeleteEndpoint = (type: string) => {
    switch (type) {
      case 's3':
        return '/storage/s3/{uuid}' as const;
      case 'postgres':
        return '/storage/postgres/{uuid}' as const;
      case 'hostfiles':
        return '/storage/hostfiles/{uuid}' as const;
      default:
        return '/storage/s3/{uuid}' as const;
    }
  };

  const handleDelete = async (record: Storage) => {
    const endpoint = getDeleteEndpoint(record.type);
    const { error } = await client.DELETE(endpoint, {
      params: { path: { uuid: record.uuid! } },
    });
    if (error) {
      message.error('Failed to delete storage');
      return;
    }
    message.success('Storage deleted');
    loadStorages();
  };

  const columns: ColumnsType<Storage> = [
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
        const config = typeLabels[type as StorageTypeKey];
        return config ? (
          <Tag color={config.color}>{config.label}</Tag>
        ) : (
          <Tag>{type}</Tag>
        );
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
            title="Delete storage"
            description="Are you sure you want to delete this storage?"
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
          Data Storages
        </Title>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate(`/w/${slug}/storages/new`)}
        >
          Add Storage
        </Button>
      </Space>
      <Table
        columns={columns}
        dataSource={storages}
        rowKey="uuid"
        loading={loading}
        pagination={false}
        onRow={(record) => ({
          onClick: () => navigate(`/w/${slug}/storages/${record.uuid}`),
          style: { cursor: 'pointer' },
        })}
      />
    </>
  );
}

export default Storages;
