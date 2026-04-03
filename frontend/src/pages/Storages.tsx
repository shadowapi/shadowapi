import { useNavigate } from 'react-router-dom'
import { Button, Table, Tag, Space } from 'antd'
import { PlusOutlined, EditOutlined, DatabaseOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

import { useApiGet } from '@/api/hooks'
import { useTitle } from '@/hooks'
import { FullLayout } from '@/layouts/FullLayout'

interface StorageRow {
  id: string
  type: string
  name: string
  state: 'Enabled' | 'Disabled'
}

export function Storages() {
  const navigate = useNavigate()
  const pageTitle = 'Storages'
  useTitle(pageTitle)

  const { data, isLoading } = useApiGet<any[]>('/storage')

  const rows: StorageRow[] =
    data?.map((item: any) => ({
      id: item.uuid,
      type: item.type,
      name: item.name,
      state: item.is_enabled ? 'Enabled' : 'Disabled',
    })) ?? []

  const typeLabel: Record<string, string> = {
    postgres: 'PostgreSQL',
    s3: 'S3',
    hostfiles: 'Hostfiles',
  }

  const columns: ColumnsType<StorageRow> = [
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      width: 160,
      render: (type: string) => (
        <Tag icon={<DatabaseOutlined />}>{typeLabel[type] ?? '... missing type'}</Tag>
      ),
    },
    { title: 'Name', dataIndex: 'name', key: 'name', width: 160 },
    {
      title: 'State',
      dataIndex: 'state',
      key: 'state',
      width: 160,
      render: (state: string) => <Tag color={state === 'Enabled' ? 'success' : 'error'}>{state}</Tag>,
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 80,
      render: (_, record) => (
        <Button
          type="text"
          icon={<EditOutlined />}
          onClick={() => navigate('/storages/' + record.id + '/storageKind/' + record.type)}
        />
      ),
    },
  ]

  return (
    <FullLayout>
      <div style={{ padding: 24 }}>
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/storages/add')}>
            Add Storage
          </Button>
          <Table<StorageRow>
            columns={columns}
            dataSource={rows}
            loading={isLoading}
            rowKey="id"
            pagination={false}
            style={{ maxWidth: 1000 }}
          />
        </Space>
      </div>
    </FullLayout>
  )
}
