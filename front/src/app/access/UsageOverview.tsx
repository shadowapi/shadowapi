import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router';
import {
  Table,
  Button,
  Space,
  Typography,
  message,
  Tag,
  Result,
  Select,
  Progress,
  Card,
  Row,
  Col,
  Statistic,
  Empty,
} from 'antd';
import { ReloadOutlined, WarningOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import { useAuth, isAdmin } from '../../lib/auth';
import type { components } from '../../api/v1';

const { Title, Text } = Typography;

type User = components['schemas']['user'];
type RegisteredWorker = components['schemas']['registered_worker'];

interface UserUsageRow {
  key: string;
  userUuid: string;
  email: string;
  limitType: 'messages_fetch' | 'messages_push';
  limitValue: number | null;
  currentUsage: number;
  remaining: number | null;
  resetPeriod: string;
  periodEnd: string;
  isLimited: boolean;
}

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

function UsageOverview() {
  const navigate = useNavigate();
  const { slug } = useWorkspace();
  const { user: currentUser } = useAuth();
  const [loading, setLoading] = useState(true);
  const [users, setUsers] = useState<User[]>([]);
  const [workers, setWorkers] = useState<RegisteredWorker[]>([]);
  const [usageData, setUsageData] = useState<UserUsageRow[]>([]);
  const [limitTypeFilter, setLimitTypeFilter] = useState<'messages_fetch' | 'messages_push' | undefined>(undefined);

  const loadData = useCallback(async () => {
    setLoading(true);

    try {
      // Load users and workers
      const [usersRes, workersRes] = await Promise.all([
        client.GET('/user'),
        client.GET('/workers'),
      ]);

      const loadedUsers = usersRes.data || [];
      const loadedWorkers = workersRes.data || [];

      setUsers(loadedUsers);
      setWorkers(loadedWorkers);

      // Get usage status for each user/worker combination
      const rows: UserUsageRow[] = [];

      // We need at least one worker to query usage status
      if (loadedWorkers.length > 0 && loadedUsers.length > 0) {
        const defaultWorker = loadedWorkers[0];

        for (const user of loadedUsers) {
          if (!user.uuid) continue;

          // Query both limit types
          for (const limitType of ['messages_fetch', 'messages_push'] as const) {
            const { data: status } = await client.GET('/access/usage-status', {
              params: {
                query: {
                  user_uuid: user.uuid,
                  worker_uuid: defaultWorker.uuid!,
                  workspace_slug: slug,
                  limit_type: limitType,
                },
              },
            });

            if (status?.user_limit) {
              rows.push({
                key: `${user.uuid}-${limitType}`,
                userUuid: user.uuid,
                email: user.email || '',
                limitType,
                limitValue: status.user_limit.limit_value ?? null,
                currentUsage: status.user_limit.current_usage || 0,
                remaining: status.user_limit.remaining ?? null,
                resetPeriod: status.user_limit.reset_period || 'monthly',
                periodEnd: status.user_limit.period_end || '',
                isLimited: status.user_limit.is_limited || false,
              });
            }
          }
        }
      }

      setUsageData(rows);
    } catch {
      message.error('Failed to load usage data');
    } finally {
      setLoading(false);
    }
  }, [slug]);

  useEffect(() => {
    if (isAdmin(currentUser)) {
      loadData();
    }
  }, [loadData, currentUser]);

  // Filter by limit type
  const filteredData = limitTypeFilter
    ? usageData.filter((row) => row.limitType === limitTypeFilter)
    : usageData;

  // Calculate summary statistics
  const stats = {
    totalUsers: users.length,
    totalWorkers: workers.length,
    usersAtLimit: usageData.filter((row) => row.isLimited && row.remaining === 0).length,
    usersNearLimit: usageData.filter(
      (row) => row.isLimited && row.remaining !== null && row.limitValue !== null && row.remaining > 0 && row.remaining < row.limitValue * 0.1
    ).length,
  };

  const columns: ColumnsType<UserUsageRow> = [
    {
      title: 'User',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: 'Limit Type',
      dataIndex: 'limitType',
      key: 'limitType',
      render: (value: string) => (
        <Tag color={value === 'messages_fetch' ? 'blue' : 'green'}>
          {limitTypeLabels[value] || value}
        </Tag>
      ),
    },
    {
      title: 'Limit',
      dataIndex: 'limitValue',
      key: 'limitValue',
      render: (value: number | null) =>
        value === null ? <Tag color="purple">Unlimited</Tag> : value.toLocaleString(),
    },
    {
      title: 'Current Usage',
      dataIndex: 'currentUsage',
      key: 'currentUsage',
      render: (value: number) => value.toLocaleString(),
    },
    {
      title: 'Usage',
      key: 'progress',
      render: (_, record) => {
        if (!record.isLimited || record.limitValue === null) {
          return <Text type="secondary">Unlimited</Text>;
        }
        const percent = Math.round((record.currentUsage / record.limitValue) * 100);
        let status: 'success' | 'normal' | 'exception' = 'success';
        if (percent >= 100) status = 'exception';
        else if (percent >= 80) status = 'normal';
        return (
          <Progress
            percent={percent}
            size="small"
            status={status}
            format={() => `${record.currentUsage}/${record.limitValue}`}
          />
        );
      },
    },
    {
      title: 'Remaining',
      dataIndex: 'remaining',
      key: 'remaining',
      render: (value: number | null, record) => {
        if (!record.isLimited || value === null) {
          return <Text type="secondary">-</Text>;
        }
        if (value === 0) {
          return <Tag color="red">Exhausted</Tag>;
        }
        if (record.limitValue && value < record.limitValue * 0.1) {
          return (
            <Space>
              <WarningOutlined style={{ color: '#faad14' }} />
              <Text>{value.toLocaleString()}</Text>
            </Space>
          );
        }
        return value.toLocaleString();
      },
    },
    {
      title: 'Reset Period',
      dataIndex: 'resetPeriod',
      key: 'resetPeriod',
      render: (value: string) => resetPeriodLabels[value] || value,
    },
    {
      title: 'Period Ends',
      dataIndex: 'periodEnd',
      key: 'periodEnd',
      render: (value: string) =>
        value ? new Date(value).toLocaleDateString() : '-',
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
        <Title level={4} style={{ margin: 0 }}>
          Usage Overview
        </Title>
        <Space>
          <Select
            placeholder="Filter by limit type"
            allowClear
            style={{ width: 180 }}
            value={limitTypeFilter}
            onChange={(value) => setLimitTypeFilter(value)}
            options={[
              { label: 'All Types', value: undefined },
              { label: 'Messages Fetch', value: 'messages_fetch' },
              { label: 'Messages Push', value: 'messages_push' },
            ]}
          />
          <Button icon={<ReloadOutlined />} onClick={loadData} loading={loading}>
            Refresh
          </Button>
        </Space>
      </Space>

      {/* Summary Cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card size="small">
            <Statistic title="Total Users" value={stats.totalUsers} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card size="small">
            <Statistic title="Total Workers" value={stats.totalWorkers} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card size="small">
            <Statistic
              title="Users at Limit"
              value={stats.usersAtLimit}
              valueStyle={stats.usersAtLimit > 0 ? { color: '#ff4d4f' } : undefined}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card size="small">
            <Statistic
              title="Users Near Limit"
              value={stats.usersNearLimit}
              valueStyle={stats.usersNearLimit > 0 ? { color: '#faad14' } : undefined}
            />
          </Card>
        </Col>
      </Row>

      {usageData.length > 0 ? (
        <Table
          columns={columns}
          dataSource={filteredData}
          loading={loading}
          pagination={{ pageSize: 20 }}
        />
      ) : (
        <Card>
          <Empty
            description={
              workers.length === 0
                ? 'No workers registered. Register a worker to see usage data.'
                : 'No usage data available.'
            }
          />
        </Card>
      )}
    </>
  );
}

export default UsageOverview;
