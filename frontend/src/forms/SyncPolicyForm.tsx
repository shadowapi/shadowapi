import { ReactElement, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, Select, Space, Switch, Typography, message } from 'antd'
import { useSWRConfig } from 'swr'

import apiClient from '@/api/client'
import { useApiGet } from '@/api/hooks'
import type { components } from '@/api/v1'

type SyncPolicyFormData = {
  name?: string
  pipeline_uuid: string
  blocklist?: string[]
  exclude_list?: string[]
  sync_all: boolean
  settings?: string
}

export function SyncPolicyForm({ policyUUID }: { policyUUID: string }): ReactElement {
  const navigate = useNavigate()
  const { mutate: globalMutate } = useSWRConfig()
  const [form] = Form.useForm<SyncPolicyFormData>()

  const isAdd = policyUUID === 'add'

  const { data: pipelinesData, isLoading: pipelinesLoading } = useApiGet<{
    pipelines: components['schemas']['pipeline'][]
  }>('/pipeline')

  const { data: policyData, isLoading: policyLoading } = useApiGet<any>(
    isAdd ? null : `/syncpolicy/${policyUUID}`
  )

  useEffect(() => {
    if (policyData && !isAdd) {
      const data = { ...policyData }
      if (data.settings) {
        data.settings = JSON.stringify(data.settings, null, 2)
      }
      form.setFieldsValue(data)
    }
  }, [policyData, isAdd, form])

  const onSubmit = async (values: SyncPolicyFormData) => {
    try {
      const payload = {
        ...values,
        settings: values.settings ? JSON.parse(values.settings) : null,
      }
      if (isAdd) {
        await apiClient.post('/syncpolicy', payload)
      } else {
        await apiClient.put(`/syncpolicy/${policyUUID}`, payload)
      }
      message.success(isAdd ? 'Sync policy created' : 'Sync policy updated')
      globalMutate('/syncpolicy')
      navigate('/syncpolicies')
    } catch (err: any) {
      const detail = err?.response?.data?.detail || err.message
      message.error(detail)
    }
  }

  const onDelete = async () => {
    try {
      await apiClient.delete(`/syncpolicy/${policyUUID}`)
      message.success('Sync policy deleted')
      globalMutate('/syncpolicy')
      navigate('/syncpolicies')
    } catch (err: any) {
      const detail = err?.response?.data?.detail || err.message
      message.error(detail)
    }
  }

  if (policyLoading && pipelinesLoading && !isAdd) return <></>

  return (
    <div style={{ display: 'flex', justifyContent: 'center', minHeight: '100vh' }}>
      <Form
        form={form}
        onFinish={onSubmit}
        layout="horizontal" labelCol={{ span: 6 }} wrapperCol={{ span: 14 }}
        style={{ width: 400 }}
        initialValues={{ settings: '{}' }}
      >
        <Typography.Title level={4}>{isAdd ? 'Add Sync Policy' : 'Edit Sync Policy'}</Typography.Title>

        <Form.Item name="name" label="Name" rules={[{ required: true, message: 'Name is required' }]}>
          <Input />
        </Form.Item>

        <Form.Item
          name="pipeline_uuid"
          label="Pipeline"
          rules={[{ required: true, message: 'Pipeline is required' }]}
        >
          <Select>
            {pipelinesData?.pipelines?.map((pipeline) => (
              <Select.Option key={pipeline.uuid} value={pipeline.uuid}>
                {pipeline.name} {pipeline.type}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item name="sync_all" label="Sync All" valuePropName="checked">
          <Switch />
        </Form.Item>

        <Form.Item
          name="blocklist"
          label="Blocklist (comma separated)"
          getValueFromEvent={(e) => e.target.value.split(',').map((s: string) => s.trim())}
          getValueProps={(val) => ({ value: Array.isArray(val) ? val.join(', ') : val })}
        >
          <Input />
        </Form.Item>

        <Form.Item
          name="exclude_list"
          label="Exclude List (comma separated)"
          getValueFromEvent={(e) => e.target.value.split(',').map((s: string) => s.trim())}
          getValueProps={(val) => ({ value: Array.isArray(val) ? val.join(', ') : val })}
        >
          <Input />
        </Form.Item>

        <Form.Item name="settings" label="Settings (JSON)">
          <Input.TextArea rows={4} />
        </Form.Item>

        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">
              {isAdd ? 'Create' : 'Update'}
            </Button>
            {!isAdd && (
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
