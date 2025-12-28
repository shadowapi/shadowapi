import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router';
import { Table, Button, Space, Typography, message, Tag, Result, Popconfirm } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import { useAuth, isAdmin } from '../../lib/auth';
import type { components } from '../../api/v1';

const { Title } = Typography;

type User = components['schemas']['user'];

function Users() {
  const navigate = useNavigate();
  const { slug } = useWorkspace();
  const { user: currentUser } = useAuth();
  const [loading, setLoading] = useState(true);
  const [users, setUsers] = useState<User[]>([]);

  // Admin access check
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

  const loadUsers = async () => {
    setLoading(true);
    const { data, error } = await client.GET('/user');
    if (error) {
      message.error('Failed to load users');
      setLoading(false);
      return;
    }
    setUsers(data || []);
    setLoading(false);
  };

  useEffect(() => {
    loadUsers();
  }, []);

  const handleDelete = async (uuid: string) => {
    const { error } = await client.DELETE('/user/{uuid}', {
      params: { path: { uuid } },
    });
    if (error) {
      message.error('Failed to delete user');
      return;
    }
    message.success('User deleted');
    loadUsers();
  };

  const columns: ColumnsType<User> = [
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: 'Name',
      key: 'name',
      render: (_, record) => `${record.first_name} ${record.last_name}`,
    },
    {
      title: 'Status',
      key: 'status',
      render: (_, record) => (
        <Space>
          {record.is_enabled ? (
            <Tag color="green">Enabled</Tag>
          ) : (
            <Tag color="red">Disabled</Tag>
          )}
          {record.roles?.some(r => r.role === 'super_admin' && r.domain === 'global') && (
            <Tag color="blue">Super Admin</Tag>
          )}
        </Space>
      ),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (value: string) =>
        value ? new Date(value).toLocaleDateString() : '-',
    },
    {
      title: '',
      key: 'actions',
      width: 60,
      render: (_, record) => (
        <Popconfirm
          title="Delete user"
          description="Are you sure you want to delete this user?"
          onConfirm={() => handleDelete(record.uuid!)}
          okButtonProps={{ danger: true }}
          okText="Delete"
        >
          <Button
            type="text"
            danger
            icon={<DeleteOutlined />}
            disabled={record.uuid === currentUser?.uuid}
            onClick={(e) => e.stopPropagation()}
          />
        </Popconfirm>
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
          Users
        </Title>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate(`/w/${slug}/users/new`)}
        >
          Add User
        </Button>
      </Space>
      <Table
        columns={columns}
        dataSource={users}
        rowKey="uuid"
        loading={loading}
        pagination={false}
        onRow={(record) => ({
          onClick: () => navigate(`/w/${slug}/users/${record.uuid}`),
          style: { cursor: 'pointer' },
        })}
      />
    </>
  );
}

export default Users;
