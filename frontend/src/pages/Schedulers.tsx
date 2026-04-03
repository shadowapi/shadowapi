import { useNavigate } from 'react-router-dom'
import { Button, Table, Tag, Space, Typography } from 'antd'
import { PlusOutlined, EditOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

import { useApiGet } from '@/api/hooks'
import { FullLayout } from '@/layouts/FullLayout'

interface SchedulerRow {
  uuid: string
  pipeline_uuid: string
  schedule_type: string
  cron_expression: string | null
  run_at: string | null
  timezone: string
  next_run: string | null
  last_run: string | null
  is_enabled: boolean
  is_paused: boolean
}

export function Schedulers() {
  const navigate = useNavigate()

  const { data, error, isLoading } = useApiGet<SchedulerRow[]>('/scheduler')

  const rows = data ?? []

  const columns: ColumnsType<SchedulerRow> = [
    {
      title: 'Pipeline UUID',
      dataIndex: 'pipeline_uuid',
      key: 'pipeline_uuid',
      render: (uuid: string) => (
        <Button type="link" size="small" onClick={() => navigate('/pipelines/' + uuid)}>
          ...{uuid.slice(-8)}
        </Button>
      ),
    },
    { title: 'Schedule Type', dataIndex: 'schedule_type', key: 'schedule_type' },
    {
      title: 'Cron Expression',
      dataIndex: 'cron_expression',
      key: 'cron_expression',
      render: (val: string | null) => val || 'N/A',
    },
    {
      title: 'Run At',
      dataIndex: 'run_at',
      key: 'run_at',
      render: (val: string | null) => (val ? new Date(val).toLocaleString() : 'N/A'),
    },
    { title: 'Timezone', dataIndex: 'timezone', key: 'timezone' },
    {
      title: 'Next Run',
      dataIndex: 'next_run',
      key: 'next_run',
      render: (val: string | null) => (val ? new Date(val).toLocaleString() : 'N/A'),
    },
    {
      title: 'Last Run',
      dataIndex: 'last_run',
      key: 'last_run',
      render: (val: string | null) => (val ? new Date(val).toLocaleString() : 'N/A'),
    },
    {
      title: 'Enabled',
      dataIndex: 'is_enabled',
      key: 'is_enabled',
      render: (val: boolean) => <Tag color={val ? 'success' : 'error'}>{val ? 'Enable' : 'Disable'}</Tag>,
    },
    {
      title: 'Paused',
      key: 'is_paused',
      render: (_, record) => (
        <Tag color={record.is_paused ? 'success' : 'error'}>{record.is_enabled ? 'Running' : 'Paused'}</Tag>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 80,
      render: (_, record) => (
        <Button type="text" icon={<EditOutlined />} onClick={() => navigate('/schedulers/' + record.uuid)} />
      ),
    },
  ]

  if (error) {
    return (
      <FullLayout>
        <div style={{ padding: 40 }}>
          <Typography.Text type="danger">Failed to load schedulers. Please try again later.</Typography.Text>
        </div>
      </FullLayout>
    )
  }

  return (
    <FullLayout>
      <div style={{ padding: 24 }}>
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/schedulers/add')}>
            Add Scheduler
          </Button>
          <Table<SchedulerRow>
            columns={columns}
            dataSource={rows}
            loading={isLoading}
            rowKey="uuid"
            pagination={false}
            style={{ maxWidth: 1400 }}
          />
        </Space>
      </div>
    </FullLayout>
  )
}
