import { useNavigate } from 'react-router-dom'
import { Button, Table, Space, Typography } from 'antd'
import { PlusOutlined, EditOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

import { useApiGet } from '@/api/hooks'
import { FullLayout } from '@/layouts/FullLayout'

interface OAuth2Row {
  uuid: string
  name: string
  provider: string
}

export function OAuth2Credentials() {
  const navigate = useNavigate()

  const { data, error, isLoading } = useApiGet<{ clients: OAuth2Row[] }>('/oauth2/client')

  const rows = data?.clients ?? []

  const columns: ColumnsType<OAuth2Row> = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Provider', dataIndex: 'provider', key: 'provider', width: 150 },
    {
      title: 'Actions',
      key: 'actions',
      width: 80,
      render: (_, record) => (
        <Button type="text" icon={<EditOutlined />} onClick={() => navigate('/oauth2/credentials/' + record.uuid)} />
      ),
    },
  ]

  if (error) {
    return (
      <FullLayout>
        <div style={{ padding: 40 }}>
          <Typography.Text type="danger">Failed to load OAuth2 credentials. Please try again later.</Typography.Text>
        </div>
      </FullLayout>
    )
  }

  return (
    <FullLayout>
      <div style={{ padding: 24 }}>
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/oauth2/credentials/add')}>
            Add OAuth2 Credentials
          </Button>
          <Table<OAuth2Row>
            columns={columns}
            dataSource={rows}
            loading={isLoading}
            rowKey="uuid"
            pagination={false}
            style={{ maxWidth: 580 }}
          />
        </Space>
      </div>
    </FullLayout>
  )
}
