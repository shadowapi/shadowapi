import { useState } from 'react'
import { Table, Button, Space, Typography, Drawer, Divider } from 'antd'
import { EyeOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

import { useApiGet } from '@/api/hooks'
import { FullLayout } from '@/layouts/FullLayout'

interface FileRow {
  uuid: string
  name: string
  mime_type: string | null
  size: number | null
  created_at: string | null
  [key: string]: any
}

export function Files() {
  const [offset, setOffset] = useState(0)
  const [limit] = useState(20)
  const [selected, setSelected] = useState<FileRow | null>(null)

  const { data, isLoading } = useApiGet<FileRow[]>(`/file?offset=${offset}&limit=${limit}`)

  const rows = data ?? []

  const columns: ColumnsType<FileRow> = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    {
      title: 'MIME Type',
      dataIndex: 'mime_type',
      key: 'mime_type',
      render: (val: string | null) => val || 'N/A',
    },
    {
      title: 'Size (bytes)',
      dataIndex: 'size',
      key: 'size',
      render: (val: number | null) => val ?? 0,
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
      align: 'right',
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
        <Typography.Title level={4}>Files</Typography.Title>
        <Table<FileRow>
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
        title="Preview File"
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
