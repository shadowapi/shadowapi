import { useState } from 'react'
import { Table, Button, Space, Typography, Drawer, Divider } from 'antd'
import { EyeOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

import apiClient from '@/api/client'
import { FullLayout } from '@/layouts/FullLayout'
import useSWR from 'swr'

interface MessageRow {
  uuid: string
  sender: string
  subject: string | null
  created_at: string | null
  [key: string]: any
}

export function Messages() {
  const [offset, setOffset] = useState(0)
  const [limit] = useState(20)
  const [selected, setSelected] = useState<MessageRow | null>(null)

  const { data, isLoading } = useSWR<MessageRow[]>(
    ['messages', offset, limit],
    async () => {
      const resp = await apiClient.post('/message/query', {
        source: 'unified',
        offset,
        limit,
      })
      return resp.data?.messages ?? []
    },
  )

  const rows = data ?? []

  const columns: ColumnsType<MessageRow> = [
    { title: 'Sender', dataIndex: 'sender', key: 'sender' },
    {
      title: 'Subject',
      dataIndex: 'subject',
      key: 'subject',
      render: (val: string | null) => val || 'No subject',
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (val: string | null) => (val ? val.slice(0, 19) : ''),
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_, record) => (
        <Button type="text" icon={<EyeOutlined />} onClick={() => setSelected(record)}>
          Preview
        </Button>
      ),
    },
  ]

  return (
    <FullLayout>
      <div style={{ padding: 24 }}>
        <Typography.Title level={4}>Messages</Typography.Title>
        <Table<MessageRow>
          columns={columns}
          dataSource={rows}
          loading={isLoading}
          rowKey="uuid"
          pagination={false}
          style={{ maxWidth: 1200 }}
        />
        <Space style={{ marginTop: 16 }}>
          {offset > 0 && (
            <Button onClick={() => setOffset(Math.max(0, offset - limit))}>Previous</Button>
          )}
          <Button onClick={() => setOffset(offset + limit)}>Next</Button>
        </Space>
      </div>

      <Drawer
        title="Preview Message"
        open={!!selected}
        onClose={() => setSelected(null)}
        width={460}
      >
        {selected && (
          <>
            <Divider />
            <pre style={{ whiteSpace: 'pre-wrap', fontSize: '0.85rem' }}>
              {JSON.stringify(selected, null, 2)}
            </pre>
          </>
        )}
      </Drawer>
    </FullLayout>
  )
}
