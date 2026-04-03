import { useNavigate } from 'react-router-dom'
import { Button, Table, Space, Typography } from 'antd'
import { PlusOutlined, EditOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

import { useApiGet } from '@/api/hooks'
import { FullLayout } from '@/layouts/FullLayout'

interface PipelineRow {
  uuid: string
  name: string
  type: string
  datasource_uuid: string
  storage_uuid: string
}

export function Pipelines() {
  const navigate = useNavigate()

  const { data, error, isLoading } = useApiGet<{ pipelines: PipelineRow[] }>('/pipeline')

  const rows = data?.pipelines ?? []

  const columns: ColumnsType<PipelineRow> = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Type', dataIndex: 'type', key: 'type' },
    {
      title: 'Data Source',
      dataIndex: 'datasource_uuid',
      key: 'datasource_uuid',
      render: (uuid: string) => (
        <Button type="link" size="small" onClick={() => navigate('/datasources/' + uuid)}>
          {uuid}
        </Button>
      ),
    },
    {
      title: 'Storage',
      dataIndex: 'storage_uuid',
      key: 'storage_uuid',
      render: (uuid: string) => (
        <Button type="link" size="small" onClick={() => navigate('/storages/' + uuid + '/storageKind/hostfiles')}>
          {uuid}
        </Button>
      ),
    },
    { title: 'Schedulers', key: 'schedulers', render: () => '-' },
    {
      title: 'Actions',
      key: 'actions',
      width: 80,
      render: (_, record) => (
        <Button type="text" icon={<EditOutlined />} onClick={() => navigate('/pipelines/' + record.uuid)} />
      ),
    },
  ]

  if (error) {
    return (
      <FullLayout>
        <div style={{ padding: 40 }}>
          <Typography.Text type="danger">Failed to load pipelines. Please try again later.</Typography.Text>
        </div>
      </FullLayout>
    )
  }

  return (
    <FullLayout>
      <div style={{ padding: 24 }}>
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/pipelines/add')}>
            Add Data Pipeline
          </Button>
          <Table<PipelineRow>
            columns={columns}
            dataSource={rows}
            loading={isLoading}
            rowKey="uuid"
            pagination={false}
            style={{ maxWidth: 1000 }}
          />
        </Space>
      </div>
    </FullLayout>
  )
}
