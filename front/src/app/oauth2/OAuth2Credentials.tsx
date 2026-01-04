import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router';
import { Typography, Button, Table, Space, message, Tag, Popconfirm } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import type { components } from '../../api/v1';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';

const { Title } = Typography;

type OAuth2Client = components['schemas']['oauth2_client'];

type ProviderKey = 'google' | 'github' | 'microsoft' | 'azure' | 'okta';

const providerLabels: Record<ProviderKey, { label: string; color: string }> = {
  google: { label: 'Google', color: 'blue' },
  github: { label: 'GitHub', color: 'purple' },
  microsoft: { label: 'Microsoft', color: 'cyan' },
  azure: { label: 'Azure AD', color: 'geekblue' },
  okta: { label: 'Okta', color: 'orange' },
};

function OAuth2Credentials() {
  const navigate = useNavigate();
  const { slug } = useWorkspace();
  const [loading, setLoading] = useState(true);
  const [clients, setClients] = useState<OAuth2Client[]>([]);

  const loadClients = useCallback(async () => {
    setLoading(true);
    const { data, error } = await client.GET('/oauth2/client');
    if (error) {
      message.error('Failed to load OAuth2 credentials');
      setLoading(false);
      return;
    }
    setClients(data?.clients || []);
    setLoading(false);
  }, []);

  useEffect(() => {
    loadClients();
  }, [loadClients]);

  const handleDelete = async (uuid: string) => {
    const { error } = await client.DELETE('/oauth2/client/{uuid}', {
      params: { path: { uuid } },
    });
    if (error) {
      message.error('Failed to delete OAuth2 credential');
      return;
    }
    message.success('OAuth2 credential deleted');
    loadClients();
  };

  const columns: ColumnsType<OAuth2Client> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Provider',
      dataIndex: 'provider',
      key: 'provider',
      render: (provider: string) => {
        const config = providerLabels[provider.toLowerCase() as ProviderKey];
        return config ? (
          <Tag color={config.color}>{config.label}</Tag>
        ) : (
          <Tag>{provider}</Tag>
        );
      },
    },
    {
      title: 'Client ID',
      dataIndex: 'client_id',
      key: 'client_id',
      render: (clientId: string) => {
        if (!clientId) return '-';
        // Truncate long client IDs
        return clientId.length > 30
          ? `${clientId.substring(0, 30)}...`
          : clientId;
      },
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) =>
        date ? new Date(date).toLocaleDateString() : '-',
    },
    {
      title: '',
      key: 'actions',
      width: 60,
      render: (_, record) => (
        <Space size="small" onClick={(e) => e.stopPropagation()}>
          <Popconfirm
            title="Delete credential"
            description="Are you sure? This may break datasources using this credential."
            onConfirm={() => handleDelete(record.uuid!)}
            okButtonProps={{ danger: true }}
            okText="Delete"
          >
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              title="Delete"
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Space style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Title level={4} style={{ margin: 0 }}>OAuth2 Credentials</Title>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate(`/w/${slug}/oauth2/credentials/new`)}
        >
          Add Credential
        </Button>
      </Space>
      <Table
        columns={columns}
        dataSource={clients}
        rowKey="uuid"
        loading={loading}
        pagination={false}
        onRow={(record) => ({
          onClick: () => navigate(`/w/${slug}/oauth2/credentials/${record.uuid}`),
          style: { cursor: 'pointer' },
        })}
      />
    </>
  );
}

export default OAuth2Credentials;
