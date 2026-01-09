import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router';
import { Table, Button, Space, Typography, message, Tag, Popconfirm, Tooltip } from 'antd';
import { PlusOutlined, DeleteOutlined, LoginOutlined, StopOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import type { components } from '../../api/v1';

const { Title } = Typography;

type Datasource = components['schemas']['datasource'];

type DatasourceTypeKey = 'email' | 'email_oauth';

const typeLabels: Record<DatasourceTypeKey, { label: string; color: string }> = {
  email: { label: 'Email IMAP', color: 'blue' },
  email_oauth: { label: 'Email OAuth', color: 'cyan' },
};

function DataSources() {
  const navigate = useNavigate();
  const { slug } = useWorkspace();
  const [loading, setLoading] = useState(true);
  const [datasources, setDatasources] = useState<Datasource[]>([]);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const loadDatasources = useCallback(async () => {
    setLoading(true);
    const { data, error } = await client.GET('/datasource');
    if (error) {
      message.error('Failed to load data sources');
      setLoading(false);
      return;
    }
    setDatasources(data || []);
    setLoading(false);
  }, []);

  useEffect(() => {
    loadDatasources();
  }, [loadDatasources]);

  const getDeleteEndpoint = (type: string) => {
    switch (type) {
      case 'email':
        return '/datasource/email/{uuid}' as const;
      case 'email_oauth':
        return '/datasource/email_oauth/{uuid}' as const;
      default:
        return '/datasource/email/{uuid}' as const;
    }
  };

  const handleDelete = async (record: Datasource) => {
    const endpoint = getDeleteEndpoint(record.type);
    const { error } = await client.DELETE(endpoint, {
      params: { path: { uuid: record.uuid! } },
    });
    if (error) {
      message.error('Failed to delete data source');
      return;
    }
    message.success('Data source deleted');
    loadDatasources();
  };

  const handleOAuthLogin = async (record: Datasource) => {
    if (record.type !== 'email_oauth') return;
    setActionLoading(record.uuid!);

    const { data, error } = await client.POST('/oauth2/login', {
      body: {
        query: { datasource_uuid: [record.uuid!], workspace_slug: [slug] },
      },
    });

    setActionLoading(null);

    if (error) {
      message.error(error.detail || 'Failed to initiate OAuth login');
      return;
    }

    if (data?.auth_code_url) {
      window.location.href = data.auth_code_url;
    }
  };

  const handleRevokeTokens = async (record: Datasource) => {
    setActionLoading(record.uuid!);

    // Fetch all tokens for this datasource
    const { data: tokens, error: listError } = await client.GET('/oauth2/client/{datasource_uuid}/token', {
      params: { path: { datasource_uuid: record.uuid! } },
    });

    if (listError) {
      message.error('Failed to fetch tokens');
      setActionLoading(null);
      return;
    }

    // Delete each token
    for (const token of tokens || []) {
      await client.DELETE('/oauth2/client/{datasource_uuid}/token/{uuid}', {
        params: { path: { datasource_uuid: record.uuid!, uuid: token.uuid! } },
      });
    }

    setActionLoading(null);
    // Update the datasource's is_oauth_authenticated status locally
    setDatasources((prev) =>
      prev.map((ds) =>
        ds.uuid === record.uuid ? { ...ds, is_oauth_authenticated: false } : ds
      )
    );
    message.success('All tokens revoked');
  };

  const columns: ColumnsType<Datasource> = [
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
        const config = typeLabels[type as DatasourceTypeKey];
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
      render: (isEnabled: boolean, record: Datasource) => (
        <Space size="small">
          <Tag color={isEnabled ? 'success' : 'default'}>
            {isEnabled ? 'Enabled' : 'Disabled'}
          </Tag>
          {record.type === 'email_oauth' && (
            <Tag
              color={record.is_oauth_authenticated ? 'success' : 'warning'}
              icon={record.is_oauth_authenticated ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
            >
              {record.is_oauth_authenticated ? 'Authenticated' : 'Not Authenticated'}
            </Tag>
          )}
        </Space>
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
      width: 140,
      render: (_, record) => (
        <Space size="small" onClick={(e) => e.stopPropagation()}>
          {record.type === 'email_oauth' && (
            <>
              <Tooltip title="Authorize OAuth">
                <Button
                  type="text"
                  icon={<LoginOutlined />}
                  loading={actionLoading === record.uuid}
                  onClick={(e) => {
                    e.stopPropagation();
                    handleOAuthLogin(record);
                  }}
                />
              </Tooltip>
              <Tooltip title="Revoke Tokens">
                <Popconfirm
                  title="Revoke all tokens"
                  description="This will revoke all OAuth tokens for this data source."
                  onConfirm={() => handleRevokeTokens(record)}
                  okButtonProps={{ danger: true }}
                  okText="Revoke"
                >
                  <Button
                    type="text"
                    danger
                    icon={<StopOutlined />}
                    onClick={(e) => e.stopPropagation()}
                  />
                </Popconfirm>
              </Tooltip>
            </>
          )}
          <Popconfirm
            title="Delete data source"
            description="Are you sure you want to delete this data source?"
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
          Data Sources
        </Title>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate(`/w/${slug}/datasources/new`)}
        >
          Add Data Source
        </Button>
      </Space>
      <Table
        columns={columns}
        dataSource={datasources}
        rowKey="uuid"
        loading={loading}
        pagination={false}
        onRow={(record) => ({
          onClick: () => navigate(`/w/${slug}/datasources/${record.uuid}`),
          style: { cursor: 'pointer' },
        })}
      />
    </>
  );
}

export default DataSources;
