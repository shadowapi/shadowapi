import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Table, Button, Space, Typography, Drawer, Divider, message } from 'antd'
import { CopyOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

import apiClient from '@/api/client'
import { FullLayout } from '@/layouts/FullLayout'
import { shortenUuid } from '@/shauth'
import useSWR from 'swr'

interface WorkerJob {
  uuid: string
  subject: string
  status: string
  scheduler_uuid: string
  job_uuid: string
  started_at: string | null
  finished_at: string | null
  [key: string]: any
}

export function WorkerJobs() {
  const navigate = useNavigate()
  const [offset, setOffset] = useState(0)
  const [limit] = useState(20)
  const [selected, setSelected] = useState<WorkerJob | null>(null)

  const copyToClipboard = (uuid: string, label: string) => {
    navigator.clipboard.writeText(uuid)
    message.success(`${label} UUID copied`)
  }

  const onCancel = (uuid: string) => {
    alert('TODO implement job cancel! job_uuid ' + uuid)
  }

  const { data, isLoading } = useSWR<WorkerJob[]>(
    ['workerjobs', offset, limit],
    async () => {
      const resp = await apiClient.get('/workerjobs', {
        params: { offset, limit },
      })
      return resp.data?.jobs ?? []
    },
  )

  const jobs = data ?? []

  const columns: ColumnsType<WorkerJob> = [
    { title: 'Subject', dataIndex: 'subject', key: 'subject' },
    { title: 'Status', dataIndex: 'status', key: 'status' },
    {
      title: 'Scheduler',
      key: 'scheduler_uuid',
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => navigate('/schedulers/' + record.scheduler_uuid)}>
            {shortenUuid(record.scheduler_uuid)}
          </Button>
          <Button
            type="text"
            size="small"
            icon={<CopyOutlined />}
            onClick={() => copyToClipboard(record.scheduler_uuid, 'Scheduler')}
          />
        </Space>
      ),
    },
    {
      title: 'Job',
      key: 'job_uuid',
      render: (_, record) => (
        <Space>
          <Button type="text" size="small" danger onClick={() => onCancel(record.job_uuid)}>
            x
          </Button>
          <Typography.Text>{shortenUuid(record.job_uuid)}</Typography.Text>
          <Button
            type="text"
            size="small"
            icon={<CopyOutlined />}
            onClick={() => copyToClipboard(record.job_uuid, 'Job')}
          />
        </Space>
      ),
    },
    {
      title: 'Started',
      dataIndex: 'started_at',
      key: 'started_at',
      render: (val: string | null) => (val ? val.slice(0, 19) : ''),
    },
    {
      title: 'Finished',
      dataIndex: 'finished_at',
      key: 'finished_at',
      render: (val: string | null) => (val ? val.slice(0, 19) : ''),
    },
    {
      title: 'Actions',
      key: 'actions',
      align: 'right',
      render: (_, record) => (
        <Button type="text" onClick={() => setSelected(record)}>
          Details
        </Button>
      ),
    },
  ]

  return (
    <FullLayout>
      <div style={{ padding: 24 }}>
        <Typography.Title level={4}>Worker Jobs</Typography.Title>
        <Table<WorkerJob>
          columns={columns}
          dataSource={jobs}
          loading={isLoading}
          rowKey="uuid"
          pagination={false}
        />
        <Space style={{ marginTop: 16 }}>
          {offset > 0 && (
            <Button onClick={() => setOffset(Math.max(0, offset - limit))}>Previous</Button>
          )}
          <Button onClick={() => setOffset(offset + limit)}>Next</Button>
        </Space>
      </div>

      <Drawer
        title="Preview Worker Job"
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
