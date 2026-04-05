import { useState } from 'react'
import { Table, Button, Space, Typography, Drawer } from 'antd'
import { EyeOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'

import apiClient from '@/api/client'
import { FullLayout } from '@/layouts/FullLayout'
import JsonPrintNoStore from '@/components/JsonPrintNoStore'
import useSWR from 'swr'

interface MessageRow {
  uuid: string
  sender: string
  subject: string | null
  body: string | null
  recipients?: string[] | null
  created_at: string | null
  [key: string]: any
}

function parseSender(val: string) {
  const match = val.match(/^(.*?)\s*<(.+)>$/)
  if (match) return { name: match[1] || match[2], email: match[2] }
  return { name: val, email: val }
}

function formatDate(val: string | null) {
  if (!val) return ''
  const d = new Date(val)
  const now = new Date()
  const isToday = d.toDateString() === now.toDateString()
  const isThisYear = d.getFullYear() === now.getFullYear()
  if (isToday) return d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })
  if (isThisYear) return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

function formatFullDate(val: string | null) {
  if (!val) return ''
  const d = new Date(val)
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' }) +
    ' ' + d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })
}

function isHtml(text: string) {
  return /<[a-z][\s\S]*>/i.test(text)
}

/** Decode HTML entities like &#39; &amp; etc and strip tags for plain-text preview */
function stripToPlain(text: string): string {
  const el = document.createElement('div')
  el.innerHTML = text
  return el.textContent || el.innerText || ''
}

/** Collapse runs of 3+ newlines into 2, trim, decode entities */
function cleanBodyForDisplay(body: string): string {
  let text = stripToPlain(body)
  // collapse 3+ consecutive newlines (with optional whitespace) into double newline
  text = text.replace(/(\s*\n\s*){3,}/g, '\n\n')
  return text.trim()
}

/** Get first N chars of body as plain text for table preview */
function bodyPreview(body: string | null, maxLen = 140): string {
  if (!body) return ''
  const plain = stripToPlain(body).replace(/\s+/g, ' ').trim()
  if (plain.length <= maxLen) return plain
  return plain.slice(0, maxLen) + '...'
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
    {
      title: 'Sender',
      dataIndex: 'sender',
      key: 'sender',
      width: 200,
      render: (val: string) => {
        const { name, email } = parseSender(val)
        return (
          <div style={{ overflow: 'hidden' }}>
            <div style={{ fontWeight: 500, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
              {name}
            </div>
            {name !== email && (
              <div style={{ fontSize: 12, color: '#8c8c8c', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                {email}
              </div>
            )}
          </div>
        )
      },
    },
    {
      title: 'Subject',
      dataIndex: 'subject',
      key: 'subject',
      render: (_: string | null, record: MessageRow) => {
        const subject = record.subject || 'No subject'
        const preview = bodyPreview(record.body)
        return (
          <div
            style={{ overflow: 'hidden', cursor: 'pointer' }}
            onClick={() => setSelected(record)}
          >
            <div style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
              {record.subject ? subject : <span style={{ color: '#8c8c8c' }}>{subject}</span>}
            </div>
            {preview && (
              <div style={{ fontSize: 12, color: '#8c8c8c', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                {preview}
              </div>
            )}
          </div>
        )
      },
    },
    {
      title: 'Date',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: (val: string | null) => (
        <span style={{ whiteSpace: 'nowrap' }}>{formatDate(val)}</span>
      ),
    },
    {
      title: '',
      key: 'actions',
      width: 50,
      render: (_, record) => (
        <Button type="text" size="small" icon={<EyeOutlined />} onClick={() => setSelected(record)} />
      ),
    },
  ]

  const selectedSender = selected ? parseSender(selected.sender) : null

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
          tableLayout="fixed"
        />
        <Space style={{ marginTop: 16 }}>
          {offset > 0 && (
            <Button onClick={() => setOffset(Math.max(0, offset - limit))}>Previous</Button>
          )}
          <Button onClick={() => setOffset(offset + limit)}>Next</Button>
        </Space>
      </div>

      <Drawer
        title={null}
        open={!!selected}
        onClose={() => setSelected(null)}
        width={560}
        styles={{ body: { padding: 0 } }}
      >
        {selected && selectedSender && (
          <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
            {/* Email header */}
            <div style={{ padding: '16px 20px', borderBottom: '1px solid #f0f0f0' }}>
              <Typography.Title level={5} style={{ margin: '0 0 12px' }}>
                {selected.subject || 'No subject'}
              </Typography.Title>
              <div style={{ display: 'flex', alignItems: 'flex-start', gap: 10 }}>
                <div style={{
                  width: 36, height: 36, borderRadius: '50%',
                  background: '#595959', color: '#fff',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  fontSize: 14, flexShrink: 0,
                }}>
                  {selectedSender.name.charAt(0).toUpperCase()}
                </div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontWeight: 500 }}>{selectedSender.name}</div>
                  <div style={{ fontSize: 12, color: '#8c8c8c' }}>{selectedSender.email}</div>
                  {selected.recipients && selected.recipients.length > 0 && (
                    <div style={{ fontSize: 12, color: '#8c8c8c', marginTop: 2 }}>
                      To: {selected.recipients.join(', ')}
                    </div>
                  )}
                </div>
                <div style={{ fontSize: 12, color: '#8c8c8c', whiteSpace: 'nowrap', flexShrink: 0 }}>
                  {formatFullDate(selected.created_at)}
                </div>
              </div>
            </div>

            {/* Email body */}
            <div style={{ flex: 1, padding: '16px 20px', overflow: 'auto' }}>
              {selected.body ? (
                <div style={{ fontSize: 14, lineHeight: 1.6, whiteSpace: 'pre-wrap', fontFamily: 'inherit' }}>
                  {cleanBodyForDisplay(selected.body)}
                </div>
              ) : (
                <Typography.Text type="secondary">No message body</Typography.Text>
              )}
            </div>

            {/* Debug section */}
            <div style={{ padding: '8px 20px 16px', borderTop: '1px solid #f0f0f0' }}>
              <JsonPrintNoStore
                title="Debug: Raw Message Data"
                data={selected}
                off={true}
              />
            </div>
          </div>
        )}
      </Drawer>
    </FullLayout>
  )
}
