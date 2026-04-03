import { ReactElement, useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Button, Form, Input, Select, Space, Typography, message } from 'antd'
import { useSWRConfig } from 'swr'

import apiClient from '@/api/client'
import { useApiGet } from '@/api/hooks'
import type { components } from '@/api/v1'

export function OAuth2CredentialForm({ clientID }: { clientID: string }): ReactElement {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { mutate: globalMutate } = useSWRConfig()
  const [form] = Form.useForm<components['schemas']['oauth2_client']>()

  const { data } = useApiGet<components['schemas']['oauth2_client']>(
    clientID !== 'add' ? `/oauth2/client/${clientID}` : null
  )

  useEffect(() => {
    if (data) {
      form.setFieldsValue(data)
    }
  }, [data, form])

  const onSubmit = async (values: components['schemas']['oauth2_client']) => {
    try {
      if (clientID === 'add') {
        await apiClient.post('/oauth2/client', {
          provider: values.provider,
          client_id: values.client_id,
          name: values.name,
          secret: values.secret,
        })
      } else {
        await apiClient.put(`/oauth2/client/${clientID}`, {
          provider: values.provider,
          name: values.name,
          secret: values.secret,
          client_id: values.client_id,
        })
      }
      message.success(clientID === 'add' ? 'OAuth2 credential created' : 'OAuth2 credential updated')
      globalMutate('/oauth2/client')
      globalMutate('/oauth2/credentials')
      const datasourceUUID = searchParams.get('datasource_uuid')
      if (datasourceUUID) {
        navigate(`/datasource/${datasourceUUID}/auth?client_id=${values.uuid}`)
        return
      }
      navigate('/oauth2/credentials')
    } catch (err: any) {
      const detail = err?.response?.data?.detail || err.message
      message.error(detail)
    }
  }

  const onDelete = async () => {
    try {
      await apiClient.delete(`/oauth2/client/${clientID}`)
      message.success('OAuth2 credential deleted')
      globalMutate('/oauth2/client')
      globalMutate('/oauth2/credentials')
      navigate('/oauth2/credentials')
    } catch (err: any) {
      const detail = err?.response?.data?.detail || err.message
      message.error(detail)
    }
  }

  return (
    <div style={{ display: 'flex', justifyContent: 'center', width: '100%', minHeight: '100vh' }}>
      <Form form={form} onFinish={onSubmit} layout="horizontal" labelCol={{ span: 6 }} wrapperCol={{ span: 14 }} style={{ width: 400 }}>
        <Typography.Title level={4}>OAuth2 Credential</Typography.Title>

        <Form.Item name="name" label="Name" rules={[{ required: true, message: 'Name is required' }]}>
          <Input />
        </Form.Item>

        <Form.Item name="provider" label="Provider" rules={[{ required: true, message: 'Provider is required' }]}>
          <Select>
            <Select.Option value="GMAIL">Gmail</Select.Option>
          </Select>
        </Form.Item>

        <Form.Item name="client_id" label="Client ID" rules={[{ required: true, message: 'Client ID is required' }]}>
          <Input />
        </Form.Item>

        <Form.Item name="secret" label="Client Secret" rules={[{ required: true, message: 'Client Secret is required' }]}>
          <Input.Password />
        </Form.Item>

        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">
              {clientID === 'add' ? 'Create' : 'Update'}
            </Button>
            {clientID !== 'add' && (
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
