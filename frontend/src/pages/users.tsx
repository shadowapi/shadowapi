import { useNavigate } from 'react-router-dom'
import { Button, Table, Space, Tag, Typography } from 'antd'
import { PlusOutlined, EditOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

import { useApiGet } from '@/api/hooks'
import { FullLayout } from '@/layouts/FullLayout'

interface UserRow {
  uuid: string
  email: string
  first_name: string
  last_name: string
  is_enabled: boolean
  is_admin: boolean
  created_at: string
  updated_at: string
}

export function Users() {
  const navigate = useNavigate()

  const { data, error, isLoading } = useApiGet<UserRow[]>('/user')

  const rows = data ?? []

  const columns: ColumnsType<UserRow> = [
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
    },
    {
      title: 'Name',
      key: 'name',
      render: (_, record) => `${record.first_name} ${record.last_name}`.trim() || '—',
    },
    {
      title: 'Enabled',
      dataIndex: 'is_enabled',
      key: 'is_enabled',
      width: 90,
      render: (val: boolean) => <Tag color={val ? 'green' : 'red'}>{val ? 'Yes' : 'No'}</Tag>,
    },
    {
      title: 'Admin',
      dataIndex: 'is_admin',
      key: 'is_admin',
      width: 80,
      render: (val: boolean) => val ? <Tag color="blue">Admin</Tag> : '—',
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (val: string) => val ? new Date(val).toLocaleDateString() : '—',
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 80,
      render: (_, record) => (
        <Button type="text" icon={<EditOutlined />} onClick={() => navigate('/users/' + record.uuid)} />
      ),
    },
  ]

  if (error) {
    return (
      <FullLayout>
        <div style={{ padding: 40 }}>
          <Typography.Text type="danger">Failed to load users. Please try again later.</Typography.Text>
        </div>
      </FullLayout>
    )
  }

  return (
    <FullLayout>
      <div style={{ padding: 24 }}>
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/users/add')}>
            Add User
          </Button>
          <Table<UserRow>
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
