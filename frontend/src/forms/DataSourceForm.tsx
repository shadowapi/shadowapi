import { ReactElement, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, Select, Space, Switch, Typography, message } from 'antd'
import { useSWRConfig } from 'swr'

import apiClient from '@/api/client'
import { useApiGet } from '@/api/hooks'
import type { components, paths } from '@/api/v1'

type DataSourceBase = {
  name: string
  type: string
  is_enabled: boolean
  user_uuid?: string
}

export type DataSourceKind = 'email' | 'email_oauth' | 'telegram' | 'whatsapp' | 'linkedin'

export type DataSourceFormData =
  | (DataSourceBase & { type: 'email' } & components['schemas']['datasource_email'])
  | (DataSourceBase & { type: 'email_oauth' } & components['schemas']['datasource_email_oauth'])
  | (DataSourceBase & { type: 'telegram' } & components['schemas']['datasource_telegram'])
  | (DataSourceBase & { type: 'whatsapp' } & components['schemas']['datasource_whatsapp'])
  | (DataSourceBase & { type: 'linkedin' } & components['schemas']['datasource_linkedin'])

const createEndpoints: Record<DataSourceKind, string> = {
  email: '/datasource/email',
  email_oauth: '/datasource/email_oauth',
  telegram: '/datasource/telegram',
  whatsapp: '/datasource/whatsapp',
  linkedin: '/datasource/linkedin',
}

const updateEndpoints: Record<DataSourceKind, string> = {
  email: '/datasource/email',
  email_oauth: '/datasource/email_oauth',
  telegram: '/datasource/telegram',
  whatsapp: '/datasource/whatsapp',
  linkedin: '/datasource/linkedin',
}

const deleteEndpoints: Record<DataSourceKind, string> = {
  email: '/datasource/email',
  email_oauth: '/datasource/email_oauth',
  telegram: '/datasource/telegram',
  whatsapp: '/datasource/whatsapp',
  linkedin: '/datasource/linkedin',
}

export function DataSourceForm({ datasourceUUID }: { datasourceUUID: string }): ReactElement {
  const navigate = useNavigate()
  const { mutate: globalMutate } = useSWRConfig()
  const [form] = Form.useForm<DataSourceFormData>()

  const watchedType = Form.useWatch('type', form)

  const { data: usersData } = useApiGet<components['schemas']['user'][]>('/user')
  const { data: oauth2ClientsData } = useApiGet<{ clients: components['schemas']['oauth2_client'][] }>('/oauth2/client')

  // In edit mode, fetch generic datasource record to get the type (dsKind)
  const { data: allDatasources } = useApiGet<components['schemas']['datasource'][]>(
    datasourceUUID !== 'add' ? '/datasource' : null
  )
  const genericData = allDatasources?.find((ds) => ds.uuid === datasourceUUID)
  const dsKind = datasourceUUID === 'add' ? undefined : (genericData?.type as DataSourceKind | undefined)
  const currentType = datasourceUUID === 'add' ? watchedType : dsKind || watchedType

  // Fetch detailed datasource data for editing
  const { data: detailData, isLoading: detailLoading } = useApiGet<DataSourceFormData>(
    datasourceUUID !== 'add' && dsKind ? `${updateEndpoints[dsKind]}/${datasourceUUID}` : null
  )

  // Reset the form when the record is loaded.
  useEffect(() => {
    if (detailData) {
      form.setFieldsValue({
        ...detailData,
        type: dsKind ?? (detailData as Partial<DataSourceFormData>).type,
      } as any)
    }
  }, [detailData, dsKind, form])

  const onSubmit = async (data: DataSourceFormData) => {
    try {
      const { type, ...payload } = data

      if (datasourceUUID === 'add') {
        const currentKind = type as DataSourceKind

        await apiClient.post(createEndpoints[currentKind], payload as any)
      } else {
        await apiClient.put(`${updateEndpoints[dsKind!]}/${datasourceUUID}`, payload as any)
      }
      message.success(datasourceUUID === 'add' ? 'Data source created' : 'Data source updated')
      globalMutate('/datasource')
      navigate('/datasources')
    } catch (err: any) {
      const detail = err?.response?.data?.detail || err.message
      message.error(detail)
    }
  }

  const onDelete = async () => {
    try {
      await apiClient.delete(`${deleteEndpoints[dsKind!]}/${datasourceUUID}`)
      message.success('Data source deleted')
      globalMutate('/datasource')
      navigate('/datasources')
    } catch (err: any) {
      const detail = err?.response?.data?.detail || err.message
      message.error(detail)
    }
  }

  if (datasourceUUID !== 'add' && (!genericData || detailLoading)) {
    return <></>
  }

  return (
    <div style={{ display: 'flex', justifyContent: 'center', width: '100%', minHeight: '100vh' }}>
      <Form form={form} onFinish={onSubmit} layout="horizontal" labelCol={{ span: 6 }} wrapperCol={{ span: 14 }} style={{ width: 400 }}>
        <Typography.Title level={4}>Data Source</Typography.Title>

        <Form.Item name="name" label="Name" rules={[{ required: true, message: 'Name is required' }]}>
          <Input />
        </Form.Item>

        <Form.Item name="user_uuid" label="User" rules={[{ required: true, message: 'User is required' }]}>
          <Select>
            {usersData?.map((user) => (
              <Select.Option key={user.uuid} value={user.uuid}>
                {user.email} {user.first_name} {user.last_name}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item name="type" label="Type" rules={[{ required: true, message: 'Type is required' }]}>
          <Select disabled={datasourceUUID !== 'add'}>
            <Select.Option value="email">Email IMAP</Select.Option>
            <Select.Option value="email_oauth">Email OAuth</Select.Option>
            <Select.Option value="telegram">Telegram</Select.Option>
            <Select.Option value="whatsapp">WhatsApp</Select.Option>
            <Select.Option value="linkedin">LinkedIn</Select.Option>
          </Select>
        </Form.Item>

        <Form.Item name="is_enabled" label="Enabled" valuePropName="checked">
          <Switch />
        </Form.Item>

        {currentType === 'email' && (
          <>
            <Form.Item name="email" label="Email">
              <Input />
            </Form.Item>
            <Form.Item name="provider" label="Provider">
              <Input />
            </Form.Item>
            <Form.Item name="imap_server" label="IMAP Server">
              <Input />
            </Form.Item>
            <Form.Item name="smtp_server" label="SMTP Server">
              <Input />
            </Form.Item>
            <Form.Item name="smtp_tls" label="SMTP TLS" valuePropName="checked">
              <Switch />
            </Form.Item>
            <Form.Item name="password" label="Password">
              <Input.Password />
            </Form.Item>
          </>
        )}

        {currentType === 'email_oauth' && (
          <>
            <Form.Item name="email" label="Email">
              <Input />
            </Form.Item>
            <Form.Item name="provider" label="Provider">
              <Input />
            </Form.Item>
            <Form.Item name="oauth2_client_uuid" label="OAuth2 Client">
              <Select>
                {oauth2ClientsData?.clients?.map((c) => (
                  <Select.Option key={c.uuid} value={c.uuid}>
                    {c.name} ({c.client_id})
                  </Select.Option>
                ))}
              </Select>
            </Form.Item>
          </>
        )}

        {currentType === 'telegram' && (
          <>
            <Form.Item name="phone_number" label="Phone Number">
              <Input />
            </Form.Item>
            <Form.Item name="provider" label="Provider">
              <Input />
            </Form.Item>
            <Form.Item name="api_id" label="API ID">
              <Input />
            </Form.Item>
            <Form.Item name="api_hash" label="API Hash">
              <Input />
            </Form.Item>
            <Form.Item name="password" label="Password">
              <Input.Password />
            </Form.Item>
          </>
        )}

        {currentType === 'whatsapp' && (
          <>
            <Form.Item name="phone_number" label="Phone Number">
              <Input />
            </Form.Item>
            <Form.Item name="provider" label="Provider">
              <Input />
            </Form.Item>
            <Form.Item name="device_name" label="Device Name">
              <Input />
            </Form.Item>
          </>
        )}

        {currentType === 'linkedin' && (
          <>
            <Form.Item name="username" label="Username">
              <Input />
            </Form.Item>
            <Form.Item name="password" label="Password">
              <Input.Password />
            </Form.Item>
            <Form.Item name="provider" label="Provider">
              <Input />
            </Form.Item>
          </>
        )}

        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">
              {datasourceUUID === 'add' ? 'Create' : 'Update'}
            </Button>
            {datasourceUUID !== 'add' && (
              <Button danger onClick={onDelete}>
                Delete
              </Button>
            )}
          </Space>
        </Form.Item>
      </Form>
    </div>
  )
}
