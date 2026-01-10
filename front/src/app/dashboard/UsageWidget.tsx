import { useEffect, useState, useCallback } from 'react';
import { Card, Progress, Typography, Space, Row, Col, Tag, Tooltip, Spin } from 'antd';
import { ThunderboltOutlined, CloudDownloadOutlined, CloudUploadOutlined } from '@ant-design/icons';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import { useAuth } from '../../lib/auth';

const { Text } = Typography;

interface LimitInfo {
  limitValue: number | null;
  currentUsage: number;
  remaining: number | null;
  resetPeriod: string;
  periodEnd: string;
  isLimited: boolean;
}

const resetPeriodLabels: Record<string, string> = {
  daily: 'daily',
  weekly: 'weekly',
  monthly: 'monthly',
  rolling_24h: 'rolling 24h',
  rolling_7d: 'rolling 7 days',
  rolling_30d: 'rolling 30 days',
};

function getProgressStatus(percent: number): 'success' | 'normal' | 'exception' {
  if (percent >= 100) return 'exception';
  if (percent >= 80) return 'normal';
  return 'success';
}

function getProgressColor(percent: number): string {
  if (percent >= 100) return '#ff4d4f';
  if (percent >= 80) return '#faad14';
  return '#52c41a';
}

interface UsageCardProps {
  title: string;
  icon: React.ReactNode;
  info: LimitInfo | null;
  loading: boolean;
}

function UsageCard({ title, icon, info, loading }: UsageCardProps) {
  if (loading) {
    return (
      <Card size="small" style={{ height: '100%' }}>
        <div style={{ textAlign: 'center', padding: 20 }}>
          <Spin size="small" />
        </div>
      </Card>
    );
  }

  if (!info) {
    return (
      <Card size="small" style={{ height: '100%' }}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Space>
            {icon}
            <Text strong>{title}</Text>
          </Space>
          <Text type="secondary">No limits configured</Text>
        </Space>
      </Card>
    );
  }

  if (!info.isLimited || info.limitValue === null) {
    return (
      <Card size="small" style={{ height: '100%' }}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Space>
            {icon}
            <Text strong>{title}</Text>
          </Space>
          <Space>
            <Tag color="purple">Unlimited</Tag>
            <Text type="secondary">{info.currentUsage.toLocaleString()} used</Text>
          </Space>
        </Space>
      </Card>
    );
  }

  const percent = Math.round((info.currentUsage / info.limitValue) * 100);
  const periodEndDate = info.periodEnd ? new Date(info.periodEnd) : null;
  const remainingText = info.remaining === 0
    ? 'Limit reached'
    : `${info.remaining?.toLocaleString()} remaining`;

  return (
    <Card size="small" style={{ height: '100%' }}>
      <Space direction="vertical" style={{ width: '100%' }}>
        <Space>
          {icon}
          <Text strong>{title}</Text>
        </Space>
        <Progress
          percent={percent}
          status={getProgressStatus(percent)}
          strokeColor={getProgressColor(percent)}
          format={() => `${info.currentUsage.toLocaleString()} / ${info.limitValue?.toLocaleString()}`}
        />
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <Text type={info.remaining === 0 ? 'danger' : 'secondary'} style={{ fontSize: 12 }}>
            {remainingText}
          </Text>
          <Tooltip title={periodEndDate ? `Resets: ${periodEndDate.toLocaleString()}` : undefined}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              Resets {resetPeriodLabels[info.resetPeriod] || info.resetPeriod}
            </Text>
          </Tooltip>
        </Space>
      </Space>
    </Card>
  );
}

function UsageWidget() {
  const { slug } = useWorkspace();
  const { user } = useAuth();
  const [loading, setLoading] = useState(true);
  const [fetchInfo, setFetchInfo] = useState<LimitInfo | null>(null);
  const [pushInfo, setPushInfo] = useState<LimitInfo | null>(null);
  const [hasWorker, setHasWorker] = useState(false);

  const loadUsage = useCallback(async () => {
    if (!user?.uuid) return;

    setLoading(true);

    try {
      // Get a worker to query usage status
      const { data: workers } = await client.GET('/workers');

      if (!workers || workers.length === 0) {
        setHasWorker(false);
        setLoading(false);
        return;
      }

      setHasWorker(true);
      const defaultWorker = workers[0];

      // Query usage for both limit types
      const [fetchRes, pushRes] = await Promise.all([
        client.GET('/access/usage-status', {
          params: {
            query: {
              user_uuid: user.uuid,
              worker_uuid: defaultWorker.uuid!,
              workspace_slug: slug,
              limit_type: 'messages_fetch',
            },
          },
        }),
        client.GET('/access/usage-status', {
          params: {
            query: {
              user_uuid: user.uuid,
              worker_uuid: defaultWorker.uuid!,
              workspace_slug: slug,
              limit_type: 'messages_push',
            },
          },
        }),
      ]);

      if (fetchRes.data?.user_limit) {
        setFetchInfo({
          limitValue: fetchRes.data.user_limit.limit_value ?? null,
          currentUsage: fetchRes.data.user_limit.current_usage || 0,
          remaining: fetchRes.data.user_limit.remaining ?? null,
          resetPeriod: fetchRes.data.user_limit.reset_period || 'monthly',
          periodEnd: fetchRes.data.user_limit.period_end || '',
          isLimited: fetchRes.data.user_limit.is_limited || false,
        });
      }

      if (pushRes.data?.user_limit) {
        setPushInfo({
          limitValue: pushRes.data.user_limit.limit_value ?? null,
          currentUsage: pushRes.data.user_limit.current_usage || 0,
          remaining: pushRes.data.user_limit.remaining ?? null,
          resetPeriod: pushRes.data.user_limit.reset_period || 'monthly',
          periodEnd: pushRes.data.user_limit.period_end || '',
          isLimited: pushRes.data.user_limit.is_limited || false,
        });
      }
    } catch {
      // Silently fail - widget is not critical
    } finally {
      setLoading(false);
    }
  }, [user?.uuid, slug]);

  useEffect(() => {
    loadUsage();
  }, [loadUsage]);

  // If no worker is registered, show a minimal state
  if (!loading && !hasWorker) {
    return null; // Don't show widget if no workers
  }

  // If no limits are configured, show nothing
  if (!loading && !fetchInfo && !pushInfo) {
    return null;
  }

  return (
    <Card
      title={
        <Space>
          <ThunderboltOutlined />
          <span>Usage Quota</span>
        </Space>
      }
      size="small"
      style={{ marginBottom: 16 }}
    >
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12}>
          <UsageCard
            title="Messages Fetch"
            icon={<CloudDownloadOutlined style={{ color: '#1890ff' }} />}
            info={fetchInfo}
            loading={loading}
          />
        </Col>
        <Col xs={24} sm={12}>
          <UsageCard
            title="Messages Push"
            icon={<CloudUploadOutlined style={{ color: '#52c41a' }} />}
            info={pushInfo}
            loading={loading}
          />
        </Col>
      </Row>
    </Card>
  );
}

export default UsageWidget;
