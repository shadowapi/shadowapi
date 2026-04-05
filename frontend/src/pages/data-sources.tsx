import { useNavigate } from 'react-router-dom'
import { Button, Table, Tag, Space, Typography, message } from 'antd'
import { PlusOutlined, EditOutlined, LoginOutlined, DeleteOutlined, MailOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { mutate } from 'swr'

import apiClient from '@/api/client'
import { useApiGet } from '@/api/hooks'
import { FullLayout } from '@/layouts/FullLayout'

interface DatasourceRow {
  key: string
  name: string
  type: string
  state: 'Enabled' | 'Disabled'
}

export function DataSources() {
  const navigate = useNavigate()

  const { data, error, isLoading } = useApiGet<any[]>('/datasource')

  const rows: DatasourceRow[] =
    data?.map((ds: any) => ({
      key: ds.uuid,
      name: ds.name,
      type: ds.type,
      state: ds.is_enabled ? 'Enabled' : 'Disabled',
    })) ?? []

  const handleRevokeTokens = async (datasourceUUID: string) => {
    try {
      const listResp = await apiClient.get(`/oauth2/client/${datasourceUUID}/token`)
      const tokens = listResp.data ?? []
      for (const tok of tokens) {
        await apiClient.delete(`/oauth2/client/${datasourceUUID}/token/${tok.uuid}`)
      }
      message.success('Tokens revoked')
      mutate('/datasource')
    } catch {
      message.error('Failed to revoke tokens')
    }
  }

  const handleOauthLogin = async (row: DatasourceRow) => {
    if (row.type !== 'email_oauth') return
    try {
      const resp = await apiClient.post('/oauth2/login', {
        query: { datasource_uuid: [row.key] },
      })
      if (resp.data?.auth_code_url) {
        window.location.href = resp.data.auth_code_url
      }
    } catch {
      message.error('Failed to initiate login')
    }
  }

  const typeBadge = (type: string) => {
    if (type === 'email') return <Tag icon={<MailOutlined />}>Email IMAP</Tag>
    if (type === 'email_oauth') return <Tag icon={<MailOutlined />}>Email OAuth</Tag>
    return <Tag>{type}</Tag>
  }

  const columns: ColumnsType<DatasourceRow> = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      width: 160,
      render: (type: string) => typeBadge(type),
    },
    {
      title: 'State',
      dataIndex: 'state',
      key: 'state',
      width: 160,
      render: (state: string) => <Tag color={state === 'Enabled' ? 'success' : 'error'}>{state}</Tag>,
    },
    {
      title: 'Re/Auth',
      key: 'login',
      width: 120,
      render: (_, record) =>
        record.type === 'email_oauth' ? (
          <Button type="text" icon={<LoginOutlined />} onClick={() => handleOauthLogin(record)} />
        ) : (
          '-'
        ),
    },
    {
      title: 'Revoke',
      key: 'revoke',
      width: 120,
      render: (_, record) =>
        record.type === 'email_oauth' ? (
          <Button type="text" danger icon={<DeleteOutlined />} onClick={() => handleRevokeTokens(record.key)} />
        ) : (
          '-'
        ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 120,
      render: (_, record) => (
        <Button type="text" icon={<EditOutlined />} onClick={() => navigate('/datasources/' + record.key)} />
      ),
    },
  ]

  if (error) {
    return (
      <FullLayout>
        <div style={{ padding: 40 }}>
          <Typography.Text type="danger">Failed to load data sources. Please try again later.</Typography.Text>
        </div>
      </FullLayout>
    )
  }

  return (
    <FullLayout>
      <div style={{ padding: 24 }}>
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/datasources/add')}>
            Add Data Source
          </Button>
          <Table<DatasourceRow>
            columns={columns}
            dataSource={rows}
            loading={isLoading}
            rowKey="key"
            pagination={false}
            style={{ maxWidth: 1000 }}
          />
        </Space>
      </div>
    </FullLayout>
  )
}
