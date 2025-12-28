import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router';
import { Table, Button, Space, Typography, message, Tag, Result, Popconfirm, Select } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import { useAuth, isAdmin } from '../../lib/auth';
import type { components } from '../../api/v1';

const { Title } = Typography;

type RBACRole = components['schemas']['rbac_role'];

function Roles() {
  const navigate = useNavigate();
  const { slug } = useWorkspace();
  const { user: currentUser } = useAuth();
  const [loading, setLoading] = useState(true);
  const [roles, setRoles] = useState<RBACRole[]>([]);
  const [scopeFilter, setScopeFilter] = useState<'global' | 'workspace' | undefined>(undefined);

  const loadRoles = useCallback(async () => {
    setLoading(true);
    const { data, error } = await client.GET('/rbac/role', {
      params: { query: scopeFilter ? { scope: scopeFilter } : {} },
    });
    if (error) {
      message.error('Failed to load roles');
      setLoading(false);
      return;
    }
    setRoles(data || []);
    setLoading(false);
  }, [scopeFilter]);

  useEffect(() => {
    if (isAdmin(currentUser)) {
      loadRoles();
    }
  }, [loadRoles, currentUser]);

  const handleDelete = async (uuid: string) => {
    const { error } = await client.DELETE('/rbac/role/{uuid}', {
      params: { path: { uuid } },
    });
    if (error) {
      message.error('Failed to delete role');
      return;
    }
    message.success('Role deleted');
    loadRoles();
  };

  const columns: ColumnsType<RBACRole> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (value: string, record) => (
        <Space>
          {value}
          {record.is_system && <Tag color="blue">System</Tag>}
        </Space>
      ),
    },
    {
      title: 'Display Name',
      dataIndex: 'display_name',
      key: 'display_name',
    },
    {
      title: 'Scope',
      dataIndex: 'scope',
      key: 'scope',
      render: (value: string) => (
        <Tag color={value === 'global' ? 'purple' : 'green'}>
          {value === 'global' ? 'Global' : 'Workspace'}
        </Tag>
      ),
    },
    {
      title: 'Permissions',
      key: 'permissions',
      render: (_, record) => record.permissions?.length || 0,
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 120,
      render: (_, record) => (
        <Space>
          {record.is_system ? (
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => navigate(`/w/${slug}/rbac/roles/${record.uuid}`)}
              title="View"
            />
          ) : (
            <>
              <Button
                type="text"
                icon={<EditOutlined />}
                onClick={() => navigate(`/w/${slug}/rbac/roles/${record.uuid}`)}
                title="Edit"
              />
              <Popconfirm
                title="Delete role"
                description="Are you sure you want to delete this role?"
                onConfirm={() => handleDelete(record.uuid!)}
                okButtonProps={{ danger: true }}
                okText="Delete"
              >
                <Button type="text" danger icon={<DeleteOutlined />} title="Delete" />
              </Popconfirm>
            </>
          )}
        </Space>
      ),
    },
  ];

  // Admin access check - after hooks
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

  return (
    <>
      <Space
        style={{
          marginBottom: 16,
          display: 'flex',
          justifyContent: 'space-between',
        }}
      >
        <Space>
          <Title level={4} style={{ margin: 0 }}>
            Roles
          </Title>
          <Select
            placeholder="Filter by scope"
            allowClear
            style={{ width: 150 }}
            value={scopeFilter}
            onChange={(value) => setScopeFilter(value)}
            options={[
              { label: 'All Scopes', value: undefined },
              { label: 'Global', value: 'global' },
              { label: 'Workspace', value: 'workspace' },
            ]}
          />
        </Space>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate(`/w/${slug}/rbac/roles/new`)}
        >
          Create Role
        </Button>
      </Space>
      <Table
        columns={columns}
        dataSource={roles}
        rowKey="uuid"
        loading={loading}
        pagination={false}
      />
    </>
  );
}

export default Roles;
