import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router';
import { Typography, Button, Table, Space, message } from 'antd';
import { PlusOutlined, EditOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import type { components } from '../../api/v1';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';

const { Title } = Typography;

type OAuth2Client = components['schemas']['oauth2_client'];

function OAuth2Credentials() {
  const navigate = useNavigate();
  const { slug } = useWorkspace();
  const [loading, setLoading] = useState(true);
  const [clients, setClients] = useState<OAuth2Client[]>([]);

  useEffect(() => {
    const loadClients = async () => {
      setLoading(true);
      const { data, error } = await client.GET('/oauth2/client');
      if (error) {
        message.error('Failed to load OAuth2 credentials');
        setLoading(false);
        return;
      }
      setClients(data?.clients || []);
      setLoading(false);
    };
    loadClients();
  }, []);

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
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 100,
      render: (_, record) => (
        <Button
          type="text"
          icon={<EditOutlined />}
          onClick={() => navigate(`/w/${slug}/oauth2/credentials/${record.uuid}`)}
        />
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
          Add OAuth2 Credential
        </Button>
      </Space>
      <Table
        columns={columns}
        dataSource={clients}
        rowKey="uuid"
        loading={loading}
        pagination={false}
      />
    </>
  );
}

export default OAuth2Credentials;
