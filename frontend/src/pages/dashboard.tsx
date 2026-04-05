import { useNavigate } from 'react-router-dom'
import { Card, Col, Row, Statistic, Steps, Button, Typography } from 'antd'
import {
  RocketOutlined,
  DatabaseOutlined,
  CloudServerOutlined,
  CheckCircleOutlined,
  PlusOutlined,
  PlayCircleOutlined,
  TrophyOutlined,
} from '@ant-design/icons'
import Markdown from 'react-markdown'

import { useApiGet } from '@/api/hooks'
import { FullLayout } from '@/layouts/FullLayout'
import guideContent from '@/notes/how_setup_pipeline.md?raw'

interface PipelineRow {
  uuid: string
  name: string
}

interface WorkerJob {
  uuid: string
  status: string
}

export function Dashboard() {
  const navigate = useNavigate()

  const { data: pipelineData, isLoading: loadingPipelines } = useApiGet<{ pipelines: PipelineRow[] }>('/pipeline')
  const { data: storageData, isLoading: loadingStorages } = useApiGet<any[]>('/storage')
  const { data: datasourceData, isLoading: loadingDatasources } = useApiGet<any[]>('/datasource')
  const { data: workerData, isLoading: loadingWorkers } = useApiGet<{ worker_jobs: WorkerJob[] }>(
    '/workerjobs?limit=100',
  )

  const pipelines = pipelineData?.pipelines ?? []
  const storages = storageData ?? []
  const datasources = datasourceData ?? []
  const workerJobs = workerData?.worker_jobs ?? []
  const runningJobs = workerJobs.filter((j) => j.status === 'running')

  const hasStorages = storages.length > 0
  const hasDatasources = datasources.length > 0
  const hasPipelines = pipelines.length > 0
  const hasRanPipeline = workerJobs.length > 0
  const allComplete = hasStorages && hasDatasources && hasPipelines && hasRanPipeline

  const loading = loadingPipelines || loadingStorages || loadingDatasources || loadingWorkers

  const currentStep = !hasStorages ? 0 : !hasDatasources ? 1 : !hasPipelines ? 2 : !hasRanPipeline ? 3 : 4

  return (
    <FullLayout>
      <div style={{ padding: 24 }}>
        {/* WIDGET AREA */}
        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={8}>
            <Card hoverable onClick={() => navigate('/pipelines')} style={{ cursor: 'pointer' }}>
              <Statistic
                title="Active Pipelines"
                value={pipelines.length}
                prefix={<RocketOutlined />}
                loading={loading}
              />
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card hoverable onClick={() => navigate('/storages')} style={{ cursor: 'pointer' }}>
              <Statistic
                title="Storages"
                value={storages.length}
                prefix={<DatabaseOutlined />}
                loading={loading}
              />
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card hoverable onClick={() => navigate('/workers')} style={{ cursor: 'pointer' }}>
              <Statistic
                title="Running Jobs"
                value={runningJobs.length}
                prefix={<CloudServerOutlined />}
                loading={loading}
              />
            </Card>
          </Col>
        </Row>

        {/* QUICK START AREA — shown until all steps complete */}
        {!loading && !allComplete && (
          <Card title="Quick Start" style={{ maxWidth: 1100 }}>
            <Row gutter={[32, 24]}>
              {/* Left column — markdown guide */}
              <Col xs={24} lg={12}>
                <div className="markdown-guide">
                  <Markdown>{guideContent}</Markdown>
                </div>
              </Col>

              {/* Right column — onboarding checklist */}
              <Col xs={24} lg={12}>
                <Typography.Title level={5} style={{ marginBottom: 16 }}>
                  Setup Progress
                </Typography.Title>
                <Steps
                  direction="vertical"
                  current={currentStep}
                  items={[
                    {
                      title: 'Create a Storage',
                      description: hasStorages ? (
                        <span>
                          <CheckCircleOutlined style={{ color: '#52c41a', marginRight: 6 }} />
                          {storages.length} storage{storages.length > 1 ? 's' : ''} configured
                        </span>
                      ) : (
                        <Button
                          size="small"
                          type="primary"
                          icon={<PlusOutlined />}
                          onClick={() => navigate('/storages/add')}
                        >
                          Add Storage
                        </Button>
                      ),
                    },
                    {
                      title: 'Add a Data Source',
                      description: hasDatasources ? (
                        <span>
                          <CheckCircleOutlined style={{ color: '#52c41a', marginRight: 6 }} />
                          {datasources.length} data source{datasources.length > 1 ? 's' : ''} configured
                        </span>
                      ) : (
                        <Button
                          size="small"
                          type="primary"
                          icon={<PlusOutlined />}
                          onClick={() => navigate('/datasources/add')}
                        >
                          Add Data Source
                        </Button>
                      ),
                    },
                    {
                      title: 'Create a Pipeline',
                      description: hasPipelines ? (
                        <span>
                          <CheckCircleOutlined style={{ color: '#52c41a', marginRight: 6 }} />
                          {pipelines.length} pipeline{pipelines.length > 1 ? 's' : ''} created
                        </span>
                      ) : hasStorages && hasDatasources ? (
                        <Button
                          size="small"
                          type="primary"
                          icon={<PlusOutlined />}
                          onClick={() => navigate('/pipelines/add')}
                        >
                          Create Pipeline
                        </Button>
                      ) : (
                        'Complete the steps above first'
                      ),
                    },
                    {
                      title: 'Run Your Pipeline',
                      icon: hasPipelines && !hasRanPipeline ? <PlayCircleOutlined /> : undefined,
                      description: hasRanPipeline ? (
                        <span>
                          <CheckCircleOutlined style={{ color: '#52c41a', marginRight: 6 }} />
                          {workerJobs.length} job{workerJobs.length > 1 ? 's' : ''} executed
                        </span>
                      ) : hasPipelines ? (
                        <Button
                          size="small"
                          type="primary"
                          icon={<PlayCircleOutlined />}
                          onClick={() => navigate('/pipelines')}
                        >
                          Go to Pipelines
                        </Button>
                      ) : (
                        'Create a pipeline first'
                      ),
                    },
                  ]}
                />
              </Col>
            </Row>
          </Card>
        )}

        {/* SUCCESS — shown when all steps complete */}
        {!loading && allComplete && (
          <Card style={{ maxWidth: 1100, textAlign: 'center', padding: '32px 0' }}>
            <TrophyOutlined style={{ fontSize: 48, color: '#52c41a', marginBottom: 16 }} />
            <Typography.Title level={3} style={{ color: '#52c41a', marginBottom: 8 }}>
              Setup Complete!
            </Typography.Title>
            <Typography.Text type="secondary" style={{ fontSize: 16 }}>
              Your pipeline is configured and running. Monitor your jobs in the{' '}
              <a onClick={() => navigate('/workers')} style={{ cursor: 'pointer' }}>
                Workers
              </a>{' '}
              section.
            </Typography.Text>
          </Card>
        )}
      </div>
    </FullLayout>
  )
}
