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

type UsageLimit = components['schemas']['usage_limit'];

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

function UsageLimits() {
  const navigate = useNavigate();
  const { slug } = useWorkspace();
  const { user: currentUser } = useAuth();
  const [loading, setLoading] = useState(true);
  const [usageLimits, setUsageLimits] = useState<UsageLimit[]>([]);
  const [policySetFilter, setPolicySetFilter] = useState<string | undefined>(undefined);

  const loadUsageLimits = useCallback(async () => {
    setLoading(true);
    const { data, error } = await client.GET('/access/usage-limits', {
      params: { query: policySetFilter ? { policy_set_name: policySetFilter } : {} },
    });
    if (error) {
      message.error('Failed to load usage limits');
      setLoading(false);
      return;
    }
    setUsageLimits(data || []);
    setLoading(false);
  }, [policySetFilter]);

  useEffect(() => {
    if (isAdmin(currentUser)) {
      loadUsageLimits();
    }
  }, [loadUsageLimits, currentUser]);

  const handleDelete = async (uuid: string) => {
    const { error } = await client.DELETE('/access/usage-limits/{uuid}', {
      params: { path: { uuid } },
    });
    if (error) {
      message.error('Failed to delete usage limit');
      return;
    }
    message.success('Usage limit deleted');
    loadUsageLimits();
  };

  const columns: ColumnsType<UsageLimit> = [
    {
      title: 'Policy Set',
      dataIndex: 'policy_set_name',
      key: 'policy_set_name',
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
      render: (value: string) => resetPeriodLabels[value] || value,
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
          title="Delete usage limit"
          description="Are you sure you want to delete this usage limit?"
          onConfirm={() => handleDelete(record.uuid!)}
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

  // Get unique policy set names for filter
  const policySetOptions = [...new Set(usageLimits.map((ul) => ul.policy_set_name))].map((name) => ({
    label: name,
    value: name,
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
            Usage Limits
          </Title>
          <Select
            placeholder="Filter by policy set"
            allowClear
            style={{ width: 200 }}
            value={policySetFilter}
            onChange={(value) => setPolicySetFilter(value)}
            options={[{ label: 'All Policy Sets', value: undefined }, ...policySetOptions]}
          />
        </Space>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate(`/w/${slug}/access/usage-limits/new`)}
        >
          Create Usage Limit
        </Button>
      </Space>
      <Table
        columns={columns}
        dataSource={usageLimits}
        rowKey="uuid"
        loading={loading}
        pagination={false}
        onRow={(record) => ({
          onClick: () => navigate(`/w/${slug}/access/usage-limits/${record.uuid}`),
          style: { cursor: 'pointer' },
        })}
      />
    </>
  );
}

export default UsageLimits;
