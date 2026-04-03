import { useNavigate } from 'react-router-dom'
import { Button, Table, Space, Typography } from 'antd'
import { PlusOutlined, EditOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

import { useApiGet } from '@/api/hooks'
import { FullLayout } from '@/layouts/FullLayout'

interface PolicyRow {
  uuid: string
  name: string
  type: string
  sync_all: boolean
  created_at: string
}

export function SyncPolicies() {
  const navigate = useNavigate()

  const { data, error, isLoading } = useApiGet<{ policies: PolicyRow[] }>('/syncpolicy')

  const rows = data?.policies ?? []

  const columns: ColumnsType<PolicyRow> = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Type', dataIndex: 'type', key: 'type' },
    {
      title: 'Sync All',
      dataIndex: 'sync_all',
      key: 'sync_all',
      render: (val: boolean) => (val ? 'Yes' : 'No'),
    },
    { title: 'Created At', dataIndex: 'created_at', key: 'created_at' },
    {
      title: 'Actions',
      key: 'actions',
      width: 80,
      render: (_, record) => (
        <Button type="text" icon={<EditOutlined />} onClick={() => navigate('/syncpolicy/' + record.uuid)} />
      ),
    },
  ]

  if (error) {
    return (
      <FullLayout>
        <div style={{ padding: 40 }}>
          <Typography.Text type="danger">Failed to load policies. Please try again later.</Typography.Text>
        </div>
      </FullLayout>
    )
  }

  return (
    <FullLayout>
      <div style={{ padding: 24 }}>
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/syncpolicy/add')}>
            Add Policy
          </Button>
          <Table<PolicyRow>
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
