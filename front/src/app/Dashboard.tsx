import { useEffect, useState, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router';
import {
  Row,
  Col,
  Card,
  Statistic,
  List,
  Tag,
  Typography,
  Badge,
  Alert,
  Progress,
  Spin,
  Empty,
  message,
} from 'antd';
import {
  DatabaseOutlined,
  NodeIndexOutlined,
  CloudServerOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  SyncOutlined,
  WarningOutlined,
} from '@ant-design/icons';
import { Link } from 'react-router';
import client from '../api/client';
import { useWorkspace } from '../lib/workspace/WorkspaceContext';
import type { components } from '../api/v1';

const { Title, Text } = Typography;

type Datasource = components['schemas']['datasource'];
type Pipeline = components['schemas']['pipeline'];
type Storage = components['schemas']['storage'];
type Scheduler = components['schemas']['scheduler'];
type WorkerJob = components['schemas']['worker_jobs'];

// Job status configuration
const jobStatusConfig: Record<string, { color: string; icon: React.ReactNode }> = {
  running: { color: 'processing', icon: <SyncOutlined spin /> },
  completed: { color: 'success', icon: <CheckCircleOutlined /> },
  failed: { color: 'error', icon: <CloseCircleOutlined /> },
  retry: { color: 'warning', icon: <WarningOutlined /> },
  pending: { color: 'default', icon: <ClockCircleOutlined /> },
};

// Format relative time
function formatTimeAgo(dateStr?: string): string {
  if (!dateStr) return '-';
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `${diffHours}h ago`;
  const diffDays = Math.floor(diffHours / 24);
  return `${diffDays}d ago`;
}

// Format datetime
function formatDateTime(dateStr?: string | null): string {
  if (!dateStr) return 'Not scheduled';
  return new Date(dateStr).toLocaleString();
}

function Dashboard() {
  const navigate = useNavigate();
  const { slug } = useWorkspace();
  const [loading, setLoading] = useState(true);

  // Data state
  const [datasources, setDatasources] = useState<Datasource[]>([]);
  const [pipelines, setPipelines] = useState<Pipeline[]>([]);
  const [storages, setStorages] = useState<Storage[]>([]);
  const [schedulers, setSchedulers] = useState<Scheduler[]>([]);
  const [recentJobs, setRecentJobs] = useState<WorkerJob[]>([]);

  const loadData = useCallback(async () => {
    setLoading(true);

    try {
      // Fetch all data in parallel
      const [
        datasourcesRes,
        pipelinesRes,
        storagesRes,
        schedulersRes,
        workerJobsRes,
      ] = await Promise.all([
        client.GET('/datasource'),
        client.GET('/pipeline'),
        client.GET('/storage'),
        client.GET('/scheduler'),
        client.GET('/workerjobs', { params: { query: { limit: 10 } } }),
      ]);

      // Set data (graceful degradation if some calls fail)
      setDatasources(datasourcesRes.data || []);
      setPipelines(pipelinesRes.data?.pipelines || []);
      setStorages(storagesRes.data || []);
      setSchedulers(schedulersRes.data || []);
      setRecentJobs(workerJobsRes.data?.jobs || []);

      if (datasourcesRes.error || pipelinesRes.error) {
        message.warning('Some data failed to load');
      }
    } catch {
      message.error('Failed to load dashboard data');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // Computed dashboard stats
  const stats = useMemo(() => {
    // Datasource auth stats (only for OAuth types)
    const oauthDatasources = datasources.filter((ds) => ds.type === 'email_oauth');
    const authenticatedCount = oauthDatasources.filter(
      (ds) => ds.is_oauth_authenticated
    ).length;

    // Pipeline stats
    const enabledPipelines = pipelines.filter((p) => p.is_enabled).length;

    // Storage by type
    const storagesByType = storages.reduce(
      (acc, s) => {
        const t = s.type || 'unknown';
        acc[t] = (acc[t] || 0) + 1;
        return acc;
      },
      {} as Record<string, number>
    );

    // Job stats
    const failedJobs = recentJobs.filter((j) => j.status === 'failed');

    // Upcoming schedules (sorted by next_run)
    const upcomingSchedules = schedulers
      .filter((s) => s.is_enabled && !s.is_paused && s.next_run)
      .sort(
        (a, b) => new Date(a.next_run!).getTime() - new Date(b.next_run!).getTime()
      )
      .slice(0, 5);

    return {
      datasourcesCount: datasources.length,
      oauthCount: oauthDatasources.length,
      authenticatedCount,
      pipelinesCount: pipelines.length,
      enabledPipelines,
      storagesCount: storages.length,
      storagesByType,
      failedJobs,
      upcomingSchedules,
      hasAuthIssues: authenticatedCount < oauthDatasources.length,
    };
  }, [datasources, pipelines, storages, schedulers, recentJobs]);

  // Create pipeline lookup for scheduler display
  const pipelineMap = useMemo(() => {
    const map: Record<string, Pipeline> = {};
    for (const p of pipelines) {
      if (p.uuid) map[p.uuid] = p;
    }
    return map;
  }, [pipelines]);

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 100 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <>
      <Title level={2}>Dashboard</Title>

      {/* Overview Statistics Cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={4}>
          <Card
            hoverable
            onClick={() => navigate(`/w/${slug}/datasources`)}
            size="small"
          >
            <Statistic
              title="Data Sources"
              value={stats.datasourcesCount}
              prefix={<DatabaseOutlined />}
            />
            {stats.oauthCount > 0 && (
              <Text type="secondary" style={{ fontSize: 12 }}>
                {stats.authenticatedCount}/{stats.oauthCount} OAuth authenticated
              </Text>
            )}
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={4}>
          <Card
            hoverable
            onClick={() => navigate(`/w/${slug}/pipelines`)}
            size="small"
          >
            <Statistic
              title="Pipelines"
              value={stats.pipelinesCount}
              prefix={<NodeIndexOutlined />}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>
              {stats.enabledPipelines} enabled
            </Text>
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={4}>
          <Card
            hoverable
            onClick={() => navigate(`/w/${slug}/storages`)}
            size="small"
          >
            <Statistic
              title="Storages"
              value={stats.storagesCount}
              prefix={<CloudServerOutlined />}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>
              {Object.entries(stats.storagesByType)
                .map(([type, count]) => `${count} ${type}`)
                .join(', ') || 'None'}
            </Text>
          </Card>
        </Col>

      </Row>

      {/* Activity and Health Row */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} lg={16}>
          <Card
            title="Recent Activity"
            extra={<Link to={`/w/${slug}/workers`}>View all</Link>}
          >
            {recentJobs.length > 0 ? (
              <List
                dataSource={recentJobs}
                renderItem={(job) => {
                  const config = jobStatusConfig[job.status] || jobStatusConfig.pending;
                  return (
                    <List.Item>
                      <List.Item.Meta
                        avatar={config.icon}
                        title={job.subject}
                        description={formatTimeAgo(job.started_at)}
                      />
                      <Tag color={config.color}>{job.status}</Tag>
                    </List.Item>
                  );
                }}
              />
            ) : (
              <Empty description="No recent activity" />
            )}
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title="Health Status">
            {/* OAuth Authentication Status */}
            <div style={{ marginBottom: 16 }}>
              <Text strong>OAuth Authentication</Text>
              {stats.oauthCount > 0 ? (
                stats.hasAuthIssues ? (
                  <Alert
                    type="warning"
                    message={`${stats.oauthCount - stats.authenticatedCount} datasource(s) need authentication`}
                    showIcon
                    style={{ marginTop: 8 }}
                  />
                ) : (
                  <div style={{ marginTop: 8 }}>
                    <Badge status="success" text="All OAuth datasources authenticated" />
                  </div>
                )
              ) : (
                <div style={{ marginTop: 8 }}>
                  <Badge status="default" text="No OAuth datasources configured" />
                </div>
              )}
            </div>

            {/* Pipeline Status */}
            <div style={{ marginBottom: 16 }}>
              <Text strong>Pipelines</Text>
              {stats.pipelinesCount > 0 ? (
                <Progress
                  percent={Math.round(
                    (stats.enabledPipelines / stats.pipelinesCount) * 100
                  )}
                  format={() => `${stats.enabledPipelines}/${stats.pipelinesCount} enabled`}
                  status={
                    stats.enabledPipelines < stats.pipelinesCount ? 'active' : 'success'
                  }
                  style={{ marginTop: 8 }}
                />
              ) : (
                <div style={{ marginTop: 8 }}>
                  <Badge status="default" text="No pipelines configured" />
                </div>
              )}
            </div>

            {/* Recent Failures */}
            {stats.failedJobs.length > 0 && (
              <Alert
                type="error"
                message={`${stats.failedJobs.length} job(s) failed recently`}
                description={<Link to={`/w/${slug}/workers`}>View details</Link>}
                showIcon
              />
            )}
          </Card>
        </Col>
      </Row>

      {/* Upcoming Schedules Row */}
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <Card
            title="Upcoming Scheduled Runs"
            extra={<Link to={`/w/${slug}/schedulers`}>Manage</Link>}
          >
            {stats.upcomingSchedules.length > 0 ? (
              <List
                dataSource={stats.upcomingSchedules}
                renderItem={(scheduler) => {
                  const pipeline = pipelineMap[scheduler.pipeline_uuid];
                  return (
                    <List.Item>
                      <List.Item.Meta
                        avatar={<ClockCircleOutlined />}
                        title={pipeline?.name || scheduler.pipeline_uuid}
                        description={`Next run: ${formatDateTime(scheduler.next_run)}`}
                      />
                      <Tag color={scheduler.is_enabled ? 'green' : 'default'}>
                        {scheduler.is_enabled ? 'Active' : 'Paused'}
                      </Tag>
                    </List.Item>
                  );
                }}
              />
            ) : (
              <Empty description="No scheduled runs" />
            )}
          </Card>
        </Col>
      </Row>
    </>
  );
}

export default Dashboard;
