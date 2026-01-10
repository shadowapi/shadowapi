import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router';
import { Table, Button, Space, Typography, message, Tag, Result, Popconfirm, Select, Switch } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import { useAuth, isAdmin } from '../../lib/auth';
import type { components } from '../../api/v1';

const { Title } = Typography;

type UserUsageLimitOverride = components['schemas']['user_usage_limit_override'];
type User = components['schemas']['user'];

const limitTypeLabels: Record<string, string> = {
  messages_fetch: 'Messages Fetch',
  messages_push: 'Messages Push',
};

const resetPeriodLabels: Record<string, string> = {
  daily: 'Daily',
  weekly: 'Weekly',
  monthly: 'Monthly',
  rolling_24h: 'Rolling 24h',
  rolling_7d: 'Rolling 7d',
  rolling_30d: 'Rolling 30d',
};

function UserUsageLimits() {
  const navigate = useNavigate();
  const { slug } = useWorkspace();
  const { user: currentUser } = useAuth();
  const [loading, setLoading] = useState(true);
  const [overrides, setOverrides] = useState<UserUsageLimitOverride[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [userFilter, setUserFilter] = useState<string | undefined>(undefined);

  const loadData = useCallback(async () => {
    setLoading(true);

    // Load users for display
    const usersRes = await client.GET('/user');
    if (usersRes.data) {
      setUsers(usersRes.data);
    }

    // Load overrides for all users (we'll filter client-side or by iterating)
    // Note: The API requires user_uuid, so we need to load for each user
    // For simplicity, we'll show all users and their overrides
    const allOverrides: UserUsageLimitOverride[] = [];

    if (usersRes.data) {
      for (const user of usersRes.data) {
        if (!user.uuid) continue;
        const { data } = await client.GET('/access/user/{user_uuid}/usage-limits', {
          params: {
            path: { user_uuid: user.uuid },
            query: { workspace_slug: slug },
          },
        });
        if (data) {
          allOverrides.push(...data);
        }
      }
    }

    setOverrides(allOverrides);
    setLoading(false);
  }, [slug]);

  useEffect(() => {
    if (isAdmin(currentUser)) {
      loadData();
    }
  }, [loadData, currentUser]);

  const handleDelete = async (userUuid: string, uuid: string) => {
    const { error } = await client.DELETE('/access/user/{user_uuid}/usage-limits/{uuid}', {
      params: { path: { user_uuid: userUuid, uuid } },
    });
    if (error) {
      message.error('Failed to delete user usage limit override');
      return;
    }
    message.success('User usage limit override deleted');
    loadData();
  };

  // Filter overrides by user
  const filteredOverrides = userFilter
    ? overrides.filter((o) => o.user_uuid === userFilter)
    : overrides;

  // Create user lookup
  const userMap = new Map(users.map((u) => [u.uuid, u]));

  const columns: ColumnsType<UserUsageLimitOverride> = [
    {
      title: 'User',
      dataIndex: 'user_uuid',
      key: 'user_uuid',
      render: (uuid: string) => {
        const user = userMap.get(uuid);
        return user?.email || uuid;
      },
    },
    {
      title: 'Workspace',
      dataIndex: 'workspace_slug',
      key: 'workspace_slug',
    },
    {
      title: 'Limit Type',
      dataIndex: 'limit_type',
      key: 'limit_type',
      render: (value: string) => (
        <Tag color={value === 'messages_fetch' ? 'blue' : 'green'}>
          {limitTypeLabels[value] || value}
        </Tag>
      ),
    },
    {
      title: 'Limit Value',
      dataIndex: 'limit_value',
      key: 'limit_value',
      render: (value: number | null) =>
        value === null ? <Tag color="purple">Unlimited</Tag> : value.toLocaleString(),
    },
    {
      title: 'Reset Period',
      dataIndex: 'reset_period',
      key: 'reset_period',
      render: (value: string | null) =>
        value ? resetPeriodLabels[value] || value : <Tag>Inherited</Tag>,
    },
    {
      title: 'Enabled',
      dataIndex: 'is_enabled',
      key: 'is_enabled',
      render: (value: boolean) => <Switch checked={value} disabled size="small" />,
    },
    {
      title: '',
      key: 'actions',
      width: 60,
      render: (_, record) => (
        <Popconfirm
          title="Delete override"
          description="Are you sure you want to delete this user usage limit override?"
          onConfirm={() => handleDelete(record.user_uuid, record.uuid!)}
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

  // Get unique users for filter
  const userOptions = users.map((u) => ({
    label: u.email,
    value: u.uuid,
  }));

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
            User Usage Overrides
          </Title>
          <Select
            placeholder="Filter by user"
            allowClear
            style={{ width: 250 }}
            value={userFilter}
            onChange={(value) => setUserFilter(value)}
            options={[{ label: 'All Users', value: undefined }, ...userOptions]}
            showSearch
            filterOption={(input, option) =>
              (option?.label as string)?.toLowerCase().includes(input.toLowerCase())
            }
          />
        </Space>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate(`/w/${slug}/access/user-usage-limits/new`)}
        >
          Create Override
        </Button>
      </Space>
      <Table
        columns={columns}
        dataSource={filteredOverrides}
        rowKey="uuid"
        loading={loading}
        pagination={false}
        onRow={(record) => ({
          onClick: () => navigate(`/w/${slug}/access/user-usage-limits/${record.uuid}?user=${record.user_uuid}`),
          style: { cursor: 'pointer' },
        })}
      />
    </>
  );
}

export default UserUsageLimits;
